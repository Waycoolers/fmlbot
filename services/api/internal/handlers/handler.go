package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Waycoolers/fmlbot/pkg/ratelimit"
	"github.com/Waycoolers/fmlbot/services/api/internal/domain"
	"github.com/Waycoolers/fmlbot/services/api/internal/usecases"
)

type Handler struct {
	uc       *usecases.UseCase
	telegram domain.Sender
	// Запросы к платной нейросети на одного пользователя
	ideaLimiter *ratelimit.Limiter
	// Заявки на пару от одного пользователя
	pairLimiter *ratelimit.Limiter
}

func New(uc *usecases.UseCase, telegram domain.Sender) *Handler {
	return &Handler{
		uc:          uc,
		telegram:    telegram,
		ideaLimiter: ratelimit.New(30, time.Hour),
		pairLimiter: ratelimit.New(20, time.Hour),
	}
}

func isValidUsername(username string) bool {
	return strings.TrimSpace(username) != "" && utf8.RuneCountInString(username) <= 255
}

func sendJson(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		slog.Error("Error sending json", "error", err)
	}
}
