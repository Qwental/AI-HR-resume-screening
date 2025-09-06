package database

import (
	"ai-hr-service/internal/models"
	"fmt"
	"gorm.io/gorm"
)

func RunMigrations(db *gorm.DB) error {
	// Автоматические миграции для всех моделей auth сервиса
	err := db.AutoMigrate(
		&models.User{},
		&models.Token{},
	)

	if err != nil {
		return fmt.Errorf("could not run GORM migrations: %w", err)
	}

	return nil
}
