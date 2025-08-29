// DTO (Data Transfer Objects)

package auth

import "ai-hr-service/internal/models"

// Структуры для запросов
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Surname  string `json:"surname" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// Структуры для ответов
type LoginResponse struct {
	Token string      `json:"token"`
	User  models.User `json:"user"`
}

type ProfileResponse struct {
	User models.User `json:"user"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
