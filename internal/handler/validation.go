// Package handler provides HTTP handlers and validation helpers for the API.
package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/lib/pq"
)

// handleBindError processes JSON binding and validation errors from Gin.
// Returns 422 Unprocessable Entity with field-specific messages for validation failures,
// or 400 Bad Request for malformed JSON that cannot be parsed.
func handleBindError(c *gin.Context, err error) {
	var validationErrs validator.ValidationErrors
	if errors.As(err, &validationErrs) {
		// Format validation errors as {"errors": {"field": "message"}} per API spec.
		errs := make(map[string]string)
		for _, e := range validationErrs {
			errs[e.Field()] = e.Error()
		}
		c.JSON(http.StatusUnprocessableEntity, gin.H{"errors": errs})
		return
	}
	// Fallback for JSON syntax errors or other binding issues.
	c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
}

// handleDBError maps PostgreSQL error codes to appropriate HTTP responses.
// Specifically handles unique constraint violations (code 23505) by returning
// a 422 response with a user-friendly message for the "short_name" field.
// All other database errors return a generic 500 Internal Server Error.
func handleDBError(c *gin.Context, err error) {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "23505" { // unique_violation
		c.JSON(http.StatusUnprocessableEntity, gin.H{"errors": gin.H{"short_name": "short name already in use"}})
		return
	}
	// Log the error in production; return generic message to avoid leaking internals.
	c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
}
