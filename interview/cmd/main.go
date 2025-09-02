package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"interview/internal/db"
	"interview/internal/handlers"
	"interview/internal/repository"
	"interview/internal/service"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env not found, using system env")
	}

	fmt.Println("k")
	db.RunMigrations()
	fmt.Println("k")

	database, err := db.NewDB()
	if err != nil {
		log.Fatalf("db connection failed: %v", err)
	}

	interviewRepo := repository.NewInterviewRepository(database)
	interviewSvc := service.NewInterviewService(interviewRepo)

	router := handlers.SetupRouter(interviewSvc)
	gin.SetMode(gin.ReleaseMode)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Println("k")

	log.Printf("server listening on :%s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
