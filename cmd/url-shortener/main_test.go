package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Инициализация перед тестами
func setupTest() {
	gin.SetMode(gin.TestMode) // отключаем логирование в тестах
}

func TestPingEndpoint(t *testing.T) {
	setupTest()

	router := NewRouter()

	// Создаём тестовый запрос
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/ping", nil)

	// Выполняем запрос через роутер
	router.ServeHTTP(w, req)

	// Проверяем ответ
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "pong", w.Body.String())
	assert.Equal(t, "text/plain; charset=utf-8", w.Header().Get("Content-Type"))
}
