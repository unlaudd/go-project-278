package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// NewRouter создаёт и настраивает роутер
func NewRouter() *gin.Engine {
	router := gin.Default() // включает Logger + Recovery

	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	return router
}

func main() {
	router := NewRouter()

	// Проверяем ошибку запуска сервера
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Не удалось запустить сервер: %v", err)
	}
}
