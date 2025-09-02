package database

import (
	"ai-hr-service/internal/config"
	"ai-hr-service/internal/models"
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect(cfg config.DatabaseConfig) *gorm.DB {
	// Сначала подключаемся к postgres (системной БД) чтобы создать нашу БД
	systemDSN := fmt.Sprintf("host=%s user=%s password=%s dbname=postgres port=%s sslmode=%s",
		cfg.Host, cfg.User, cfg.Password, cfg.Port, cfg.SSLMode)

	systemDB, err := gorm.Open(postgres.Open(systemDSN), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to system database:", err)
	}

	// Создаем нашу базу данных если она не существует
	createDBSQL := fmt.Sprintf("CREATE DATABASE %s", cfg.DBName)
	result := systemDB.Exec(createDBSQL)
	if result.Error != nil {
		// Игнорируем ошибку если БД уже существует
		log.Printf("Database creation result (ignore if exists): %v", result.Error)
	} else {
		log.Printf("Database %s created successfully", cfg.DBName)
	}

	// Закрываем соединение с системной БД
	sqlDB, _ := systemDB.DB()
	sqlDB.Close()

	// Теперь подключаемся к нашей БД
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		cfg.Host, cfg.User, cfg.Password, cfg.DBName, cfg.Port, cfg.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Автомиграция почему-то не всегда работает ((((((
	if err := db.AutoMigrate(&models.User{}); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	log.Println("Database connected successfully")
	return db
}
