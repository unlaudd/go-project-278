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

// NewRouter initializes and configures the Gin router with middleware,
// database connections, and API route handlers.
func NewRouter() *gin.Engine {
	router := gin.Default()

	// Trust Cloudflare headers for accurate client IP resolution
	// when the app runs behind Render's proxy.
	router.TrustedPlatform = gin.PlatformCloudflare

	// Initialize Sentry for error tracking if DSN is provided.
	if dsn := os.Getenv("SENTRY_DSN"); dsn != "" {
		if err := sentry.Init(sentry.ClientOptions{
			Dsn:         dsn,
			Environment: os.Getenv("ENVIRONMENT"),
		}); err != nil {
			log.Printf("Sentry initialization failed: %v", err)
		} else {
			log.Println("Sentry initialized")
			router.Use(sentrygin.New(sentrygin.Options{}))
		}
	}

	// Configure CORS middleware for local development.
	// Allows requests from the frontend dev server at localhost:5173.
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		ExposeHeaders:    []string{"Content-Range"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Initialize database connection if DATABASE_URL is set.
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Println("DATABASE_URL not set, skipping DB setup (local mode)")
	} else {
		dbConn, err := sql.Open("postgres", dbURL)
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}
		if err := dbConn.Ping(); err != nil {
			log.Fatalf("Failed to ping database: %v", err)
		}
		log.Println("Database connected")

		baseURL := os.Getenv("BASE_URL")
		if baseURL == "" {
			baseURL = "http://localhost:8080"
		}

		repo := repository.NewLinkRepository(dbConn)
		visitRepo := repository.NewLinkVisitRepository(dbConn)
		linkHandler := handler.NewLinkHandler(repo, baseURL)

		// Register API route group.
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
			// Endpoint for listing link visits with pagination.
			api.GET("/link_visits", handler.NewLinkVisitHandler(visitRepo).List)
		}

		// Redirect endpoint with visit tracking.
		// Records analytics (IP, user agent, referer, status) before redirecting.
		router.GET("/r/:shortName", func(c *gin.Context) {
			shortName := c.Param("shortName")

			link, err := repo.GetByShortName(c.Request.Context(), shortName, baseURL)
			if err != nil {
				// Record failed lookup attempt (404) for analytics.
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

			// Record successful redirect (301) for analytics.
			if visitRepo != nil {
				err := visitRepo.Create(c.Request.Context(), &repository.LinkVisit{
					LinkID: link.ID, IP: c.ClientIP(), UserAgent: c.Request.UserAgent(),
					Referer: c.Request.Referer(), Status: http.StatusMovedPermanently,
				})
				if err != nil {
					log.Printf("Failed to record redirect visit: %v", err)
				}
			}

			c.Redirect(http.StatusMovedPermanently, link.OriginalURL)
		})
	}

	// Health check endpoint for load balancers and monitoring.
	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	// Debug endpoint for testing Sentry error reporting.
	// Should be disabled or protected in production.
	router.GET("/debug/error", func(c *gin.Context) {
		panic("test error for Sentry verification")
	})

	return router
}

func main() {
	router := NewRouter()

	// Determine backend port: use env var or default to 8080.
	// In production, Caddy proxies external traffic to this port.
	port := os.Getenv("BACKEND_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Backend starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
