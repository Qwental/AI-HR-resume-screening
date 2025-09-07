package models

import (
	"time"
)

// User - обновленная модель пользователя под новую схему БД
type User struct {
	ID           uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	Username     string    `json:"username" gorm:"type:varchar(100);unique;not null"`
	Surname      string    `json:"surname" gorm:"type:varchar(100);not null"`
	Email        string    `json:"email" gorm:"type:varchar(255);unique;not null"`
	PasswordHash string    `json:"-" gorm:"column:password_hash;type:varchar(255);not null"`                                                           // изменено на password_hash
	Role         string    `json:"role" gorm:"type:varchar(255);not null;check:role IN ('hr_specialist','candidate','admin');default:'hr_specialist'"` // обновленные роли
	IsActive     bool      `json:"is_active" gorm:"default:true"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime" db:"updated_at"`
}

func (User) TableName() string {
	return "users"
}

// Token - модель для хранения токенов в БД
type Token struct {
	ID        uint       `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID    uint       `json:"user_id" gorm:"not null;index"`
	TokenHash string     `json:"-" gorm:"type:varchar(255);unique;not null"`
	ExpiresAt time.Time  `json:"expires_at" gorm:"type:timestamptz;not null"`
	IsRevoked bool       `json:"is_revoked" gorm:"default:false"`
	CreatedAt time.Time  `json:"created_at" gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP"`
	RevokedAt *time.Time `json:"revoked_at,omitempty" gorm:"type:timestamptz"`

	// Связь с пользователем
	User User `json:"user,omitempty" gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE"`
}

func (Token) TableName() string {
	return "tokens"
}
