package auth

import (
	"encoding/json"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

// dto_test.go
func TestRegisterRequestValidation(t *testing.T) {
	tests := []struct {
		name          string
		req           RegisterRequest
		expectedError bool
	}{
		{
			name: "валидные данные",
			req: RegisterRequest{
				Username: "Владимир",
				Surname:  "Петров",
				Email:    "valid@example.com",
				Password: "validpassword123",
			},
			expectedError: false,
		},
		{
			name: "валидный username с пробелами",
			req: RegisterRequest{
				Username: "Амелия Добронравовна",
				Surname:  "ValidSurname",
				Email:    "valid@example.com",
				Password: "validpassword123",
			},
			expectedError: false,
		},
		{
			name: "пустой username",
			req: RegisterRequest{
				Username: "",
				Surname:  "ValidSurname",
				Email:    "valid@example.com",
				Password: "validpassword123",
			},
			expectedError: true,
		},
		{
			name: "короткий username",
			req: RegisterRequest{
				Username: "ab",
				Surname:  "ValidSurname",
				Email:    "valid@example.com",
				Password: "validpassword123",
			},
			expectedError: true,
		},
		{
			name: "пустая фамилия",
			req: RegisterRequest{
				Username: "validuser",
				Surname:  "",
				Email:    "valid@example.com",
				Password: "validpassword123",
			},
			expectedError: true,
		},
		{
			name: "фамилия с цифрами",
			req: RegisterRequest{
				Username: "validuser",
				Surname:  "Invalid123",
				Email:    "valid@example.com",
				Password: "validpassword123",
			},
			expectedError: true,
		},
		{
			name: "невалидный email",
			req: RegisterRequest{
				Username: "validuser",
				Surname:  "ValidSurname",
				Email:    "invalid-email",
				Password: "validpassword123",
			},
			expectedError: true,
		},
		{
			name: "email с кириллицей",
			req: RegisterRequest{
				Username: "validuser",
				Surname:  "ValidSurname",
				Email:    "тест@example.com",
				Password: "validpassword123",
			},
			expectedError: true,
		},
		{
			name: "слабый пароль только буквы",
			req: RegisterRequest{
				Username: "validuser",
				Surname:  "ValidSurname",
				Email:    "valid@example.com",
				Password: "password",
			},
			expectedError: true,
		},
		{
			name: "слабый пароль только цифры",
			req: RegisterRequest{
				Username: "validuser",
				Surname:  "ValidSurname",
				Email:    "valid@example.com",
				Password: "123456",
			},
			expectedError: true,
		},
		{
			name: "короткий пароль",
			req: RegisterRequest{
				Username: "validuser",
				Surname:  "ValidSurname",
				Email:    "valid@example.com",
				Password: "ab1",
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()

			// Подробное логирование для отладки
			if err != nil {
				t.Logf("Получена ошибка валидации: %v", err)
				if validationErrors, ok := err.(validator.ValidationErrors); ok {
					for _, ve := range validationErrors {
						t.Logf("  - Поле: %s, Тег: %s, Значение: %v",
							ve.Field(), ve.Tag(), ve.Value())
					}
				}
			} else {
				t.Logf("Валидация прошла успешно")
			}

			if tt.expectedError {
				assert.Error(t, err, "Ожидалась ошибка валидации")
			} else {
				assert.NoError(t, err, "Валидация должна пройти без ошибок")
			}
		})
	}
}

// Дополнительный тест для проверки кастомных валидаторов
func TestCustomValidators(t *testing.T) {
	t.Run("ASCII Email валидатор", func(t *testing.T) {
		err := validate.Var("valid@example.com", "ascii_email")
		assert.NoError(t, err)

		err = validate.Var("тест@example.com", "ascii_email")
		assert.Error(t, err)
	})

	t.Run("Strong Password валидатор", func(t *testing.T) {
		err := validate.Var("validpassword123", "strong_password")
		assert.NoError(t, err)

		err = validate.Var("password", "strong_password") // только буквы
		assert.Error(t, err)

		err = validate.Var("123456", "strong_password") // только цифры
		assert.Error(t, err)
	})

	t.Run("Alpha Unicode валидатор", func(t *testing.T) {
		err := validate.Var("ValidSurname", "alpha_unicode")
		assert.NoError(t, err)

		err = validate.Var("Тестовый", "alpha_unicode")
		assert.NoError(t, err)

		err = validate.Var("Test123", "alpha_unicode") // содержит цифры
		assert.Error(t, err)
	})
}

// TestDTO_LoginRequest_Validation тестирует валидацию LoginRequest
func TestDTO_LoginRequest_Validation(t *testing.T) {
	tests := []struct {
		name          string
		req           LoginRequest
		expectedError bool
	}{
		{
			name: "валидные данные",
			req: LoginRequest{
				Email:    "valid@example.com",
				Password: "validpassword123",
			},
			expectedError: false,
		},
		{
			name: "пустой email",
			req: LoginRequest{
				Email:    "",
				Password: "validpassword123",
			},
			expectedError: true,
		},
		{
			name: "невалидный email",
			req: LoginRequest{
				Email:    "invalid-email",
				Password: "validpassword123",
			},
			expectedError: true,
		},
		{
			name: "email с кириллицей",
			req: LoginRequest{
				Email:    "тест@example.com",
				Password: "validpassword123",
			},
			expectedError: true,
		},
		{
			name: "пустой password",
			req: LoginRequest{
				Email:    "valid@example.com",
				Password: "",
			},
			expectedError: true,
		},
		{
			name: "короткий password",
			req: LoginRequest{
				Email:    "valid@example.com",
				Password: "123",
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Используем метод структуры вместо нового валидатора
			err := tt.req.Validate()

			if tt.expectedError {
				assert.Error(t, err, "Ожидалась ошибка валидации")
				if err != nil {
					t.Logf("Получена ожидаемая ошибка: %v", err)
				}
			} else {
				assert.NoError(t, err, "Валидация должна пройти без ошибок")
				if err != nil {
					t.Logf("Неожиданная ошибка: %v", err)
				}
			}
		})
	}
}

// TestDTO_RefreshTokenRequest_Validation тестирует валидацию RefreshTokenRequest
func TestDTO_RefreshTokenRequest_Validation(t *testing.T) {
	tests := []struct {
		name          string
		req           RefreshTokenRequest
		expectedError bool
	}{
		{
			name: "валидный refresh token",
			req: RefreshTokenRequest{
				RefreshToken: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			},
			expectedError: false,
		},
		{
			name: "пустой refresh token",
			req: RefreshTokenRequest{
				RefreshToken: "",
			},
			expectedError: true,
		},
		{
			name: "невалидный JWT формат",
			req: RefreshTokenRequest{
				RefreshToken: "invalid.jwt.token",
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()

			if tt.expectedError {
				assert.Error(t, err, "Ожидалась ошибка валидации")
			} else {
				assert.NoError(t, err, "Валидация должна пройти без ошибок")
			}
		})
	}
}

// TestDTO_LogoutRequest_Validation тестирует валидацию LogoutRequest
func TestDTO_LogoutRequest_Validation(t *testing.T) {
	tests := []struct {
		name          string
		req           LogoutRequest
		expectedError bool
	}{
		{
			name: "валидный refresh token",
			req: LogoutRequest{
				RefreshToken: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			},
			expectedError: false,
		},
		{
			name: "пустой refresh token",
			req: LogoutRequest{
				RefreshToken: "",
			},
			expectedError: true,
		},
		{
			name: "невалидный JWT формат",
			req: LogoutRequest{
				RefreshToken: "invalid.token.format",
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()

			if tt.expectedError {
				assert.Error(t, err, "Ожидалась ошибка валидации")
			} else {
				assert.NoError(t, err, "Валидация должна пройти без ошибок")
			}
		})
	}
}

// TestDTO_ChangePasswordRequest_Validation тестирует валидацию ChangePasswordRequest
func TestDTO_ChangePasswordRequest_Validation(t *testing.T) {
	tests := []struct {
		name          string
		req           ChangePasswordRequest
		expectedError bool
	}{
		{
			name: "валидные данные",
			req: ChangePasswordRequest{
				CurrentPassword: "currentpass123",
				NewPassword:     "newpassword123",
				ConfirmPassword: "newpassword123",
			},
			expectedError: false,
		},
		{
			name: "пароли не совпадают",
			req: ChangePasswordRequest{
				CurrentPassword: "currentpass123",
				NewPassword:     "newpassword123",
				ConfirmPassword: "differentpass123",
			},
			expectedError: true,
		},
		{
			name: "слабый новый пароль",
			req: ChangePasswordRequest{
				CurrentPassword: "currentpass123",
				NewPassword:     "password",
				ConfirmPassword: "password",
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()

			if tt.expectedError {
				assert.Error(t, err, "Ожидалась ошибка валидации")
			} else {
				assert.NoError(t, err, "Валидация должна пройти без ошибок")
			}
		})
	}
}

// TestDTO_UpdateProfileRequest_Validation тестирует валидацию UpdateProfileRequest
func TestDTO_UpdateProfileRequest_Validation(t *testing.T) {
	tests := []struct {
		name          string
		req           UpdateProfileRequest
		expectedError bool
	}{
		{
			name: "валидные данные",
			req: UpdateProfileRequest{
				Username: "newuser",
				Surname:  "NewSurname",
				Email:    "new@example.com",
			},
			expectedError: false,
		},
		{
			name:          "пустые данные (все опционально)",
			req:           UpdateProfileRequest{},
			expectedError: false,
		},
		{
			name: "невалидный email",
			req: UpdateProfileRequest{
				Email: "invalid-email",
			},
			expectedError: true,
		},
		{
			name: "короткий username",
			req: UpdateProfileRequest{
				Username: "ab",
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()

			if tt.expectedError {
				assert.Error(t, err, "Ожидалась ошибка валидации")
			} else {
				assert.NoError(t, err, "Валидация должна пройти без ошибок")
			}
		})
	}
}

// TestDTO_JSON_Marshaling тестирует сериализацию/десериализацию JSON
func TestDTO_JSON_Marshaling(t *testing.T) {
	t.Run("RegisterRequest JSON marshaling", func(t *testing.T) {
		req := RegisterRequest{
			Username: "testuser",
			Surname:  "TestSurname",
			Email:    "test@example.com",
			Password: "password123",
		}

		// Serialize to JSON
		jsonData, err := json.Marshal(req)
		assert.NoError(t, err)
		assert.Contains(t, string(jsonData), "testuser")
		assert.Contains(t, string(jsonData), "test@example.com")

		// Deserialize from JSON
		var parsedReq RegisterRequest
		err = json.Unmarshal(jsonData, &parsedReq)
		assert.NoError(t, err)
		assert.Equal(t, req.Username, parsedReq.Username)
		assert.Equal(t, req.Email, parsedReq.Email)
	})

	t.Run("LoginResponse JSON marshaling", func(t *testing.T) {
		resp := LoginResponse{
			AccessToken:  "access_token_123",
			RefreshToken: "refresh_token_456",
			TokenType:    "Bearer",
			ExpiresIn:    1800,
		}

		jsonData, err := json.Marshal(resp)
		assert.NoError(t, err)

		var parsedResp LoginResponse
		err = json.Unmarshal(jsonData, &parsedResp)
		assert.NoError(t, err)
		assert.Equal(t, resp.AccessToken, parsedResp.AccessToken)
		assert.Equal(t, resp.TokenType, parsedResp.TokenType)
		assert.Equal(t, resp.ExpiresIn, parsedResp.ExpiresIn)
	})
}

// TestDTO_EdgeCases тестирует граничные случаи
func TestDTO_EdgeCases(t *testing.T) {
	t.Run("очень длинные поля", func(t *testing.T) {
		longString := strings.Repeat("a", 100) // Укорачиваем для тестов

		req := RegisterRequest{
			Username: longString[:51], // Ограничиваем по max длине
			Surname:  longString[:51],
			Email:    "test@example.com",
			Password: "password123",
		}

		err := req.Validate()
		// Должно пройти валидацию в пределах лимитов
		assert.Error(t, err) // Username слишком длинный для alphanum
	})

	t.Run("специальные символы в username", func(t *testing.T) {
		req := RegisterRequest{
			Username: "user@#$%",
			Surname:  "ValidSurname",
			Email:    "valid@example.com",
			Password: "password123",
		}

		err := req.Validate()
		// alphanum валидатор не разрешает спецсимволы
		assert.Error(t, err)
	})

	t.Run("unicode символы", func(t *testing.T) {
		req := RegisterRequest{
			Username: "пользователь",
			Surname:  "Фамилия",
			Email:    "тест@example.com",
			Password: "пароль123",
		}

		err := req.Validate()
		// Кириллические символы должны вызвать ошибки
		assert.Error(t, err)
	})

	t.Run("фамилия с unicode", func(t *testing.T) {
		req := RegisterRequest{
			Username: "validuser",
			Surname:  "Фамилия",
			Email:    "valid@example.com",
			Password: "password123",
		}

		err := req.Validate()
		// Unicode фамилия должна пройти валидацию
		assert.NoError(t, err)
	})
}

// TestDTO_SecurityConcerns тестирует вопросы безопасности
func TestDTO_SecurityConcerns(t *testing.T) {
	t.Run("пароль не должен сериализоваться в JSON ответе", func(t *testing.T) {
		resp := LoginResponse{
			AccessToken:  "access_token",
			RefreshToken: "refresh_token",
			TokenType:    "Bearer",
			ExpiresIn:    1800,
		}

		jsonData, err := json.Marshal(resp)
		assert.NoError(t, err)

		// Убеждаемся, что в JSON нет поля password
		jsonString := string(jsonData)
		assert.NotContains(t, strings.ToLower(jsonString), "password")
	})

	t.Run("ErrorResponse должен быть безопасным", func(t *testing.T) {
		resp := ErrorResponse{
			Error: "Invalid credentials",
		}

		jsonData, err := json.Marshal(resp)
		assert.NoError(t, err)

		jsonString := string(jsonData)
		assert.Contains(t, jsonString, "Invalid credentials")
		// Не должно содержать внутреннюю информацию
		assert.NotContains(t, jsonString, "database")
		assert.NotContains(t, jsonString, "sql")
	})
}

// Benchmark тесты для производительности
func BenchmarkRegisterRequest_Validation(b *testing.B) {
	req := RegisterRequest{
		Username: "testuser",
		Surname:  "TestSurname",
		Email:    "test@example.com",
		Password: "password123",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = req.Validate()
	}
}

func BenchmarkLoginResponse_JSONMarshaling(b *testing.B) {
	resp := LoginResponse{
		AccessToken:  "access_token_123",
		RefreshToken: "refresh_token_456",
		TokenType:    "Bearer",
		ExpiresIn:    1800,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(resp)
	}
}
