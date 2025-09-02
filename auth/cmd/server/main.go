package main

import (
	"ai-hr-service/internal/auth"
	"ai-hr-service/internal/config"
	"ai-hr-service/internal/database"
	"ai-hr-service/internal/middleware"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	// грузим кфг для БДшки
	cfg := config.Load()
	db := database.Connect(cfg.Database) // коннект к базе по данным из кфг

	// создаем сервисы
	authRepo := auth.NewRepository(db)          // сервис-бд
	authService := auth.NewService(authRepo)    // бизнес-логика
	authHandler := auth.NewHandler(authService) // обработчик http

	// роутер
	r := gin.Default()

	// стаст файлы для фронта
	r.Static("/static", "./static")
	r.StaticFile("/", "./static/index.html")
	r.StaticFile("/login", "./static/login.html")
	r.StaticFile("/register", "./static/register.html")
	r.StaticFile("/dashboard", "./static/dashboard.html")

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "AI-HR-SERVICE is running!",
		})
	})

	// API роуты - тут вся магия происходит
	api := r.Group("/api")
	{
		// паблик роуты - сюда может любой  зайти
		authGroup := api.Group("/auth")
		{
			authGroup.POST("/register", authHandler.Register)
			authGroup.POST("/login", authHandler.Login)
		}

		// протектед роуты - только для авторизованных пользователей
		protected := api.Group("/")
		protected.Use(middleware.AuthMiddleware()) // Middleware проверяет токен
		{
			protected.GET("/profile", authHandler.GetProfile)  // Профиль юзера
			protected.GET("/protected", authHandler.Protected) // Тестовый защищенный роут
		}

		// HR роуты только для HR-ов
		hr := api.Group("/hr")
		hr.Use(middleware.AuthMiddleware(), middleware.RequireRole("hr"))
		{
			hr.GET("/dashboard", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "Welcome to HR Dashboard"})
			})
		}
	}
	//старт
	log.Printf("Server starting on port %s", cfg.Server.Port)
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
