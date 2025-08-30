package repository

import (
	"context"
	"gorm.io/gorm"
	"interview/internal/models"
)

type InterviewRepository interface {
	Create(ctx context.Context, interview *models.Interview) error
	GetByID(ctx context.Context, id string) (*models.Interview, error)
	Update(ctx context.Context, interview *models.Interview) error
	Delete(ctx context.Context, interview *models.Interview) error
	ListByVacancy(ctx context.Context, vacancyID string) ([]models.Interview, error)
}

type interviewRepository struct {
	db *gorm.DB
}

func NewInterviewRepository(db *gorm.DB) InterviewRepository {
	return &interviewRepository{db: db}
}

func (r *interviewRepository) Create(ctx context.Context, interview *models.Interview) error {
	return r.db.WithContext(ctx).Create(interview).Error
}

func (r *interviewRepository) GetByID(ctx context.Context, id string) (*models.Interview, error) {
	var interview models.Interview
	if err := r.db.WithContext(ctx).First(&interview, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &interview, nil
}

func (r *interviewRepository) Update(ctx context.Context, interview *models.Interview) error {
	return r.db.WithContext(ctx).Save(interview).Error
}

func (r *interviewRepository) Delete(ctx context.Context, interview *models.Interview) error {
	return r.db.WithContext(ctx).Delete(interview).Error
}

func (r *interviewRepository) ListByVacancy(ctx context.Context, id string) ([]models.Interview, error) {
	var interviews []models.Interview
	if err := r.db.WithContext(ctx).Where("vacancy_id = ?", id).Find(&interviews).Error; err != nil {
		return nil, err
	}
	return interviews, nil
}
