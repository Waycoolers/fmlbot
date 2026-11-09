FROM golang:1.25.4-alpine AS builder
WORKDIR /app

COPY go.work go.work.sum ./
COPY pkg/go.mod pkg/go.sum ./pkg/
COPY services/api/go.mod services/api/go.sum ./services/api/
COPY services/auth/go.mod services/auth/go.sum ./services/auth/
COPY services/bot/go.mod services/bot/go.sum ./services/bot/

RUN go mod download

WORKDIR /app/services/auth
COPY pkg/ /app/pkg/
COPY services/auth/ ./
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -o /auth ./cmd/auth

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/
COPY --from=builder /auth .
EXPOSE 8081
CMD ["./auth"]
