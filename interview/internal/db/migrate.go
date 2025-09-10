package db

import (
	"interview/internal/models"

	"fmt"
	"gorm.io/gorm"
)

func RunMigrations(db *gorm.DB) error {
	// Автоматические миграции для всех моделей auth сервиса
	err := db.AutoMigrate(
		&models.Vacancy{},
		&models.Resume{},
		&models.Interview{},
	)

	if err != nil {
		return fmt.Errorf("could not run GORM migrations: %w", err)
	}

	return nil
}
