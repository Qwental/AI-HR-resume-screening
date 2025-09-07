package utils

import "github.com/gin-gonic/gin"

// ErrorResponse отправляет стандартный ответ с ошибкой
func ErrorResponse(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, gin.H{
		"error":   message,
		"success": false,
	})
}

// SuccessResponse отправляет стандартный успешный ответ
func SuccessResponse(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, gin.H{
		"data":    data,
		"success": true,
	})
}

// MessageResponse отправляет ответ с сообщением
func MessageResponse(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, gin.H{
		"message": message,
		"success": true,
	})
}
