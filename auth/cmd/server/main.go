package main

import (
	"ai-hr-service/internal/auth"
	"ai-hr-service/internal/config"
	"ai-hr-service/internal/database"
	"ai-hr-service/internal/middleware"
	"ai-hr-service/internal/utils"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	// грузим кфг для БДшки
	cfg := config.Load()

	// Инициализируем JWT конфигурацию
	utils.InitJWT(&cfg.JWT)

	db := database.Connect(cfg.Database) // коннект к базе по данным из кфг

	// создаем сервисы
	authRepo := auth.NewRepository(db)
	tokenService := auth.NewTokenService(db)               // НОВОЕ: сервис токенов
	authService := auth.NewService(authRepo, tokenService) // ОБНОВЛЕНО: передаем tokenService
	authHandler := auth.NewHandler(authService)

	// создаем middleware
	authMiddleware := middleware.NewAuthMiddleware(db) // НОВОЕ: обновленный middleware

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
			"time":    time.Now().Format(time.RFC3339),
		})
	})

	// API роуты - тут вся магия происходит
	api := r.Group("/api")
	{
		// паблик роуты - сюда может любой зайти
		authGroup := api.Group("/auth")
		{
			authGroup.POST("/register", authHandler.Register)
			authGroup.POST("/login", authHandler.Login)
			authGroup.POST("/refresh", authHandler.RefreshToken) // НОВОЕ: обновление токенов
		}

		// протектед роуты - только для авторизованных пользователей
		protected := api.Group("/")
		protected.Use(authMiddleware.TokenAuth()) // ОБНОВЛЕНО: новый middleware
		{
			protected.GET("/profile", authHandler.GetProfile) // Профиль юзера

			// НОВЫЕ: управление сессиями
			protected.POST("/auth/logout", authHandler.Logout)        // Выход
			protected.POST("/auth/logout-all", authHandler.LogoutAll) // Выход со всех устройств

			// Тестовый защищенный роут
			protected.GET("/protected", func(c *gin.Context) {
				username, _ := c.Get("username")
				role, _ := c.Get("role")
				utils.SuccessResponse(c, http.StatusOK, gin.H{
					"message":  "This is a protected endpoint",
					"username": username,
					"role":     role,
				})
			})
		}

		// HR роуты только для HR-ов
		hr := api.Group("/hr")
		hr.Use(authMiddleware.TokenAuth(), authMiddleware.RequireRole("hr_specialist")) // ОБНОВЛЕНО: новая роль
		{
			hr.GET("/dashboard", func(c *gin.Context) {
				username, _ := c.Get("username")
				c.JSON(200, gin.H{
					"message": "Welcome to HR Dashboard",
					"user":    username,
				})
			})

			// НОВОЕ: дополнительные HR endpoints
			hr.GET("/vacancies", func(c *gin.Context) {
				c.JSON(200, gin.H{
					"message": "HR Vacancies",
					"data":    []string{"Vacancy 1", "Vacancy 2"}, // заглушка
				})
			})
		}

		// НОВОЕ: роуты для кандидатов
		candidate := api.Group("/candidate")
		candidate.Use(authMiddleware.TokenAuth(), authMiddleware.RequireRole("candidate"))
		{
			candidate.GET("/interviews", func(c *gin.Context) {
				username, _ := c.Get("username")
				c.JSON(200, gin.H{
					"message": "Your interviews",
					"user":    username,
				})
			})
		}

		// НОВОЕ: админские роуты
		admin := api.Group("/admin")
		admin.Use(authMiddleware.TokenAuth(), authMiddleware.RequireRole("admin"))
		{
			admin.GET("/users", func(c *gin.Context) {
				c.JSON(200, gin.H{
					"message": "All users (admin only)",
					"admin":   c.MustGet("username"),
				})
			})

			admin.GET("/stats", func(c *gin.Context) {
				userID := c.MustGet("user_id").(uint)
				count, _ := tokenService.GetUserActiveTokensCount(userID)

				c.JSON(200, gin.H{
					"message":       "System stats",
					"active_tokens": count,
				})
			})
		}
	}

	// НОВОЕ: Добавляем периодическую очистку просроченных токенов
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := tokenService.CleanupExpiredTokens(); err != nil {
					log.Printf("Failed to cleanup expired tokens: %v", err)
				} else {
					log.Println("Expired tokens cleaned up successfully")
				}
			}
		}
	}()

	//старт
	log.Printf("🚀 Server starting on port %s", cfg.Server.Port)
	log.Printf("💾 Database: %s:%s/%s", cfg.Database.Host, cfg.Database.Port, cfg.Database.DBName)
	log.Printf("🔐 JWT Access TTL: %s", cfg.JWT.AccessTokenTTL)
	log.Printf("🔄 JWT Refresh TTL: %s", cfg.JWT.RefreshTokenTTL)

	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
