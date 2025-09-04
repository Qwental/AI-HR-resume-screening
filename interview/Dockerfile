# Dockerfile для Interview Service
FROM golang:1.23-alpine AS builder

# Устанавливаем git для загрузки зависимостей
RUN apk add --no-cache git

WORKDIR /app

# Копируем go mod файлы
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь код
COPY . .

# Собираем приложение
RUN go build -o main ./cmd/main.go

# Финальный образ
FROM alpine:latest

# Устанавливаем CA сертификаты для HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Копируем бинарник
COPY --from=builder /app/main .

# Открываем порт
EXPOSE 8081

# Запускаем приложение
CMD ["./main"]
