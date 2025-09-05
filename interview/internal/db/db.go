package db

import (
	"fmt"
	"interview/internal/config"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

//func NewDB() (*gorm.DB, error) {
//	dsn := os.Getenv("DATABASE_URL")
//	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
//}

func Connect(cfg config.DatabaseConfig) *gorm.DB {
	// Формируем DSN (Data Source Name) для pgx
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		cfg.Host,
		cfg.User,
		cfg.Password,
		cfg.DBName,
		cfg.Port,
	)

	// Подключаемся с помощью нового драйвера
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Println("Database connection successful")

	// Выполняем автомиграцию
	//err = db.AutoMigrate(&models.User{}, &models.Token{}) // <-- Убедитесь, что все ваши модели здесь
	//if err != nil {
	//	log.Fatalf("Failed to migrate database: %v", err)
	//}

	log.Println("Database migrated successfully")

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

	//var count int64
	//checkQuery := "SELECT COUNT(*) FROM pg_database WHERE datname = $1"
	//err = db.Raw(checkQuery, cfg.DBName).Scan(&count).Error

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

func NewDB() (*gorm.DB, error) {
	// Формируем DSN из переменных окружения для Docker
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "ai_hr_user")
	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		return nil, fmt.Errorf("DB_PASSWORD environment variable is required")
	}
	dbName := getEnv("DB_NAME", "ai_hr_db")
	sslMode := getEnv("DB_SSLMODE", "disable")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		dbHost, dbUser, dbPassword, dbName, dbPort, sslMode)

	log.Printf("Connecting to database: host=%s, port=%s, dbname=%s, user=%s",
		dbHost, dbPort, dbName, dbUser)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Настройка пула соединений
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
