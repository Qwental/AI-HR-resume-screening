package auth

import (
	"ai-hr-service/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// Register регистрирует нового пользователя
func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	response, err := h.service.Register(req)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "user with this email already exists" ||
			err.Error() == "user with this username already exists" {
			status = http.StatusConflict
		}
		utils.ErrorResponse(c, status, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, response)
}

// Login авторизует пользователя
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	response, err := h.service.Login(req)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "invalid credentials" || err.Error() == "account is deactivated" {
			status = http.StatusUnauthorized
		}
		utils.ErrorResponse(c, status, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, response)
}

// RefreshToken обновляет токены
func (h *Handler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	response, err := h.service.RefreshTokens(req)
	if err != nil {
		status := http.StatusUnauthorized
		if err.Error() == "invalid or expired refresh token" {
			status = http.StatusUnauthorized
		}
		utils.ErrorResponse(c, status, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, response)
}

// Logout выходит из системы (отзывает refresh token)
func (h *Handler) Logout(c *gin.Context) {
	var req LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.service.Logout(req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.MessageResponse(c, http.StatusOK, "Successfully logged out")
}

// LogoutAll выходит из всех устройств (отзывает все refresh токены пользователя)
func (h *Handler) LogoutAll(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "user ID not found")
		return
	}

	if err := h.service.LogoutAll(userID.(uint)); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.MessageResponse(c, http.StatusOK, "Successfully logged out from all devices")
}

// GetProfile возвращает профиль текущего пользователя
func (h *Handler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "user ID not found")
		return
	}

	user, err := h.service.GetProfile(userID.(uint))
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "user not found" {
			status = http.StatusNotFound
		}
		utils.ErrorResponse(c, status, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, ProfileResponse{User: *user})
}
