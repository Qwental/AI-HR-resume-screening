package models

import (
	"time"

	"gorm.io/datatypes"
)

type Interview struct {
	ID         string         `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	ResumeID   *string        `gorm:"type:uuid" json:"resume_id,omitempty"`
	VacancyID  string         `gorm:"type:uuid;index" json:"vacancy_id"`
	Status     string         `gorm:"type:text;not null;default:'pending'" json:"status"`
	TextJSONB  datatypes.JSON `gorm:"type:jsonb" json:"text_jsonb,omitempty"`
	AudioURL   *string        `json:"audio_url,omitempty"`
	URLToken   *string        `gorm:"uniqueIndex" json:"url_token,omitempty"`
	ScoreJSONB datatypes.JSON `gorm:"type:jsonb" json:"score_jsonb,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
	DateStart  *time.Time     `json:"date_start,omitempty"`
	UpdatedAt  *time.Time     `json:"updated_at,omitempty"`
}
