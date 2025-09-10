# Multi-stage build для auth service
FROM golang:1.24-alpine AS builder

# Устанавливаем git для загрузки зависимостей
RUN apk add --no-cache git

WORKDIR /app

# Копируем go mod файлы и загружаем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь код
COPY . .

# Собираем приложение
RUN go build -o main ./cmd/server/main.go

# Финальный образ
FROM alpine:latest

# Устанавливаем CA сертификаты для HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Копируем бинарник из builder stage
COPY --from=builder /app/main .

# Копируем статические файлы
#COPY --from=builder /app/static ./static

# Открываем порт
EXPOSE 8080

# Запускаем приложение
CMD ["./main"]
