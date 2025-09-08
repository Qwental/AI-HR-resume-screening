package main

import (
	"context"
	"github.com/gin-contrib/cors"
	"interview/internal/broker"
	"interview/internal/config"
	"interview/internal/db"
	"interview/internal/handlers"
	"interview/internal/repository"
	"interview/internal/service"
	"interview/internal/storage"
	"interview/middleware"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	log.Println("Starting Interview Service...")

	// Загружаем конфиг
	cfg := config.Load()

	// Подключаемся к БД
	log.Println("Connecting to database...")
	database := db.Connect(cfg.Database)
	log.Println("Database connected successfully")

	// Инициализируем S3 storage
	log.Println("Initializing S3 storage...")
	s3Client := storage.NewS3Client(
		getEnv("AWS_S3_ENDPOINT", "http://minio:9000"),
		getEnv("AWS_REGION", "ru-1"),
		getEnv("AWS_ACCESS_KEY_ID", "minioadmin"),
		getEnv("AWS_SECRET_ACCESS_KEY", "minioadmin123"),
	)

	s3Storage := storage.NewS3Storage(
		s3Client,
		getEnv("AWS_S3_BUCKET", "interview-files"),
		getEnv("AWS_REGION", "ru-1"),
	)
	log.Println("S3 storage initialized")

	// Инициализируем RabbitMQ Publisher
	log.Println("Initializing RabbitMQ publisher...")
	var publisher broker.Publisher
	if rabbitmqPublisher, err := broker.NewRabbitMQPublisher(
		getEnv("RABBITMQ_URL", "amqp://guest:guest@rabbitmq:5672/"),
		getEnv("RABBITMQ_EXCHANGE", "resume_exchange"),
		getEnv("RABBITMQ_QUEUE", "resume_analysis_queue"),
	); err != nil {
		log.Printf("Failed to create RabbitMQ publisher: %v", err)
		log.Println("Using NullPublisher as fallback...")
		publisher = broker.NewNullPublisher()
	} else {
		log.Println("RabbitMQ publisher initialized")
		publisher = rabbitmqPublisher
	}

	// Создаем repositories
	log.Println("Initializing repositories...")
	vacancyRepo := repository.NewVacancyRepository(database)
	resumeRepo := repository.NewResumeRepository(database)
	interviewRepo := repository.NewInterviewRepository(database)
	log.Println("Repositories initialized")

	// Создаем services
	log.Println("Initializing services...")
	vacancySvc := service.NewVacancyService(vacancyRepo, s3Storage)
	service.NewResumeService(resumeRepo, s3Storage, vacancyRepo, publisher)
	service.NewInterviewService(interviewRepo)
	log.Println("Services initialized")

	// Настраиваем Gin режим
	if getEnv("GIN_MODE", "debug") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Создаем роутер
	router := gin.Default()

	router.Use(middleware.CorsMiddleware())

	// Настройка CORS для работы с Next.js
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://frontend:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "Interview Service is running!",
			"time":    time.Now().Format(time.RFC3339),
		})
	})

	// 🔥 ИСПРАВЛЕНИЕ: Создаем handler и настраиваем роуты
	log.Println("Setting up routes...")

	// Создаем handler
	vacancyHandler := handlers.NewVacancyHandler(vacancySvc)

	// Группа API роутов

	api := router.Group("/api")
	{
		// Защищенная группа роутов
		authorized := api.Group("/")
		// Применяем middleware для проверки токена
		authorized.Use(middleware.TokenAuthMiddleware())
		{
			// Роуты для вакансий
			vacancies := authorized.Group("/vacancies")
			{
				// Для создания вакансии нужна роль hr_specialist
				vacancies.POST("", middleware.RequireRoleMiddleware("hr_specialist"), vacancyHandler.Create)

				// Для просмотра списка вакансий достаточно быть авторизованным
				vacancies.GET("", vacancyHandler.GetAll)
				vacancies.GET("/:id", vacancyHandler.GetByID)
				vacancies.GET("/:id/download", vacancyHandler.GetDownloadLink)

				// Для изменения и удаления тоже нужна роль HR
				hrActions := vacancies.Group("")
				hrActions.Use(middleware.RequireRoleMiddleware("hr_specialist"))
				{
					hrActions.PUT("/:id", vacancyHandler.Update)
					hrActions.PUT("/:id/file", vacancyHandler.UpdateWithFile)
					hrActions.DELETE("/:id", vacancyHandler.Delete)
				}
			}
		}
	}

	//api := router.Group("/api")
	//{
	//	// Вакансии
	//	vacancies := api.Group("/vacancies")
	//	{
	//		vacancies.POST("", vacancyHandler.Create)
	//		vacancies.GET("", vacancyHandler.GetAll)
	//		vacancies.GET("/:id", vacancyHandler.GetByID)
	//		vacancies.GET("/:id/download", vacancyHandler.GetDownloadLink)
	//		vacancies.PUT("/:id", vacancyHandler.Update)
	//		vacancies.PUT("/:id/file", vacancyHandler.UpdateWithFile)
	//		vacancies.DELETE("/:id", vacancyHandler.Delete)
	//	}
	//
	//}

	log.Println("Routes configured")

	// Настраиваем сервер
	port := getEnv("PORT", "8081")
	server := &http.Server{
		Addr:           ":" + port,
		Handler:        router,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		IdleTimeout:    120 * time.Second,
		MaxHeaderBytes: (1 << 20) * 10, // 10MB
	}

	// Канал для graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Запускаем сервер в отдельной горутине
	go func() {
		log.Printf("Interview Service starting on port %s", port)
		log.Printf("Environment: %s", getEnv("GIN_MODE", "debug"))
		log.Printf("Database: %s:%s/%s", cfg.Database.Host, cfg.Database.Port, cfg.Database.DBName)
		log.Printf("S3 Endpoint: %s", getEnv("AWS_S3_ENDPOINT", "http://minio:9000"))
		log.Printf("RabbitMQ: %s", getEnv("RABBITMQ_URL", "amqp://guest:guest@rabbitmq:5672/"))

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Ждем сигнал завершения
	<-quit
	log.Println("Server is shutting down...")

	// Graceful shutdown с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped gracefully")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
