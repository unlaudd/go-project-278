package handler

import (
	"fmt"
	"net/http"

	"url-shortener/internal/repository"
	"url-shortener/internal/utils"

	"github.com/gin-gonic/gin"
)

type LinkVisitHandler struct {
	repo repository.LinkVisitRepository
}

func NewLinkVisitHandler(repo repository.LinkVisitRepository) *LinkVisitHandler {
	return &LinkVisitHandler{repo: repo}
}

// GET /api/link_visits — список посещений с пагинацией
func (h *LinkVisitHandler) List(c *gin.Context) {
	start, end, err := utils.ParseRange(c.Query("range"))
	if err != nil {
		start, end = 0, 9 // дефолт
	}

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

	actualEnd := start + len(visits) - 1
	if len(visits) == 0 {
		actualEnd = start - 1
	}
	c.Header("Content-Range", fmt.Sprintf("link_visits %d-%d/%d", start, actualEnd, total))

	c.JSON(http.StatusOK, visits)
}
