package models

import (
	"time"

	"gorm.io/datatypes"
)

type Interview struct {
	ID          string         `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	ResumeID    *string        `gorm:"type:uuid" json:"resume_id,omitempty"`
	VacancyID   string         `gorm:"type:uuid;index" json:"vacancy_id"`
	Status      string         `gorm:"type:text;not null;default:'pending'" json:"status"`
	TextJSONB   datatypes.JSON `gorm:"type:jsonb;column:text_jsonb" json:"text_jsonb,omitempty"`
	URLToken    *string        `gorm:"uniqueIndex" json:"url_token,omitempty"`
	ScoreJSONB  datatypes.JSON `gorm:"type:jsonb;column:score_jsonb" json:"score_jsonb,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	ScheduledAt *time.Time     `json:"scheduled_at,omitempty" gorm:"column:date_start"`
	ExpiresAt   *time.Time     `json:"expires_at,omitempty" gorm:"column:updated_at"`
	StartedAt   *time.Time     `json:"started_at,omitempty"`
}
