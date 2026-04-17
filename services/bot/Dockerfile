FROM golang:1.25.4-alpine AS builder
WORKDIR /app

# Копируем весь проект
COPY . .

# Переходим в директорию bot-сервиса
WORKDIR /app/services/bot

# Скачиваем зависимости
RUN go mod download

# Собираем бинарный файл
RUN CGO_ENABLED=0 GOOS=linux go build -o /bot ./cmd/bot

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/
COPY --from=builder /bot .
EXPOSE 8082
CMD ["./bot"]