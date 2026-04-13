// Package handler provides HTTP handlers for the link shortener API.
package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"url-shortener/internal/repository"
	"url-shortener/internal/utils"
)

// LinkVisitHandler handles HTTP requests for link visit analytics.
type LinkVisitHandler struct {
	repo repository.LinkVisitRepository
}

// NewLinkVisitHandler creates a new LinkVisitHandler with the given repository.
func NewLinkVisitHandler(repo repository.LinkVisitRepository) *LinkVisitHandler {
	return &LinkVisitHandler{repo: repo}
}

// List handles GET /api/link_visits — returns a paginated list of link visits.
// Supports RFC 7233-style range queries via the "range" query parameter: ?range=[start,end].
// Sets the Content-Range response header in the format: "link_visits start-end/total".
func (h *LinkVisitHandler) List(c *gin.Context) {
	// Parse the range parameter; fall back to default [0,9] if missing or malformed.
	start, end, err := utils.ParseRange(c.Query("range"))
	if err != nil {
		start, end = 0, 9
	}

	// Range is inclusive: [0,10] means 11 items.
	limit := end - start + 1
	offset := start

	visits, err := h.repo.List(c.Request.Context(), int32(limit), int32(offset))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list visits"})
		return
	}

	total, err := h.repo.Count(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to count visits"})
		return
	}

	// Format Content-Range header per RFC 7233.
	// For empty results, use "start-(start-1)/total" to indicate no content in range.
	actualEnd := start + len(visits) - 1
	if len(visits) == 0 {
		actualEnd = start - 1
	}
	c.Header("Content-Range", fmt.Sprintf("link_visits %d-%d/%d", start, actualEnd, total))

	c.JSON(http.StatusOK, visits)
}
