package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"url-shortener/internal/handler"
	"url-shortener/internal/repository"
)

// Моковый репозиторий
type mockRepo struct {
	mock.Mock
}

func (m *mockRepo) Create(ctx context.Context, link *repository.Link, baseURL string) error {
	args := m.Called(ctx, link, baseURL)
	return args.Error(0)
}
// ... остальные методы мока ...

func TestCreateLink(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := new(mockRepo)
	h := handler.NewLinkHandler(repo, "http://test.local")

	repo.On("Create", mock.Anything, mock.Anything, "http://test.local").Return(nil)

	w := httptest.NewRecorder()
	body := bytes.NewBufferString(`{"original_url":"https://example.com"}`)
	req, _ := http.NewRequest("POST", "/api/links", body)
	req.Header.Set("Content-Type", "application/json")

	h.Create(gin.CreateTestContext(w).Request)

	assert.Equal(t, http.StatusCreated, w.Code)
	repo.AssertExpectations(t)
}
