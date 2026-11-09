package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/Waycoolers/fmlbot/pkg/errs"

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

	var req domain.LeisureIdeaRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for _, field := range []string{req.Location, req.ActivityLevel, req.Budget} {
		if strings.TrimSpace(field) == "" || utf8.RuneCountInString(field) > 80 {
			http.Error(w, "invalid criteria", http.StatusBadRequest)
			return
		}
	}
	if utf8.RuneCountInString(req.ExtraContext) > 1000 {
		http.Error(w, "extra context is too long", http.StatusBadRequest)
		return
	}
	if !h.ideaLimiter.Allow(strconv.FormatInt(userID, 10)) {
		http.Error(w, "too many requests", http.StatusTooManyRequests)
		return
	}
	idea, err := h.uc.GetLeisureIdea(ctx, userID, req.Location, req.ActivityLevel, req.Budget, req.ExtraContext)
	if errors.Is(err, errs.ErrPartnerNotFound) {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
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
