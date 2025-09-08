FROM golang:1.24-alpine AS builder

# Устанавливаем необходимые пакеты
RUN apk add --no-cache git

# Создаем рабочую директорию
WORKDIR /app

# Копируем go.mod и go.sum для кэширования зависимостей
COPY go.mod go.sum ./

# Загружаем зависимости
RUN go mod download

# Копируем весь исходный код
COPY . .

# Собираем приложение
# CGO_ENABLED=0 для статической сборки
# -ldflags="-w -s" для уменьшения размера бинарника
RUN go build -o main ./cmd/main.go
# Этап 2: Финальный образ
FROM alpine:latest

# Устанавливаем ca-certificates для HTTPS и timezone data
RUN apk --no-cache add ca-certificates

# Создаем рабочую директорию
WORKDIR /app

# Копируем скомпилированный бинарник из builder этапа
COPY --from=builder /app/main .

#Копируем .env файл (если нужен)
COPY --from=builder /app/.env .env

# Открываем порт (должен совпадать с PORT в переменных окружения)
EXPOSE 8081
# Запускаем приложение
CMD ["./main"]
