package models

import (
	"gorm.io/datatypes"
	"time"
)

type Resume struct {
	ID        string    `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	VacancyID string    `gorm:"type:uuid;not null;index" json:"vacancy_id"`
	CreatedAt time.Time `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"created_at"`

	// Основные поля из БД
	FileURL    string `gorm:"type:text;column:file_url" json:"file_url,omitempty"`
	StorageKey string `gorm:"type:varchar(255);column:storage_key" json:"-"`
	Text       string `gorm:"type:text" json:"text,omitempty"`
	Status     string `gorm:"type:text;default:'pending'" json:"status"`
	Mail       string `gorm:"type:text" json:"mail,omitempty"`

	// JSONB поля для анализа
	ResultJSONB         datatypes.JSON `gorm:"type:jsonb;column:result_jsonb" json:"result_jsonb,omitempty"`
	ResumeAnalysisJSONB datatypes.JSON `gorm:"type:jsonb;column:resume_analysis_jsonb" json:"resume_analysis_jsonb,omitempty"`
}

// TableName указывает имя таблицы для GORM
func (Resume) TableName() string {
	return "resumes"
}
