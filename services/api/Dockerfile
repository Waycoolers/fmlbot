FROM golang:1.25.4-alpine AS builder
WORKDIR /app

# Копируем ВЕСЬ проект (корневой каталог)
COPY . .

# Переходим в директорию сервиса
WORKDIR /app/services/api

# Скачиваем зависимости
RUN go mod download

# Собираем бинарник
RUN CGO_ENABLED=0 GOOS=linux go build -o /api ./cmd/api

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/
COPY --from=builder /api .
EXPOSE 8080
CMD ["./api"]