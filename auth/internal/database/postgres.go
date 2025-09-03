package database

import (
	"ai-hr-service/internal/config"
	"ai-hr-service/internal/models"
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect(cfg config.DatabaseConfig) *gorm.DB {
	// Ждем, пока PostgreSQL будет готов
	log.Println("Waiting for PostgreSQL to be ready...")
	if err := waitForDB(cfg); err != nil {
		log.Fatal("Failed to wait for database:", err)
	}

	// Создаем базу данных если она не существует
	if err := createDatabaseIfNotExists(cfg); err != nil {
		log.Fatal("Failed to create database:", err)
	}

	// Теперь подключаемся к нашей БД
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		cfg.Host, cfg.User, cfg.Password, cfg.DBName, cfg.Port, cfg.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Автомиграция для всех моделей
	if err := db.AutoMigrate(&models.User{}, &models.Token{}); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	log.Println("Database connected successfully")
	return db
}

// waitForDB ждет, пока PostgreSQL будет готов принимать соединения
func waitForDB(cfg config.DatabaseConfig) error {
	// Подключаемся к системной БД postgres для проверки готовности
	systemDSN := fmt.Sprintf("host=%s user=%s password=%s dbname=postgres port=%s sslmode=%s",
		cfg.Host, cfg.User, cfg.Password, cfg.Port, cfg.SSLMode)

	for i := 0; i < 30; i++ { // Ждем до 30 секунд
		db, err := gorm.Open(postgres.Open(systemDSN), &gorm.Config{})
		if err == nil {
			sqlDB, _ := db.DB()
			if err := sqlDB.Ping(); err == nil {
				sqlDB.Close()
				log.Println("PostgreSQL is ready!")
				return nil
			}
			sqlDB.Close()
		}
		log.Printf("PostgreSQL not ready yet, retrying in 1 second... (attempt %d/30)", i+1)
		time.Sleep(1 * time.Second)
	}

	return fmt.Errorf("PostgreSQL did not become ready within 30 seconds")
}

// createDatabaseIfNotExists создает базу данных если она не существует
func createDatabaseIfNotExists(cfg config.DatabaseConfig) error {
	// Подключаемся к системной БД postgres
	systemDSN := fmt.Sprintf("host=%s user=%s password=%s dbname=postgres port=%s sslmode=%s",
		cfg.Host, cfg.User, cfg.Password, cfg.Port, cfg.SSLMode)

	db, err := gorm.Open(postgres.Open(systemDSN), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to postgres database: %v", err)
	}

	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	// Проверяем, существует ли уже наша БД
	var count int64
	checkQuery := "SELECT COUNT(*) FROM pg_database WHERE datname = ?"
	err = db.Raw(checkQuery, cfg.DBName).Scan(&count).Error
	if err != nil {
		return fmt.Errorf("failed to check if database exists: %v", err)
	}

	if count == 0 {
		// Создаем базу данных
		createQuery := fmt.Sprintf("CREATE DATABASE %s", cfg.DBName)
		err = db.Exec(createQuery).Error
		if err != nil {
			return fmt.Errorf("failed to create database: %v", err)
		}
		log.Printf("Database %s created successfully", cfg.DBName)
	} else {
		log.Printf("Database %s already exists", cfg.DBName)
	}

	return nil
}
