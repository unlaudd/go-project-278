// Package handler provides HTTP handlers for the link shortener API.
package handler

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"url-shortener/internal/repository"
	"url-shortener/internal/utils"
)

// LinkHandler handles HTTP requests for link resources.
type LinkHandler struct {
	repo    repository.LinkRepository
	baseURL string
}

// NewLinkHandler creates a new LinkHandler with the given repository and base URL
// for constructing short link addresses.
func NewLinkHandler(repo repository.LinkRepository, baseURL string) *LinkHandler {
	return &LinkHandler{repo: repo, baseURL: baseURL}
}

// Create handles POST /api/links — creates a new shortened link.
// Validates input, generates a short name if not provided, and returns 201 Created.
func (h *LinkHandler) Create(c *gin.Context) {
	var req struct {
		OriginalURL string  `json:"original_url" binding:"required,url"`
		ShortName   *string `json:"short_name" binding:"omitempty,min=3,max=32"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		handleBindError(c, err)
		return
	}

	// Generate a random short name if the client did not provide one.
	shortName := req.ShortName
	if shortName == nil || *shortName == "" {
		gen := generateShortName()
		shortName = &gen
	}

	link := &repository.Link{
		OriginalURL: req.OriginalURL,
		ShortName:   *shortName,
	}

	if err := h.repo.Create(c.Request.Context(), link, h.baseURL); err != nil {
		handleDBError(c, err)
		return
	}

	c.JSON(http.StatusCreated, link)
}

// GetByID handles GET /api/links/:id — retrieves a link by its numeric ID.
func (h *LinkHandler) GetByID(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	link, err := h.repo.GetByID(c.Request.Context(), id, h.baseURL)
	if err != nil {
		// Simple string comparison for "not found" — consider using errors.Is in production.
		if err.Error() == "link not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "link not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get link"})
		return
	}
	c.JSON(http.StatusOK, link)
}

// List handles GET /api/links — returns a paginated list of links.
// Supports RFC 7233-style Range header via query parameter: ?range=[start,end].
func (h *LinkHandler) List(c *gin.Context) {
	start, end, err := utils.ParseRange(c.Query("range"))
	if err != nil {
		// Fall back to default range [0,9] if parameter is missing or malformed.
		start, end = 0, 9
	}

	// Range is inclusive: [0,10] means 11 items.
	limit := end - start + 1
	offset := start

	links, err := h.repo.List(c.Request.Context(), int32(limit), int32(offset), h.baseURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list links"})
		return
	}

	total, err := h.repo.Count(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to count links"})
		return
	}

	// Format Content-Range header per RFC 7233: "links start-end/total".
	// For empty results, use "start-(start-1)/total" to indicate no content in range.
	actualEnd := start + len(links) - 1
	if len(links) == 0 {
		actualEnd = start - 1
	}
	c.Header("Content-Range", fmt.Sprintf("links %d-%d/%d", start, actualEnd, total))

	c.JSON(http.StatusOK, links)
}

// Update handles PUT /api/links/:id — updates an existing link.
// Fields are optional: only provided fields are updated.
func (h *LinkHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid link id"})
		return
	}

	var req struct {
		OriginalURL *string `json:"original_url" binding:"omitempty,url"`
		ShortName   *string `json:"short_name" binding:"omitempty,min=3,max=32"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		handleBindError(c, err)
		return
	}

	link, err := h.repo.Update(c.Request.Context(), int32(id), req.OriginalURL, req.ShortName, h.baseURL)
	if err != nil {
		if err.Error() == "link not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "link not found"})
			return
		}
		handleDBError(c, err)
		return
	}

	c.JSON(http.StatusOK, link)
}

// Delete handles DELETE /api/links/:id — removes a link by ID.
// Returns 204 No Content on success.
func (h *LinkHandler) Delete(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.repo.Delete(c.Request.Context(), id); err != nil {
		if err.Error() == "link not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "link not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete link"})
		return
	}
	c.Status(http.StatusNoContent)
}

// parseID converts a string ID from the URL path to int32.
func parseID(s string) (int32, error) {
	v, err := strconv.ParseInt(s, 10, 32)
	return int32(v), err
}

// generateShortName creates a random 8-character alphanumeric string.
// Note: Uses math/rand, which is not cryptographically secure — acceptable for short names.
func generateShortName() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 8)
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := range b {
		b[i] = charset[rng.Intn(len(charset))]
	}
	return string(b)
}
