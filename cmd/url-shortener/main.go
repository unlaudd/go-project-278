// В NewRouter(), после инициализации Sentry:

// Подключаем БД
dbConn, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
if err != nil {
	log.Fatalf("Failed to connect to database: %v", err)
}
if err := dbConn.Ping(); err != nil {
	log.Fatalf("Failed to ping database: %v", err)
}
log.Println("[DB] Connected")

repo := repository.NewLinkRepository(dbConn)
baseURL := os.Getenv("BASE_URL")
if baseURL == "" {
	baseURL = "https://url-shortener-452x.onrender.com" // фолбэк на Render
}
linkHandler := handler.NewLinkHandler(repo, baseURL)

// API группа
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
}

// 🔗 Редирект с короткой ссылки на оригинал
router.GET("/r/:shortName", func(c *gin.Context) {
	shortName := c.Param("shortName")
	link, err := repo.GetByShortName(c.Request.Context(), shortName, "")
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "link not found"})
		return
	}
	c.Redirect(http.StatusMovedPermanently, link.OriginalURL)
})
