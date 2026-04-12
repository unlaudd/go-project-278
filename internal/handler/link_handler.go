package handler

import (
	"net/http"
	"strconv"

	"url-shortener/internal/repository"

	"github.com/gin-gonic/gin"
)

type LinkHandler struct {
	repo    repository.LinkRepository
	baseURL string
}

func NewLinkHandler(repo repository.LinkRepository, baseURL string) *LinkHandler {
	return &LinkHandler{repo: repo, baseURL: baseURL}
}

func (h *LinkHandler) Create(c *gin.Context) {
	var req struct {
		OriginalURL string `json:"original_url" binding:"required"`
		ShortName   string `json:"short_name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	shortName := req.ShortName
	if shortName == "" {
		shortName = generateShortName(8)
	}

	link := &repository.Link{
		OriginalURL: req.OriginalURL,
		ShortName:   shortName,
	}

	if err := h.repo.Create(c.Request.Context(), link, h.baseURL); err != nil {
		if err.Error() == "short_name already exists" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create link"})
		return
	}

	c.JSON(http.StatusCreated, link)
}

func (h *LinkHandler) GetByID(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	link, err := h.repo.GetByID(c.Request.Context(), id, h.baseURL)
	if err != nil {
		if err.Error() == "link not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "link not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get link"})
		return
	}
	c.JSON(http.StatusOK, link)
}

func (h *LinkHandler) List(c *gin.Context) {
	limit, offset := int32(100), int32(0)
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = int32(v)
		}
	}
	if o := c.Query("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = int32(v)
		}
	}

	links, err := h.repo.List(c.Request.Context(), limit, offset, h.baseURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list links"})
		return
	}
	c.JSON(http.StatusOK, links)
}

func (h *LinkHandler) Update(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req struct {
		OriginalURL *string `json:"original_url"`
		ShortName   *string `json:"short_name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	link, err := h.repo.Update(c.Request.Context(), id, req.OriginalURL, req.ShortName, h.baseURL)
	if err != nil {
		if err.Error() == "link not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "link not found"})
			return
		}
		if err.Error() == "short_name already exists" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update link"})
		return
	}
	c.JSON(http.StatusOK, link)
}

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

func parseID(s string) (int32, error) {
	v, err := strconv.ParseInt(s, 10, 32)
	return int32(v), err
}

func generateShortName(length int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = chars[i%len(chars)]
	}
	return string(result)
}
