package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/Waycoolers/fmlbot/common/errs"
	"github.com/Waycoolers/fmlbot/common/jwtmiddleware"
	"github.com/Waycoolers/fmlbot/services/api/internal/domain"
)

func (h *Handler) GetMyUserConfig(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, ok := ctx.Value(jwtmiddleware.UserIDKey).(int64)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	userConfig, err := h.uc.GetMyUserConfig(ctx, id)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		slog.Error("Unexpected error", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	sendJson(w, http.StatusOK, userConfig)
}

func (h *Handler) GetPartnerUserConfig(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, ok := ctx.Value(jwtmiddleware.UserIDKey).(int64)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	userConfig, err := h.uc.GetPartnerUserConfig(ctx, id)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		slog.Error("Unexpected error", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	sendJson(w, http.StatusOK, userConfig)
}

func (h *Handler) UpdateUserConfig(w http.ResponseWriter, r *http.Request) {
	var userConfig domain.UserConfigPatch
	ctx := r.Context()
	id, ok := ctx.Value(jwtmiddleware.UserIDKey).(int64)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	err := json.NewDecoder(r.Body).Decode(&userConfig)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.uc.UpdateUserConfig(ctx, id, &userConfig)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		slog.Error("Unexpected error", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}
