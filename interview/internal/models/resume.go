package models

import (
	"gorm.io/datatypes"
	"time"
)

type Resume struct {
	ID        string    `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	VacancyID string    `gorm:"type:uuid;index" json:"vacancy_id"`
	CreatedAt time.Time `json:"created_at"`
	Status    string    `gorm:"type:text;default:'pending'" json:"status"`

	// StorageKey - внутренний ключ
	StorageKey string `json:"-" gorm:"column:storage_key"`

	// FileURL - временная ссылка для скачивания, не сохраняется в БД
	FileURL string `json:"file_url,omitempty" gorm:"-"`

	Mail string `json:"mail"`
	Text string `json:"text"`

	Result datatypes.JSON `gorm:"type:jsonb;column:result_jsonb" json:"result_jsonb,omitempty"`
}
