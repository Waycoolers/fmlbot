package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/Waycoolers/fmlbot/services/api/internal/usecases"
)

type Handler struct {
	uc *usecases.UseCase
}

func New(uc *usecases.UseCase) *Handler {
	return &Handler{uc: uc}
}

func sendJson(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		slog.Error("Error sending json", "error", err)
	}
}
