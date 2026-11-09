package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/Waycoolers/fmlbot/pkg/middlewares"
	"github.com/Waycoolers/fmlbot/services/api/internal/domain"
)

func (h *Handler) GetLeisureIdea(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value(middlewares.UserIDKey).(int64)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if !h.ideaLimiter.Allow(strconv.FormatInt(userID, 10)) {
		http.Error(w, "too many requests", http.StatusTooManyRequests)
		return
	}

	var req domain.LeisureIdeaRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	idea, err := h.uc.GetLeisureIdea(ctx, req.Location, req.ActivityLevel, req.Budget, req.ExtraContext)
	if err != nil {
		slog.Error("Error getting leisure idea", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	resp := struct {
		Idea string `json:"idea"`
	}{
		Idea: idea,
	}
	sendJson(w, http.StatusOK, resp)
}
