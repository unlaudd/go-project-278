//go:build unit
// +build unit

package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"url-shortener/url-shortener/internal/handler"
	"url-shortener/url-shortener/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// 🔹 Моковый репозиторий
type mockLinkRepo struct {
	mock.Mock
}

func (m *mockLinkRepo) Create(ctx context.Context, link *repository.Link, baseURL string) error {
	args := m.Called(ctx, link, baseURL)
	return args.Error(0)
}

func (m *mockLinkRepo) GetByID(ctx context.Context, id int64, baseURL string) (*repository.Link, error) {
	args := m.Called(ctx, id, baseURL)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.Link), args.Error(1)
}

func (m *mockLinkRepo) GetByShortName(ctx context.Context, shortName string, baseURL string) (*repository.Link, error) {
	args := m.Called(ctx, shortName, baseURL)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.Link), args.Error(1)
}

func (m *mockLinkRepo) List(ctx context.Context, limit, offset int, baseURL string) ([]*repository.Link, error) {
	args := m.Called(ctx, limit, offset, baseURL)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*repository.Link), args.Error(1)
}

func (m *mockLinkRepo) Update(ctx context.Context, id int64, originalURL, shortName *string, baseURL string) (*repository.Link, error) {
	args := m.Called(ctx, id, originalURL, shortName, baseURL)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.Link), args.Error(1)
}

func (m *mockLinkRepo) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// 🔹 Вспомогательная функция для создания тестового контекста Gin
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

