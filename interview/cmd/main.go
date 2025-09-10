package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"interview/internal/broker"
	"interview/internal/config"
	"interview/internal/db"
	"interview/internal/handlers"
	"interview/internal/repository"
	"interview/internal/service"
	"interview/internal/storage"
	"interview/middleware"
)

func main() {
	log.Println("🚀 Starting AI-HR Interview Service...")

	// ====================================
	// 1. ЗАГРУЗКА КОНФИГУРАЦИИ
	// ====================================
	log.Println("📁 Loading configuration...")
	cfg := config.Load()

	// ====================================
	// 2. ПОДКЛЮЧЕНИЕ К БАЗЕ ДАННЫХ
	// ====================================
	log.Println("🗄️ Connecting to PostgreSQL database...")
	database := db.Connect(cfg.Database)
	log.Println("✅ Database connected successfully")

	// ====================================
	// 3. ИНИЦИАЛИЗАЦИЯ S3 ХРАНИЛИЩА
	// ====================================
	log.Println("🗃️ Initializing S3 storage...")
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
	log.Println("✅ S3 storage initialized")

	// ====================================
	// 4. ИНИЦИАЛИЗАЦИЯ RABBITMQ
	// ====================================
	log.Println("🐰 Initializing RabbitMQ publisher...")
	var publisher broker.Publisher

	if rabbitmqPublisher, err := broker.NewRabbitMQPublisher(
		getEnv("RABBITMQ_URL", "amqp://guest:guest@rabbitmq:5672/"),
		getEnv("RABBITMQ_EXCHANGE", "resume_exchange"),
		getEnv("RABBITMQ_QUEUE", "resume_analysis_queue"),
	); err != nil {
		log.Printf("⚠️ Failed to create RabbitMQ publisher: %v", err)
		log.Println("🔄 Using NullPublisher as fallback...")
		publisher = broker.NewNullPublisher()
	} else {
		log.Println("✅ RabbitMQ publisher initialized")
		publisher = rabbitmqPublisher
	}

	// ====================================
	// 5. СОЗДАНИЕ РЕПОЗИТОРИЕВ (DATA LAYER)
	// ====================================
	log.Println("🗂️ Initializing repositories...")
	vacancyRepo := repository.NewVacancyRepository(database)
	resumeRepo := repository.NewResumeRepository(database)
	//interviewRepo := repository.NewInterviewRepository(database)
	log.Println("✅ Repositories initialized")

	// ====================================
	// 6. СОЗДАНИЕ СЕРВИСОВ (BUSINESS LOGIC)
	// ====================================
	log.Println("⚙️ Initializing services...")
	vacancySvc := service.NewVacancyService(vacancyRepo, s3Storage)
	resumeSvc := service.NewResumeService(resumeRepo, s3Storage, vacancyRepo, publisher)
	//interviewSvc := service.NewInterviewService(interviewRepo)
	log.Println("✅ Services initialized")

	// ====================================
	// 7. НАСТРОЙКА GIN ФРЕЙМВОРКА
	// ====================================
	if getEnv("GIN_MODE", "debug") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// ====================================
	// 8. СОЗДАНИЕ РОУТЕРА И MIDDLEWARE
	// ====================================
	router := gin.Default()

	// Базовый CORS middleware
	router.Use(middleware.CorsMiddleware())

	// Расширенная настройка CORS для работы с Next.js фронтендом
	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:3000", // Next.js dev сервер
			"http://frontend:3000",  // Docker контейнер фронтенда
		},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{
			"Origin", "Content-Type", "Accept", "Authorization",
			"X-Requested-With", "Content-Length",
		},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// ====================================
	// 9. СОЗДАНИЕ HTTP HANDLERS
	// ====================================
	log.Println("🔗 Setting up handlers...")
	vacancyHandler := handlers.NewVacancyHandler(vacancySvc)
	resumeHandler := handlers.NewResumeHandler(resumeSvc)
	//interviewHandler := handlers.NewInterviewHandler(interviewSvc)

	// ====================================
	// 10. НАСТРОЙКА МАРШРУТОВ (ROUTES)
	// ====================================
	log.Println("🛣️ Setting up routes...")

	// Health check endpoint (без авторизации)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"message":   "AI-HR Interview Service is running!",
			"timestamp": time.Now().Format(time.RFC3339),
			"version":   "1.0.0",
		})
	})

	// Группа API маршрутов
	api := router.Group("/api")

	// Защищенная группа маршрутов (требует JWT токен)
	authorized := api.Group("/")
	authorized.Use(middleware.TokenAuthMiddleware())

	// ====================================
	// 10.1 МАРШРУТЫ ДЛЯ ВАКАНСИЙ
	// ====================================
	vacancies := authorized.Group("/vacancies")
	{
		// Создание вакансии (только для HR специалистов)
		vacancies.POST("",
			middleware.RequireRoleMiddleware("hr_specialist"),
			vacancyHandler.Create)

		// Просмотр списка вакансий (все авторизованные пользователи)
		vacancies.GET("", vacancyHandler.GetAll)
		vacancies.GET("/:id", vacancyHandler.GetByID)
		vacancies.GET("/:id/download", vacancyHandler.GetDownloadLink)

		// ✅ ДОБАВЬТЕ ЭТУ СТРОКУ - роут для получения резюме по вакансии
		vacancies.GET("/:id/resumes", resumeHandler.GetByVacancy)

		// Группа действий только для HR специалистов
		hrVacancyActions := vacancies.Group("")
		hrVacancyActions.Use(middleware.RequireRoleMiddleware("hr_specialist"))
		{
			hrVacancyActions.PUT("/:id", vacancyHandler.Update)
			hrVacancyActions.PUT("/:id/file", vacancyHandler.UpdateWithFile)
			hrVacancyActions.DELETE("/:id", vacancyHandler.Delete)
		}
	}

	// ====================================
	// 10.2 МАРШРУТЫ ДЛЯ РЕЗЮМЕ
	// ====================================
	resumes := authorized.Group("/resumes")
	{
		// Загрузка резюме (доступно всем авторизованным пользователям)
		resumes.POST("", resumeHandler.Create)
		resumes.GET("/:id", resumeHandler.GetByID)

		// Группа действий только для HR специалистов
		hrResumeActions := resumes.Group("")
		hrResumeActions.Use(middleware.RequireRoleMiddleware("hr_specialist"))
		{
			//hrResumeActions.GET("", resumeHandler.GetAll) // Список всех резюме
			hrResumeActions.GET("/:id/download", resumeHandler.GetDownloadLink)
			//hrResumeActions.PUT("/:id", resumeHandler.Update)
			hrResumeActions.DELETE("/:id", resumeHandler.Delete)
		}
	}

	log.Println("✅ Routes configured successfully")

	// ====================================
	// 11. НАСТРОЙКА HTTP СЕРВЕРА
	// ====================================
	port := getEnv("PORT", "8081")
	server := &http.Server{
		Addr:           ":" + port,
		Handler:        router,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		IdleTimeout:    120 * time.Second,
		MaxHeaderBytes: (1 << 20) * 10, // 10MB
	}

	// ====================================
	// 12. GRACEFUL SHUTDOWN
	// ====================================
	// Канал для получения сигналов завершения
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Запускаем сервер в отдельной горутине
	go func() {
		log.Println("🎯 Server configuration:")
		log.Printf("   ➤ Port: %s", port)
		log.Printf("   ➤ Environment: %s", getEnv("GIN_MODE", "debug"))
		log.Printf("   ➤ Database: %s:%s/%s", cfg.Database.Host, cfg.Database.Port, cfg.Database.DBName)
		log.Printf("   ➤ S3 Endpoint: %s", getEnv("AWS_S3_ENDPOINT", "http://minio:9000"))
		log.Printf("   ➤ RabbitMQ: %s", getEnv("RABBITMQ_URL", "amqp://guest:guest@rabbitmq:5672/"))

		log.Printf("🌟 AI-HR Interview Service started successfully on port %s", port)
		log.Printf("🌍 Available endpoints:")
		log.Printf("   ➤ Health Check: http://localhost:%s/health", port)
		log.Printf("   ➤ API Base: http://localhost:%s/api", port)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Server failed to start: %v", err)
		}
	}()

	// Ждем сигнал завершения
	<-quit
	log.Println("🛑 Shutting down server...")

	// Graceful shutdown с таймаутом 30 секунд
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("❌ Server forced to shutdown: %v", err)
	}

	log.Println("✅ AI-HR Interview Service stopped gracefully")
}

// ====================================
// ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ
// ====================================

// getEnv возвращает значение переменной окружения или значение по умолчанию
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
