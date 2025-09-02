package db

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"os"

	_ "github.com/lib/pq"
)

func NewDB() (*gorm.DB, error) {
	dsn := os.Getenv("POSTGRES_DSN")
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}
