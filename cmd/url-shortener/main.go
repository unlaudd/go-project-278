// cmd/url-shortener/main.go
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
)

// NewRouter создаёт роутер с подключёнными middleware.
// Вынесен отдельно для удобства тестирования.
func NewRouter() *gin.Engine {
	router := gin.Default() // включает Logger + Recovery

	// 🪵 Подключаем Sentry, если задан DSN
	if dsn := os.Getenv("SENTRY_DSN"); dsn != "" {
		err := sentry.Init(sentry.ClientOptions{
			Dsn:         dsn,
			Environment: os.Getenv("ENVIRONMENT"),
		})
		if err != nil {
			log.Printf("[Sentry] Init failed: %v", err)
		} else {
			log.Println("[Sentry] Initialized")
			router.Use(sentrygin.New(sentrygin.Options{}))
		}
	}

	// ✅ Эндпоинт /ping
	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	// 🧪 Тестовый эндпоинт для проверки Sentry
	router.GET("/debug/error", func(c *gin.Context) {
	 	panic("test error for Sentry verification")
	})

	return router
}

func main() {
	router := NewRouter()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Запускаем сервер
	// Ошибка запуска логируется и завершает процесс
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
