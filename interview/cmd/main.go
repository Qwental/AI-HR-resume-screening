package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"interview/internal/db"
	"interview/internal/handlers"
	"interview/internal/repository"
	"interview/internal/service"
	"interview/internal/storage"
)

func main() {
	// Загружаем .env файл
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	log.Println("Starting application...")

	// Запускаем миграции
	log.Println("Running database migrations...")
	db.RunMigrations()
	log.Println("Migrations completed")

	// Подключаемся к БД
	log.Println("Connecting to database...")
	database, err := db.NewDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Println("Database connected successfully")

	// Инициализируем S3 storage
	log.Println("Initializing S3 storage...")
	s3Client := storage.NewS3Client(
		getEnv("S3_ENDPOINT", ""),
		getEnv("S3_REGION", "ru-1"),
		getEnv("S3_ACCESS_KEY", ""),
		getEnv("S3_SECRET_KEY", ""),
	)

	s3Storage := storage.NewS3Storage(
		s3Client,
		getEnv("S3_BUCKET", "interview-files"),
		getEnv("S3_REGION", "ru-1"),
	)
	log.Println("S3 storage initialized")

	// Создаем repositories
	log.Println("Initializing repositories...")
	vacancyRepo := repository.NewVacancyRepository(database)
	resumeRepo := repository.NewResumeRepository(database)
	interviewRepo := repository.NewInterviewRepository(database)
	log.Println("Repositories initialized")

	// Создаем services
	log.Println("Initializing services...")
	vacancySvc := service.NewVacancyService(vacancyRepo, s3Storage)
	resumeSvc := service.NewResumeService(resumeRepo, s3Storage)
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
	port := getEnv("PORT", "8080")
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
		log.Printf("Server starting on port %s", port)
		log.Printf("Environment: %s", getEnv("GIN_MODE", "debug"))

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
