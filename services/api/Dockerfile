FROM golang:1.25.4-alpine AS builder
WORKDIR /app

# Копируем весь проект (включая common и все сервисы)
COPY . .

# Переходим в директорию api-сервиса
WORKDIR /app/services/api

# Скачиваем зависимости (теперь common уже на месте)
RUN go mod download

# Собираем бинарный файл
RUN CGO_ENABLED=0 GOOS=linux go build -o /api ./cmd/api

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/
COPY --from=builder /api .
EXPOSE 8080
CMD ["./api"]