package repository

import (
	"context"
	"gorm.io/gorm"
	"interview/internal/models"
)

type VacancyRepository interface {
	Create(ctx context.Context, vacancy *models.Vacancy) error
	GetByID(ctx context.Context, id string) (*models.Vacancy, error)
	GetAll(ctx context.Context) ([]*models.Vacancy, error)
	Update(ctx context.Context, vacancy *models.Vacancy) error
	Delete(ctx context.Context, id string) error
}

type vacancyRepository struct {
	db *gorm.DB
}

func NewVacancyRepository(db *gorm.DB) VacancyRepository {
	return &vacancyRepository{db: db}
}

func (r *vacancyRepository) Create(ctx context.Context, vacancy *models.Vacancy) error {
	return r.db.WithContext(ctx).Create(vacancy).Error
}

func (r *vacancyRepository) GetByID(ctx context.Context, id string) (*models.Vacancy, error) {
	var vacancy models.Vacancy
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&vacancy).Error
	if err != nil {
		return nil, err
	}
	return &vacancy, nil
}

func (r *vacancyRepository) GetAll(ctx context.Context) ([]*models.Vacancy, error) {
	var vacancies []*models.Vacancy
	err := r.db.WithContext(ctx).Find(&vacancies).Error
	return vacancies, err
}

func (r *vacancyRepository) Update(ctx context.Context, vacancy *models.Vacancy) error {
	return r.db.WithContext(ctx).Save(vacancy).Error
}

func (r *vacancyRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.Vacancy{}, "id = ?", id).Error
}
