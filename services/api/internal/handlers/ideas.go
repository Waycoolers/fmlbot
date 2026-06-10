package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Waycoolers/fmlbot/pkg/middlewares"
	"github.com/Waycoolers/fmlbot/services/api/internal/domain"
)

func (h *Handler) GetLeisureIdea(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, ok := ctx.Value(middlewares.UserIDKey).(int64)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	resp := struct {
		Idea string `json:"idea"`
	}{
		Idea: idea,
	}
	sendJson(w, http.StatusOK, resp)
}
