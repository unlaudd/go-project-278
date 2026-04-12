package handler

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"url-shortener/internal/repository"
	"url-shortener/internal/utils"

	"github.com/gin-gonic/gin"
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
		OriginalURL string  `json:"original_url" binding:"required,url"`
		ShortName   *string `json:"short_name" binding:"omitempty,min=3,max=32"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		handleBindError(c, err)
		return
	}

	// Генерируем short_name, если не передан или пустой
	shortName := req.ShortName
	if shortName == nil || *shortName == "" {
		gen := generateShortName()
		shortName = &gen
	}

	// Создаём структуру ссылки
	link := &repository.Link{
		OriginalURL: req.OriginalURL,
		ShortName:   *shortName,
	}

	// repo.Create возвращает только error
	if err := h.repo.Create(c.Request.Context(), link, h.baseURL); err != nil {
		handleDBError(c, err)
		return
	}

	// Возвращаем 201 + заполненную ссылку
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
	// Парсим параметр range=[start,end]
	start, end, err := utils.ParseRange(c.Query("range"))
	if err != nil {
		// Если параметр не указан или некорректен — используем дефолт [0,9] (10 записей)
		start, end = 0, 9
	}

	limit := end - start + 1 // inclusive: [0,10] → 11 записей
	offset := start

	links, err := h.repo.List(c.Request.Context(), int32(limit), int32(offset), h.baseURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list links"})
		return
	}

	// Получаем общее количество для заголовка Content-Range
	total, err := h.repo.Count(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to count links"})
		return
	}

	// Формируем заголовок Content-Range: links start-end/total
	// Пример: Content-Range: links 0-10/42
	actualEnd := start + len(links) - 1
	if len(links) == 0 {
		actualEnd = start - 1 // пустой результат: 0-(-1)/0
	}
	c.Header("Content-Range", fmt.Sprintf("links %d-%d/%d", start, actualEnd, total))

	c.JSON(http.StatusOK, links)
}

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
		handleDBError(c, err)
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

func generateShortName() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 8)
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := range b {
		b[i] = charset[rng.Intn(len(charset))]
	}
	return string(b)
}
