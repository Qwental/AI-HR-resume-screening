package db

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrations() {
	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		log.Fatal("POSTGRES_DSN is not set")
	}

	// Путь до папки migrations
	absPath, err := filepath.Abs("../migrations")
	if err != nil {
		log.Fatalf("cannot resolve migrations path: %v", err)
	}

	m, err := migrate.New(
		fmt.Sprintf("file://%s", absPath),
		dsn,
	)
	if err != nil {
		log.Fatalf("migration init failed: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("migration failed: %v", err)
	}

	log.Println("✅ Migrations applied")
}
