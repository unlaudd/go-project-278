package main

import (
	"log"
	"net/http"
	"os"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
)

// NewRouter создаёт роутер с подключёнными middleware
func NewRouter() *gin.Engine {
	router := gin.Default()

	// 🪵 Подключаем Sentry (только если задан DSN)
	if dsn := os.Getenv("SENTRY_DSN"); dsn != "" {
		err := sentry.Init(sentry.ClientOptions{
			Dsn:              dsn,
			EnableTracing:    false, // можно включить позже
			TracesSampleRate: 0,
			Environment:      os.Getenv("ENVIRONMENT"), // "production", "staging" и т.д.
		})
		if err != nil {
			log.Printf("[Sentry] Init failed: %v", err)
		} else {
			log.Println("[Sentry] Initialized")
			// Подключаем middleware для Gin
			router.Use(sentrygin.New(sentrygin.Options{}))
		}
	}

	// ✅ Эндпоинт /ping
	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	// 🧪 Тестовый эндпоинт для генерации ошибки (только для проверки!)
	//router.GET("/debug/error", func(c *gin.Context) {
	//	// Имитируем панику — Sentry должен её перехватить
	//	panic("test error for Sentry verification")
	//})

	return router
}

func main() {
	router := NewRouter()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Graceful shutdown для отправки оставшихся событий в Sentry
	go func() {
		if err := router.Run(":" + port); err != nil {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Ожидаем сигнал завершения
	// (в реальном проекте добавьте обработку os.Interrupt, syscall.SIGTERM)
	select {}
}
