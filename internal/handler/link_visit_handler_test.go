//go:build unit
// +build unit

package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"url-shortener/internal/handler"
	"url-shortener/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Моковый репозиторий посещений
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

func TestListVisits_WithRange(t *testing.T) {
	repo := new(mockVisitRepo)
	h := handler.NewLinkVisitHandler(repo)

	expected := make([]*repository.LinkVisit, 6) // [5,10] → 6 записей
	repo.On("List", mock.Anything, int32(6), int32(5)).Return(expected, nil)
	repo.On("Count", mock.Anything).Return(int64(20), nil)

	w := httptest.NewRecorder()
	c, _ := createTestContext(w, "GET", "/api/link_visits?range=[5,10]", nil)
	h.List(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "link_visits 5-10/20", w.Header().Get("Content-Range"))
	repo.AssertExpectations(t)
}