// 🔹 Тест: POST /api/links — создание ссылки
func TestCreateLink(t *testing.T) {
	repo := new(mockLinkRepo)
	h := handler.NewLinkHandler(repo, "https://test.local")

	expectedLink := &repository.Link{
		ID:          1,
		OriginalURL: "https://example.com",
		ShortName:   "abc123",
		ShortURL:    "https://test.local/r/abc123",
	}

	repo.On("Create", mock.Anything, mock.AnythingOfType("*repository.Link"), "https://test.local").
		Run(func(args mock.Arguments) {
			// Проверяем, что short_name сгенерировался, если не передан
			link := args.Get(1).(*repository.Link)
			assert.Equal(t, "https://example.com", link.OriginalURL)
		}).
		Return(nil)

	w := httptest.NewRecorder()
	c, _ := createTestContext(w, "POST", "/api/links", []byte(`{"original_url":"https://example.com"}`))

	h.Create(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	repo.AssertExpectations(t)
}

// 🔹 Тест: GET /api/links/:id — получение по ID
func TestGetLinkByID(t *testing.T) {
	repo := new(mockLinkRepo)
	h := handler.NewLinkHandler(repo, "https://test.local")

	expectedLink := &repository.Link{
		ID:          1,
		OriginalURL: "https://example.com",
		ShortName:   "abc123",
		ShortURL:    "https://test.local/r/abc123",
	}

	repo.On("GetByID", mock.Anything, int64(1), "https://test.local").Return(expectedLink, nil)

	w := httptest.NewRecorder()
	c, _ := createTestContext(w, "GET", "/api/links/1", nil)
	c.AddParam("id", "1")

	h.GetByID(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp repository.Link
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Equal(t, expectedLink.ID, resp.ID)
	repo.AssertExpectations(t)
}

// 🔹 Тест: GET /api/links/:id — 404 если не найдено
func TestGetLinkByID_NotFound(t *testing.T) {
	repo := new(mockLinkRepo)
	h := handler.NewLinkHandler(repo, "https://test.local")

	repo.On("GetByID", mock.Anything, int64(999), "https://test.local").
		Return((*repository.Link)(nil), errors.New("link not found"))

	w := httptest.NewRecorder()
	c, _ := createTestContext(w, "GET", "/api/links/999", nil)
	c.AddParam("id", "999")

	h.GetByID(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	repo.AssertExpectations(t)
}

// 🔹 Тест: GET /api/links — список ссылок
func TestListLinks(t *testing.T) {
	repo := new(mockLinkRepo)
	h := handler.NewLinkHandler(repo, "https://test.local")

	expectedLinks := []*repository.Link{
		{ID: 1, OriginalURL: "https://example.com/1", ShortName: "abc1", ShortURL: "https://test.local/r/abc1"},
		{ID: 2, OriginalURL: "https://example.com/2", ShortName: "abc2", ShortURL: "https://test.local/r/abc2"},
	}

	repo.On("List", mock.Anything, 100, 0, "https://test.local").Return(expectedLinks, nil)

	w := httptest.NewRecorder()
	c, _ := createTestContext(w, "GET", "/api/links", nil)

	h.List(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp []*repository.Link
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Len(t, resp, 2)
	repo.AssertExpectations(t)
}

// 🔹 Тест: PUT /api/links/:id — обновление
func TestUpdateLink(t *testing.T) {
	repo := new(mockLinkRepo)
	h := handler.NewLinkHandler(repo, "https://test.local")

	expectedLink := &repository.Link{
		ID:          1,
		OriginalURL: "https://new.example.com",
		ShortName:   "abc123",
		ShortURL:    "https://test.local/r/abc123",
	}

	repo.On("Update", mock.Anything, int64(1), mock.Anything, mock.Anything, "https://test.local").
		Return(expectedLink, nil)

	w := httptest.NewRecorder()
	c, _ := createTestContext(w, "PUT", "/api/links/1", []byte(`{"original_url":"https://new.example.com"}`))
	c.AddParam("id", "1")

	h.Update(c)

	assert.Equal(t, http.StatusOK, w.Code)
	repo.AssertExpectations(t)
}

// 🔹 Тест: DELETE /api/links/:id — удаление
func TestDeleteLink(t *testing.T) {
	repo := new(mockLinkRepo)
	h := handler.NewLinkHandler(repo, "https://test.local")

	repo.On("Delete", mock.Anything, int64(1)).Return(nil)

	w := httptest.NewRecorder()
	c, _ := createTestContext(w, "DELETE", "/api/links/1", nil)
	c.AddParam("id", "1")

	h.Delete(c)

	assert.Equal(t, http.StatusNoContent, w.Code)
	repo.AssertExpectations(t)
}

// 🔹 Тест: Конфликт short_name при создании
func TestCreateLink_Conflict(t *testing.T) {
	repo := new(mockLinkRepo)
	h := handler.NewLinkHandler(repo, "https://test.local")

	repo.On("Create", mock.Anything, mock.Anything, "https://test.local").
		Return(errors.New("short_name already exists"))

	w := httptest.NewRecorder()
	c, _ := createTestContext(w, "POST", "/api/links", []byte(`{"original_url":"https://example.com","short_name":"taken"}`))

	h.Create(c)

	assert.Equal(t, http.StatusConflict, w.Code)
	repo.AssertExpectations(t)
}

// 🔹 Тест: пагинация по умолчанию (без параметра range)
func TestListLinks_DefaultPagination(t *testing.T) {
	repo := new(mockLinkRepo)
	h := handler.NewLinkHandler(repo, "https://test.local")

	expectedLinks := make([]*repository.Link, 10) // дефолт: 10 записей [0,9]
	for i := range expectedLinks {
		expectedLinks[i] = &repository.Link{ID: int32(i + 1)}
	}

	repo.On("List", mock.Anything, int32(10), int32(0), "https://test.local").Return(expectedLinks, nil)
	repo.On("Count", mock.Anything).Return(int64(42), nil)

	w := httptest.NewRecorder()
	c, _ := createTestContext(w, "GET", "/api/links", nil)

	h.List(c)

	assert.Equal(t, http.StatusOK, w.Code)
	// Проверяем заголовок Content-Range
	assert.Equal(t, "links 0-9/42", w.Header().Get("Content-Range"))
	repo.AssertExpectations(t)
}

// 🔹 Тест: пагинация с параметром range=[5,10]
func TestListLinks_WithRange(t *testing.T) {
	repo := new(mockLinkRepo)
	h := handler.NewLinkHandler(repo, "https://test.local")

	expectedLinks := make([]*repository.Link, 6) // [5,10] → 6 записей
	for i := range expectedLinks {
		expectedLinks[i] = &repository.Link{ID: int32(i + 5)}
	}

	repo.On("List", mock.Anything, int32(6), int32(5), "https://test.local").Return(expectedLinks, nil)
	repo.On("Count", mock.Anything).Return(int64(20), nil)

	w := httptest.NewRecorder()
	c, _ := createTestContext(w, "GET", "/api/links?range=[5,10]", nil)

	h.List(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "links 5-10/20", w.Header().Get("Content-Range"))
	repo.AssertExpectations(t)
}

// 🔹 Тест: некорректный формат range → дефолтная пагинация
func TestListLinks_InvalidRange_UsesDefault(t *testing.T) {
	repo := new(mockLinkRepo)
	h := handler.NewLinkHandler(repo, "https://test.local")

	expectedLinks := make([]*repository.Link, 10)
	repo.On("List", mock.Anything, int32(10), int32(0), "https://test.local").Return(expectedLinks, nil)
	repo.On("Count", mock.Anything).Return(int64(15), nil)

	w := httptest.NewRecorder()
	// Некорректный формат: нет квадратных скобок
	c, _ := createTestContext(w, "GET", "/api/links?range=5,10", nil)

	h.List(c)

	assert.Equal(t, http.StatusOK, w.Code)
	// Должна сработать дефолтная пагинация [0,9]
	assert.Equal(t, "links 0-9/15", w.Header().Get("Content-Range"))
	repo.AssertExpectations(t)
}

// 🔹 Тест: пустой результат пагинации
func TestListLinks_EmptyResult(t *testing.T) {
	repo := new(mockLinkRepo)
	h := handler.NewLinkHandler(repo, "https://test.local")

	repo.On("List", mock.Anything, int32(10), int32(100), "https://test.local").Return([]*repository.Link{}, nil)
	repo.On("Count", mock.Anything).Return(int64(50), nil)

	w := httptest.NewRecorder()
	c, _ := createTestContext(w, "GET", "/api/links?range=[100,109]", nil)

	h.List(c)

	assert.Equal(t, http.StatusOK, w.Code)
	// Пустой результат: start-(start-1)/total
	assert.Equal(t, "links 100-99/50", w.Header().Get("Content-Range"))
	repo.AssertExpectations(t)
}

// 🔹 Тест: парсинг range
func TestParseRange(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantStart int
		wantEnd   int
		wantErr   bool
	}{
		{"valid no spaces", "[0,10]", 0, 10, false},
		{"valid with spaces", "[5, 15]", 5, 15, false},
		{"single element", "[3,3]", 3, 3, false},
		{"empty", "", 0, 0, true},
		{"no brackets", "0,10", 0, 0, true},
		{"negative start", "[-1,10]", 0, 0, true},
		{"end before start", "[10,5]", 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end, err := handler.ParseRangeForTest(tt.input) // см. примечание ниже
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantStart, start)
				assert.Equal(t, tt.wantEnd, end)
			}
		})
	}
}

// Тест: невалидный URL
func TestCreateLink_InvalidURL(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := createTestContext(w, "POST", "/api/links", nil)
	c.Request.Body = io.NopCloser(strings.NewReader(`{"original_url":"not-a-url"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	h.Create(c)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	assert.Contains(t, w.Body.String(), "url")
}
