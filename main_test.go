package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// setupTest configures the test environment by switching Gin to test mode,
// which disables verbose request logging during test execution.
func setupTest() {
	gin.SetMode(gin.TestMode)
}

// TestPingEndpoint verifies that the /ping health check endpoint
// returns the expected status code, response body, and content type.
func TestPingEndpoint(t *testing.T) {
	setupTest()

	router := NewRouter()

	w := httptest.NewRecorder()
	// Error is intentionally ignored: the request parameters are static and known to be valid.
	req, _ := http.NewRequest(http.MethodGet, "/ping", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "expected HTTP status code")
	assert.Equal(t, "pong", w.Body.String(), "expected response body")
	assert.Equal(t, "text/plain; charset=utf-8", w.Header().Get("Content-Type"), "expected Content-Type header")
}
