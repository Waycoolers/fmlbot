# syntax=docker/dockerfile:1
FROM golang:1.25.4-alpine AS builder

WORKDIR /app

# Скачиваем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем проект
COPY . .

# Собираем бинарник
RUN go build -o fmlbot cmd/fmlbot/main.go

# Минимальный образ для запуска
FROM alpine:3.18

WORKDIR /app

# Устанавливаем таймзону
RUN apk add --no-cache tzdata
ENV TZ=Europe/Moscow

# Копируем бинарник из builder
COPY --from=builder /app/fmlbot .

# Копируем конфиг
COPY .env .

# Переменные окружения
ENV GIN_MODE=release

# Запуск
CMD ["./fmlbot"]
