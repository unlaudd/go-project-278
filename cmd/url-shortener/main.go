package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
	_ "github.com/lib/pq"

	"url-shortener/internal/handler"
	"url-shortener/internal/repository"
)

// NewRouter создаёт роутер с подключёнными middleware и API
func NewRouter() *gin.Engine {
	router := gin.Default()

	// 🔐 Доверять заголовкам Cloudflare (Render использует CF)
	router.TrustedPlatform = gin.PlatformCloudflare

	// 🪵 Sentry
	if dsn := os.Getenv("SENTRY_DSN"); dsn != "" {
		if err := sentry.Init(sentry.ClientOptions{
			Dsn:         dsn,
			Environment: os.Getenv("ENVIRONMENT"),
		}); err != nil {
			log.Printf("[Sentry] Init failed: %v", err)
		} else {
			log.Println("[Sentry] Initialized")
			router.Use(sentrygin.New(sentrygin.Options{}))
		}
	}

	// 🔐 CORS для локальной разработки
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		ExposeHeaders:    []string{"Content-Range"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// 🗃️ Подключаем БД
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Println("[DB] DATABASE_URL not set, skipping DB setup (local mode)")
	} else {
		dbConn, err := sql.Open("postgres", dbURL)
		if err != nil {
			log.Fatalf("Failed to connect to DB: %v", err)
		}
		if err := dbConn.Ping(); err != nil {
			log.Fatalf("Failed to ping DB: %v", err)
		}
		log.Println("[DB] Connected")

		baseURL := os.Getenv("BASE_URL")
		if baseURL == "" {
			baseURL = "http://localhost:8080"
		}

		repo := repository.NewLinkRepository(dbConn)
		visitRepo := repository.NewLinkVisitRepository(dbConn)
		linkHandler := handler.NewLinkHandler(repo, baseURL)

		// API маршруты
		api := router.Group("/api")
		{
			links := api.Group("/links")
			{
				links.POST("", linkHandler.Create)
				links.GET("", linkHandler.List)
				links.GET("/:id", linkHandler.GetByID)
				links.PUT("/:id", linkHandler.Update)
				links.DELETE("/:id", linkHandler.Delete)
			}
			api.GET("/link_visits", handler.NewLinkVisitHandler(visitRepo).List)
		}

		// Редирект с записью посещения
		router.GET("/r/:shortName", func(c *gin.Context) {
			shortName := c.Param("shortName")

			// Ищем ссылку
			link, err := repo.GetByShortName(c.Request.Context(), shortName, baseURL)
			if err != nil {
				// Записываем попытку доступа к несуществующей ссылке (404)
				if visitRepo != nil {
					err := visitRepo.Create(c.Request.Context(), &repository.LinkVisit{
						LinkID: 0, IP: c.ClientIP(), UserAgent: c.Request.UserAgent(),
						Referer: c.Request.Referer(), Status: http.StatusNotFound,
					})
					if err != nil {
						log.Printf("Failed to record 404 visit: %v", err)
					}
				}
				c.JSON(http.StatusNotFound, gin.H{"error": "link not found"})
				return
			}

			// Записываем успешное посещение (301)
			if visitRepo != nil {
				err := visitRepo.Create(c.Request.Context(), &repository.LinkVisit{
					LinkID: link.ID, IP: c.ClientIP(), UserAgent: c.Request.UserAgent(),
					Referer: c.Request.Referer(), Status: http.StatusMovedPermanently,
				})
				if err != nil {
					log.Printf("Failed to record redirect visit: %v", err)
				}
			}

			// Выполняем редирект
			c.Redirect(http.StatusMovedPermanently, link.OriginalURL)
		})
	}

	// ✅ Старый эндпоинт /ping
	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	// 🧪 Тестовый эндпоинт Sentry (раскомментируйте при необходимости)
	router.GET("/debug/error", func(c *gin.Context) {
		panic("test error for Sentry verification")
	})

	return router
}

func main() {
	router := NewRouter()

	// В продакшене слушаем на 8080 (Caddy проксирует на этот порт)
	port := os.Getenv("BACKEND_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Backend starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
