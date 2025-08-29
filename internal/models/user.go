package models

import (
	"time"
)

// User - модель пользователя для GORM

type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Username  string    `json:"username" gorm:"unique;not null"`
	Surname   string    `json:"surname" gorm:"not null"`
	Email     string    `json:"email" gorm:"unique;not null"`
	Password  string    `json:"-" gorm:"not null"` // скрывается в JSON
	Role      string    `json:"role" gorm:"default:'hr';not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName задает имя таблицы для GORM
// По дефолту GORM использует множественное число от названия структуры
// Но лучше явно указать чтобы потом не было сюрпризов
func (User) TableName() string {
	return "users" // Таблица будет называться "users"
}
