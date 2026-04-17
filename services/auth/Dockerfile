FROM golang:1.25.4-alpine AS builder
WORKDIR /app

# Копируем весь проект
COPY . .

# Переходим в директорию auth-сервиса
WORKDIR /app/services/auth

# Скачиваем зависимости
RUN go mod download

# Собираем бинарный файл
RUN CGO_ENABLED=0 GOOS=linux go build -o /auth ./cmd/auth

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/
COPY --from=builder /auth .
EXPOSE 8081
CMD ["./auth"]