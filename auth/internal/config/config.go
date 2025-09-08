package config

import (
	"log"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
}

type ServerConfig struct {
	Port string
	Mode string // gin.Mode: debug, release, test
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type JWTConfig struct {
	Secret          string
	AccessTokenTTL  string // время жизни access token
	RefreshTokenTTL string // время жизни refresh token
}

func Load() *Config {
	// В Docker полагаемся на переменные окружения, установленные контейнером
	// godotenv.Load() не нужен, так как Docker уже устанавливает переменные

	return &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
			Mode: getEnv("GIN_MODE", "debug"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "ai_hr_user"),
			Password: getEnvRequired("DB_PASSWORD"),
			DBName:   getEnv("DB_NAME", "ai_hr_db"), // Исправлено имя БД
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		JWT: JWTConfig{
			Secret:          getEnvRequired("JWT_SECRET"),
			AccessTokenTTL:  getEnv("JWT_ACCESS_TTL", "30m"),
			RefreshTokenTTL: getEnv("JWT_REFRESH_TTL", "7d"),
		},
	}
}

//func Load() *Config {
//	// Загружаем .env файл (игнорируем ошибку если файла нет)
//	//_ = godotenv.Load("../../.env")
//
//	return &Config{
//		Server: ServerConfig{
//			Port: getEnv("PORT", "8080"),
//			Mode: getEnv("GIN_MODE", "debug"),
//		},
//		Database: DatabaseConfig{
//			Host:     getEnv("DB_HOST", "localhost"),
//			Port:     getEnv("DB_PORT", "5432"),
//			User:     getEnv("DB_USER", "postgres"),
//			Password: getEnvRequired("DB_PASSWORD"), // Обязательная переменная
//			DBName:   getEnv("DB_NAME", "ai_hr_service_db"),
//			SSLMode:  getEnv("DB_SSLMODE", "disable"),
//		},
//		JWT: JWTConfig{
//			Secret:          getEnvRequired("JWT_SECRET"),    // Обязательная переменная
//			AccessTokenTTL:  getEnv("JWT_ACCESS_TTL", "30m"), // 30 минут
//			RefreshTokenTTL: getEnv("JWT_REFRESH_TTL", "7d"), // 7 дней
//		},
//	}
//}

// Вспомогательные функции
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvRequired(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("Environment variable %s is required", key)
	}
	return value
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue string) time.Duration {
	value := getEnv(key, defaultValue)
	duration, err := time.ParseDuration(value)
	if err != nil {
		log.Fatalf("Invalid duration format for %s: %s", key, value)
	}
	return duration
}

// Функция для генерации случайного JWT секрета (для разработки)
func GenerateJWTSecret() string {
	// Используй только для разработки!
	// В продакшене используй криптографически стойкие генераторы
	return "dev-secret-key-change-in-production-" + strconv.FormatInt(time.Now().UnixNano(), 10)
}
