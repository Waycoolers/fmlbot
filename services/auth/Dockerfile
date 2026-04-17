FROM golang:1.25.4-alpine AS builder
WORKDIR /app

COPY . .
WORKDIR /app/services/auth
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /auth ./cmd/auth

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/
COPY --from=builder /auth .
EXPOSE 8081
CMD ["./auth"]