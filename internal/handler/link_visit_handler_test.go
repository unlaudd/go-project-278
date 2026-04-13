//go:build unit
// +build unit

package handler_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"url-shortener/internal/handler"
	"url-shortener/internal/repository"
)

// createTestContext initializes a Gin context and HTTP request for unit testing handlers.
// It sets the request method, path, optional JSON body, and Content-Type header.
// This helper is duplicated here to avoid cross-file dependencies in test compilation.
func createTestContext(w *httptest.ResponseRecorder, method, path string, body []byte) (*gin.Context, *http.Request) {
	gin.SetMode(gin.TestMode)
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	w = httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	return c, req
}

// mockVisitRepo is a test double for repository.LinkVisitRepository.
// It records method calls and returns stubbed values using stretchr/testify/mock.
type mockVisitRepo struct {
	mock.Mock
}

func (m *mockVisitRepo) Create(ctx context.Context, visit *repository.LinkVisit) error {
	args := m.Called(ctx, visit)
	return args.Error(0)
}

func (m *mockVisitRepo) List(ctx context.Context, limit, offset int32) ([]*repository.LinkVisit, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*repository.LinkVisit), args.Error(1)
}

func (m *mockVisitRepo) Count(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

// TestListVisits_DefaultPagination verifies that omitting the range parameter
// defaults to fetching the first 10 visits [0,9] and sets the Content-Range header
// in the format "link_visits start-end/total".
func TestListVisits_DefaultPagination(t *testing.T) {
	repo := new(mockVisitRepo)
	h := handler.NewLinkVisitHandler(repo)

	expected := make([]*repository.LinkVisit, 10)
	repo.On("List", mock.Anything, int32(10), int32(0)).Return(expected, nil)
	repo.On("Count", mock.Anything).Return(int64(50), nil)

	w := httptest.NewRecorder()
	c, _ := createTestContext(w, "GET", "/api/link_visits", nil)
	h.List(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "link_visits 0-9/50", w.Header().Get("Content-Range"))
	repo.AssertExpectations(t)
}

// TestListVisits_WithRange verifies that a valid range=[5,10] query parameter
// fetches exactly 6 records (inclusive range) and sets the correct Content-Range header.
func TestListVisits_WithRange(t *testing.T) {
	repo := new(mockVisitRepo)
	h := handler.NewLinkVisitHandler(repo)

	// Range [5,10] is inclusive: 5,6,7,8,9,10 → 6 items.
	expected := make([]*repository.LinkVisit, 6)
	repo.On("List", mock.Anything, int32(6), int32(5)).Return(expected, nil)
	repo.On("Count", mock.Anything).Return(int64(20), nil)

	w := httptest.NewRecorder()
	c, _ := createTestContext(w, "GET", "/api/link_visits?range=[5,10]", nil)
	h.List(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "link_visits 5-10/20", w.Header().Get("Content-Range"))
	repo.AssertExpectations(t)
}
