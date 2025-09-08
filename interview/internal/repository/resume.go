package repository

import (
	"context"
	"gorm.io/gorm"
	"interview/internal/models"
)

type ResumeRepository interface {
	Create(ctx context.Context, resume *models.Resume) error
	GetByID(ctx context.Context, id string) (*models.Resume, error)
	GetByVacancy(ctx context.Context, id string) ([]*models.Resume, error)
	Delete(ctx context.Context, id string) error
	UpdateStatus(ctx context.Context, id, status string) error
	UpdateStatusAndResult(ctx context.Context, id, status string, result map[string]interface{}) error
	UpdateResult(ctx context.Context, id string, result map[string]interface{}) error
	Update(ctx context.Context, resume *models.Resume) error
	UpdateText(ctx context.Context, id, text string) error // ← добавлено

}

type resumeRepository struct {
	db *gorm.DB
}

func NewResumeRepository(db *gorm.DB) ResumeRepository {
	return &resumeRepository{db: db}
}

func (r *resumeRepository) Create(ctx context.Context, resume *models.Resume) error {
	return r.db.WithContext(ctx).Create(resume).Error
}

func (r *resumeRepository) GetByID(ctx context.Context, id string) (*models.Resume, error) {
	var resume models.Resume
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&resume).Error
	if err != nil {
		return nil, err
	}
	return &resume, nil
}

func (r *resumeRepository) GetByVacancy(ctx context.Context, id string) ([]*models.Resume, error) {
	var resumes []*models.Resume
	err := r.db.WithContext(ctx).Where("vacancy_id = ?", id).Find(&resumes).Error
	if err != nil {
		return nil, err
	}
	return resumes, nil
}

func (r *resumeRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.Resume{}, "id = ?", id).Error
}

func (r *resumeRepository) UpdateStatus(ctx context.Context, id, status string) error {
	return r.db.WithContext(ctx).
		Model(&models.Resume{}).
		Where("id = ?", id).
		Update("status", status).Error
}

func (r *resumeRepository) UpdateStatusAndResult(ctx context.Context, id, status string, result map[string]interface{}) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if result != nil {
		updates["result"] = result
	}

	return r.db.WithContext(ctx).
		Model(&models.Resume{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (r *resumeRepository) Update(ctx context.Context, resume *models.Resume) error {
	return r.db.WithContext(ctx).Save(resume).Error
}

func (r *resumeRepository) UpdateResult(ctx context.Context, id string, result map[string]interface{}) error {
	updates := map[string]interface{}{}
	if result != nil {
		updates["result"] = result
	}

	return r.db.WithContext(ctx).
		Model(&models.Resume{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (r *resumeRepository) UpdateText(ctx context.Context, id, text string) error {
	return r.db.WithContext(ctx).
		Model(&models.Resume{}).
		Where("id = ?", id).
		Update("text", text).Error
}
