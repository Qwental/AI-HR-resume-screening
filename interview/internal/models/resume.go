package models

import (
	"gorm.io/datatypes"
	"time"
)

type Resume struct {
	ID        string         `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	VacancyID string         `gorm:"type:uuid;index" json:"vacancy_id"`
	FileURL   string         `json:"file_url,omitempty"`
	Text      datatypes.JSON `gorm:"type:jsonb" json:"text_jsonb,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	Result    datatypes.JSON `gorm:"type:jsonb" json:"result_jsonb,omitempty"`
}
