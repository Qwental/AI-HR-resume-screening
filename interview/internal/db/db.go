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

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
