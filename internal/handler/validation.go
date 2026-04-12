package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/lib/pq"
)

// handleBindError обрабатывает ошибки парсинга JSON и валидации
func handleBindError(c *gin.Context, err error) {
	var validationErrs validator.ValidationErrors
	if errors.As(err, &validationErrs) {
		// Формат ошибки совпадает с заданием
		errs := make(map[string]string)
		for _, e := range validationErrs {
			errs[e.Field()] = e.Error()
		}
		c.JSON(http.StatusUnprocessableEntity, gin.H{"errors": errs})
		return
	}
	// Некорректный JSON
	c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
}

// handleDBError преобразует ошибки БД в HTTP-ответы
func handleDBError(c *gin.Context, err error) {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "23505" { // unique_violation
		c.JSON(http.StatusUnprocessableEntity, gin.H{"errors": gin.H{"short_name": "short name already in use"}})
		return
	}
	// Другие ошибки БД
	c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
}
