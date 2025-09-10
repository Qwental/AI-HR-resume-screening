package models

import (
	"gorm.io/datatypes"
	"time"
)

type Vacancy struct {
	ID          string     `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	Title       string     `json:"title"`
	Description *string    `json:"description,omitempty"`
	UsersID     string     `gorm:"type:uuid;index" json:"users_id"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`

	// StorageKey - внутренний ключ
	StorageKey string `json:"-" gorm:"column:storage_key"`

	// FileURL - временная ссылка для скачивания, не сохраняется в БД
	FileURL string `json:"file_url,omitempty" gorm:"-"`

	WeightSoft int            `gorm:"type:int;default:33;check:weight_soft>=0 AND weight_soft<=100" json:"weight_soft"`
	WeightHard int            `gorm:"type:int;default:33;check:weight_hard>=0 AND weight_hard<=100" json:"weight_hard"`
	WeightCase int            `gorm:"type:int;default:34;check:weight_case>=0 AND weight_case<=100" json:"weight_case"`
	TextJSONB  datatypes.JSON `gorm:"type:jsonb;column:text_jsonb" json:"text_jsonb,omitempty"`
}
