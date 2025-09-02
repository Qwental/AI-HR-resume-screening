package repository

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"
	"interview/internal/models"
)

type InterviewRepository interface {
	Create(ctx context.Context, interview *models.Interview) error
	GetByID(ctx context.Context, id string) (*models.Interview, error)
	GetByToken(ctx context.Context, token string) (*models.Interview, error)
	GetByVacancyID(ctx context.Context, vacancyID string) ([]*models.Interview, error)
	Update(ctx context.Context, interview *models.Interview) error
	Delete(ctx context.Context, id string) error
	DeleteInterview(ctx context.Context, interview *models.Interview) error
	ListByVacancy(ctx context.Context, vacancyID string) ([]models.Interview, error) // Для обратной совместимости
}

type interviewRepository struct {
	db *gorm.DB
}

func NewInterviewRepository(db *gorm.DB) InterviewRepository {
	return &interviewRepository{db: db}
}

func (r *interviewRepository) Create(ctx context.Context, interview *models.Interview) error {
	if interview == nil {
		return errors.New("interview cannot be nil")
	}
	return r.db.WithContext(ctx).Create(interview).Error
}

func (r *interviewRepository) GetByID(ctx context.Context, id string) (*models.Interview, error) {
	if id == "" {
		return nil, errors.New("id cannot be empty")
	}

	var interview models.Interview
	if err := r.db.WithContext(ctx).First(&interview, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("interview not found")
		}
		return nil, err
	}
	return &interview, nil
}

func (r *interviewRepository) GetByToken(ctx context.Context, token string) (*models.Interview, error) {
	if token == "" {
		return nil, errors.New("token cannot be empty")
	}

	var interview models.Interview

	query := r.db.WithContext(ctx)

	if strings.Contains(token, "/interview/") {
		err := query.Where("url_token = ?", token).First(&interview).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("interview not found")
			}
			return nil, err
		}
	} else {
		err := query.Where("url_token LIKE ?", "%/interview/"+token).First(&interview).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("interview not found")
			}
			return nil, err
		}
	}

	return &interview, nil
}

func (r *interviewRepository) GetByVacancyID(ctx context.Context, vacancyID string) ([]*models.Interview, error) {
	if vacancyID == "" {
		return nil, errors.New("vacancyID cannot be empty")
	}

	var interviews []*models.Interview
	if err := r.db.WithContext(ctx).Where("vacancy_id = ?", vacancyID).Find(&interviews).Error; err != nil {
		return nil, err
	}
	return interviews, nil
}

func (r *interviewRepository) Update(ctx context.Context, interview *models.Interview) error {
	if interview == nil {
		return errors.New("interview cannot be nil")
	}
	if interview.ID == "" {
		return errors.New("interview ID cannot be empty")
	}

	return r.db.WithContext(ctx).Model(interview).Select("*").Updates(interview).Error
}

func (r *interviewRepository) Delete(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("id cannot be empty")
	}

	result := r.db.WithContext(ctx).Delete(&models.Interview{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("interview not found")
	}

	return nil
}

func (r *interviewRepository) DeleteInterview(ctx context.Context, interview *models.Interview) error {
	if interview == nil {
		return errors.New("interview cannot be nil")
	}
	return r.db.WithContext(ctx).Delete(interview).Error
}

func (r *interviewRepository) ListByVacancy(ctx context.Context, vacancyID string) ([]models.Interview, error) {
	if vacancyID == "" {
		return nil, errors.New("vacancyID cannot be empty")
	}

	var interviews []models.Interview
	if err := r.db.WithContext(ctx).Where("vacancy_id = ?", vacancyID).Find(&interviews).Error; err != nil {
		return nil, err
	}
	return interviews, nil
}

func (r *interviewRepository) GetActiveInterviews(ctx context.Context) ([]*models.Interview, error) {
	var interviews []*models.Interview
	if err := r.db.WithContext(ctx).
		Where("status IN ?", []string{"pending", "started"}).
		Where("date_start <= NOW()").
		Where("updated_at >= NOW()").
		Find(&interviews).Error; err != nil {
		return nil, err
	}
	return interviews, nil
}

func (r *interviewRepository) GetExpiredInterviews(ctx context.Context) ([]*models.Interview, error) {
	var interviews []*models.Interview
	if err := r.db.WithContext(ctx).
		Where("status IN ?", []string{"pending", "started"}).
		Where("updated_at < NOW()").
		Find(&interviews).Error; err != nil {
		return nil, err
	}
	return interviews, nil
}

func (r *interviewRepository) CountByStatus(ctx context.Context, status string) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Interview{}).
		Where("status = ?", status).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *interviewRepository) GetInterviewsByDateRange(ctx context.Context, startDate, endDate string) ([]*models.Interview, error) {
	var interviews []*models.Interview
	if err := r.db.WithContext(ctx).
		Where("date_start BETWEEN ? AND ?", startDate, endDate).
		Find(&interviews).Error; err != nil {
		return nil, err
	}
	return interviews, nil
}

func (r *interviewRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	if id == "" {
		return errors.New("id cannot be empty")
	}
	if status == "" {
		return errors.New("status cannot be empty")
	}

	result := r.db.WithContext(ctx).Model(&models.Interview{}).
		Where("id = ?", id).
		Update("status", status)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("interview not found")
	}

	return nil
}
