package main

import (
	"context"
	"fmt"
	"interview/internal/broker"
	"interview/internal/config"
	"interview/internal/db"
	"interview/internal/handlers"
	"interview/internal/repository"
	"interview/internal/service"
	"interview/internal/storage"
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

	//migrationsPath := "./migrations"
	//if err := db.RunMigrations(database); err != nil {
	//	log.Fatalf("Failed to migrate database: %v", err)
	//}
	// Инициализируем S3 storage
	//	log.Println("Initializing S3 storage...")
	//	s3Client := storage.NewS3Client(
	//		getEnv("S3_ENDPOINT", "http://localhost:9000"),
	//		getEnv("S3_REGION", "us-east-1"),
	//		getEnv("S3_ACCESS_KEY", "minioadmin"),
	//		getEnv("S3_SECRET_KEY", "minioadmin123"),
	//	)
	log.Println("Initializing S3 storage...")
	s3Client := storage.NewS3Client(
		getEnv("AWS_S3_ENDPOINT", "http://minio:9000"), // ← ИСПРАВЛЕНО
		getEnv("AWS_REGION", "ru-1"),
		getEnv("AWS_ACCESS_KEY_ID", "minioadmin"),
		getEnv("AWS_SECRET_ACCESS_KEY", "minioadmin123"),
	)

	//s3Storage := storage.NewS3Storage(
	//	s3Client,
	//	getEnv("S3_BUCKET", "interview-files"),
	//	getEnv("S3_REGION", "us-east-1"),
	//)
	//log.Println("S3 storage initialized")

	s3Storage := storage.NewS3Storage(
		s3Client,
		getEnv("AWS_S3_BUCKET", "interview-files"),
		getEnv("AWS_REGION", "ru-1"),
	)
	log.Println("S3 storage initialized")

	// 🚀 Инициализируем RabbitMQ Publisher
	log.Println("Initializing RabbitMQ publisher...")
	var publisher broker.Publisher // ← ИСПРАВЛЕНО: интерфейс

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

	// Создаем services с publisher
	log.Println("Initializing services...")
	vacancySvc := service.NewVacancyService(vacancyRepo, s3Storage)
	resumeSvc := service.NewResumeService(resumeRepo, s3Storage, vacancyRepo, publisher)
	interviewSvc := service.NewInterviewService(interviewRepo)
	log.Println("Services initialized")

	// Настраиваем Gin режим
	if getEnv("GIN_MODE", "debug") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Создаем роутер
	log.Println("Setting up routes...")
	router := handlers.SetupRouter(vacancySvc, resumeSvc, interviewSvc)
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
		log.Printf("S3 Endpoint: %s", getEnv("S3_ENDPOINT", "http://minio:9000"))
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

// getEnv возвращает значение переменной окружения или дефолтное значение
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// validateEnv проверяет обязательные переменные окружения
func validateRequiredEnvVars() error {
	required := []string{
		"DATABASE_URL",
		"S3_ACCESS_KEY",
		"S3_SECRET_KEY",
		"S3_BUCKET",
	}

	for _, env := range required {
		if os.Getenv(env) == "" {
			return fmt.Errorf("required environment variable %s is not set", env)
		}
	}

	return nil
}
