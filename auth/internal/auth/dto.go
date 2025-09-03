// dto.go
package auth

import (
	"ai-hr-service/internal/models"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"regexp"
	"strings"
	"unicode"
)

var validate *validator.Validate

// Инициализируем валидатор один раз
func init() {
	validate = validator.New()
	validate.RegisterValidation("name", validateName)       // Новая валидация для имени
	validate.RegisterValidation("surname", validateSurname) // Новая валидация для фамилии
	// Обязательно регистрируем все кастомные валидаторы
	validate.RegisterValidation("ascii_email", validateASCIIEmail)
	validate.RegisterValidation("strong_password", validateStrongPassword)
	validate.RegisterValidation("alpha_unicode", validateAlphaUnicode)
	validate.RegisterValidation("jwt_token", validateJWTToken)

}
func validateName(fl validator.FieldLevel) bool {
	value := fl.Field().String()

	if len(value) < 2 || len(value) > 50 {
		return false
	}

	// Разрешаем: буквы (включая Unicode), дефисы, апострофы, пробелы
	nameRegex := regexp.MustCompile(`^[\p{L}'-][\p{L} ' -]*[\p{L}'-]?$`)
	if !nameRegex.MatchString(value) {
		return false
	}

	// Запрещаем множественные пробелы подряд
	if strings.Contains(value, "  ") {
		return false
	}

	// Запрещаем начинаться или заканчиваться пробелом, дефисом или апострофом
	if strings.HasPrefix(value, " ") || strings.HasPrefix(value, "-") || strings.HasPrefix(value, "'") ||
		strings.HasSuffix(value, " ") || strings.HasSuffix(value, "-") || strings.HasSuffix(value, "'") {
		return false
	}

	return true
}

func validateSurname(fl validator.FieldLevel) bool {
	value := fl.Field().String()

	if len(value) < 2 || len(value) > 100 {
		return false
	}

	// Разрешаем: буквы (включая Unicode), дефисы
	surnameRegex := regexp.MustCompile(`^[\p{L}-]+$`)
	if !surnameRegex.MatchString(value) {
		return false
	}

	// Запрещаем начинаться или заканчиваться дефисом
	if strings.HasPrefix(value, "-") || strings.HasSuffix(value, "-") {
		return false
	}

	// Запрещаем множественные дефисы подряд
	if strings.Contains(value, "--") {
		return false
	}

	return true
}

func validateJWTToken(fl validator.FieldLevel) bool {
	tokenString := fl.Field().String()

	tokenString = strings.Trim(tokenString, `"`)

	// Парсим без проверки подписи, только структуру
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	_, _, err := parser.ParseUnverified(tokenString, &jwt.RegisteredClaims{})
	return err == nil

	/*

		token := fl.Field().String()
		if token == "invalid.jwt.token" {
			return false
		}
		token = strings.Trim(token, `"`)
		if len(token) < 10 {
			return false
		}

		parts := strings.Split(token, ".")
		if len(parts) != 3 {
			return false
		}
		if len(parts[0]) == 0 || len(parts[1]) == 0 || len(parts[2]) == 0 {
			return false
		}

		for _, part := range parts {
			if len(part) == 0 {
				return false
			}
			// Проверяем каждый символ на соответствие base64url
			for _, r := range part {
				if !((r >= 'A' && r <= 'Z') ||
					(r >= 'a' && r <= 'z') ||
					(r >= '0' && r <= '9') ||
					r == '-' || r == '_') {
					return false
				}
			}
		}

		return true*/
}

// Исправленные кастомные валидаторы
func validateASCIIEmail(fl validator.FieldLevel) bool {
	email := fl.Field().String()

	// Проверяем что все символы ASCII
	for _, r := range email {
		if r > 127 { // unicode.MaxASCII = 127
			return false
		}
	}

	// Проверяем формат email
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

func validateStrongPassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	// Минимум 6 символов (как в вашем тесте)
	if len(password) < 6 {
		return false
	}

	hasDigit := false
	hasLetter := false

	for _, r := range password {
		if unicode.IsDigit(r) {
			hasDigit = true
		}
		if unicode.IsLetter(r) {
			hasLetter = true
		}
	}

	return hasDigit && hasLetter
}

func validateAlphaUnicode(fl validator.FieldLevel) bool {
	value := fl.Field().String()

	if len(value) == 0 {
		return false
	}

	// Разрешаем только буквы (включая Unicode)
	for _, r := range value {
		if !unicode.IsLetter(r) {
			return false
		}
	}

	return true
}

// Методы валидации для всех структур
// Структуры для запросов
type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50,name"`
	Surname  string `json:"surname" validate:"required,surname"`
	Email    string `json:"email" validate:"required,ascii_email"`
	Password string `json:"password" validate:"required,min=6,strong_password"`
}

type UpdateProfileRequest struct {
	Username string `json:"username" validate:"omitempty,min=3,max=50,name"`
	Surname  string `json:"surname" validate:"omitempty,surname"`
	Email    string `json:"email" validate:"omitempty,ascii_email"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,ascii_email"`
	Password string `json:"password" validate:"required,min=6"`
}

// Методы валидации
func (r *RegisterRequest) Validate() error {
	return validate.Struct(r)
}

func (r *LoginRequest) Validate() error {
	return validate.Struct(r)
}

func (r *RefreshTokenRequest) Validate() error {
	return validate.Struct(r)
}

func (r *LogoutRequest) Validate() error {
	return validate.Struct(r)
}

func (r *ChangePasswordRequest) Validate() error {
	return validate.Struct(r)
}

func (r *UpdateProfileRequest) Validate() error {
	return validate.Struct(r)
}

// Структуры для запросов

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required,jwt_token"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required,jwt_token"`
}

// Дополнительные структуры запросов
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required,min=6"`
	NewPassword     string `json:"new_password" validate:"required,strong_password"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=NewPassword"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,ascii_email"`
}

type ResetPasswordRequest struct {
	Token           string `json:"token" validate:"required,min=10"`
	NewPassword     string `json:"new_password" validate:"required,strong_password"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=NewPassword"`
}

// Структуры для ответов
type LoginResponse struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	TokenType    string      `json:"token_type"`
	ExpiresIn    int         `json:"expires_in"` // в секундах
	User         models.User `json:"user"`
}

type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

type ProfileResponse struct {
	User models.User `json:"user"`
}

type ErrorResponse struct {
	Error   string            `json:"error"`
	Details map[string]string `json:"details,omitempty"`
}

type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Структура для детальных ошибок валидации
type ValidationError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Value   string `json:"value"`
	Message string `json:"message"`
}

type ValidationErrorResponse struct {
	Error  string            `json:"error"`
	Fields []ValidationError `json:"fields"`
}
