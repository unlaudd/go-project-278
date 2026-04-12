package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"url-shortener/internal/repository"
)

type LinkHandler struct {
	repo    repository.LinkRepository
	baseURL string
}

func NewLinkHandler(repo repository.LinkRepository, baseURL string) *LinkHandler {
	return &LinkHandler{repo: repo, baseURL: baseURL}
}

// POST /api/links
func (h *LinkHandler) Create(c *gin.Context) {
	var req struct {
		OriginalURL string `json:"original_url" binding:"required"`
		ShortName   string `json:"short_name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// Если short_name не указан — генерируем
	shortName := req.ShortName
	if shortName == "" {
		shortName = generateShortName(6) // простая генерация
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

// GET /api/links/:id
func (h *LinkHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
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

// GET /api/links
func (h *LinkHandler) List(c *gin.Context) {
	limit := 100
	offset := 0
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}
	if o := c.Query("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}

	links, err := h.repo.List(c.Request.Context(), limit, offset, h.baseURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list links"})
		return
	}

	c.JSON(http.StatusOK, links)
}

// PUT /api/links/:id
func (h *LinkHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
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

// DELETE /api/links/:id
func (h *LinkHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
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

// Простая генерация короткого имени (base62)
func generateShortName(length int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = chars[i%len(chars)] // упрощённо; в продакшене используйте crypto/rand
	}
	return string(result)
}
