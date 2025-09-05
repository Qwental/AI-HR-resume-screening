package db

////
////import (
////	"fmt"
////	"log"
////	"os"
////	"path/filepath"
////
////	"github.com/golang-migrate/migrate/v4"
////	_ "github.com/golang-migrate/migrate/v4/database/postgres"
////	_ "github.com/golang-migrate/migrate/v4/source/file"
////)
////
////func RunMigrations() {
////	dsn := os.Getenv("POSTGRES_DSN")
////	if dsn == "" {
////		log.Fatal("POSTGRES_DSN is not set")
////	}
////
////	// Путь до папки migrations
////	absPath, err := filepath.Abs("../migrations")
////	if err != nil {
////		log.Fatalf("cannot resolve migrations path: %v", err)
////	}
////
////	m, err := migrate.New(
////		fmt.Sprintf("file://%s", absPath),
////		dsn,
////	)
////	if err != nil {
////		log.Fatalf("migration init failed: %v", err)
////	}
////
////	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
////		log.Fatalf("migration failed: %v", err)
////	}
////
////	log.Println("✅ Migrations applied")
////}
//
//package db
//
//import (
//	_ "github.com/golang-migrate/migrate/v4/database/postgres"
//	_ "github.com/golang-migrate/migrate/v4/source/file"
//)
//
////func RunMigrations() {
////	// Формируем DSN для миграций
////	dbHost := getEnv("DB_HOST", "localhost")
////	dbPort := getEnv("DB_PORT", "5432")
////	dbUser := getEnv("DB_USER", "ai_hr_user")
////	dbPassword := os.Getenv("DB_PASSWORD")
////	if dbPassword == "" {
////		log.Fatal("DB_PASSWORD environment variable is required")
////	}
////	dbName := getEnv("DB_NAME", "ai_hr_db")
////	sslMode := getEnv("DB_SSLMODE", "disable")
////
////	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
////		dbUser, dbPassword, dbHost, dbPort, dbName, sslMode)
////
////	// Путь до папки migrations в Docker контейнере
////	migrationsPath := "./migrations"
////	absPath, err := filepath.Abs(migrationsPath)
////	if err != nil {
////		log.Printf("Warning: cannot resolve migrations path, using relative: %v", err)
////		absPath = migrationsPath
////	}
////
////	log.Printf("Running migrations from: %s", absPath)
////	log.Printf("Database DSN: postgres://%s:***@%s:%s/%s?sslmode=%s",
////		dbUser, dbHost, dbPort, dbName, sslMode)
////
////	m, err := migrate.New(
////		fmt.Sprintf("file://%s", absPath),
////		dsn,
////	)
////	if err != nil {
////		log.Fatalf("migration init failed: %v", err)
////	}
////
////	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
////		log.Fatalf("migration failed: %v", err)
////	}
////
////	log.Println("✅ Migrations applied successfully")
////}
