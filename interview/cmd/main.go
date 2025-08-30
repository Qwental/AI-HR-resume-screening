package main

import (
	_ "database/sql"
	"log"

	"github.com/joho/godotenv"
	"interview/internal/db"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env not found")
	}
	db.RunMigrations()

	_, err := db.NewDB()
	if err != nil {
		log.Fatalf("DB connection failed: %v", err)
	}
}
