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
		w.WriteHeader(http.StatusInternalServerError)
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
		w.WriteHeader(http.StatusInternalServerError)
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
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) ResetMyUserConfig(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, ok := ctx.Value(jwtmiddleware.UserIDKey).(int64)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	err := h.uc.ResetMyUserConfig(ctx, id)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		slog.Error("Unexpected error", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) ResetPartnerUserConfig(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, ok := ctx.Value(jwtmiddleware.UserIDKey).(int64)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	err := h.uc.ResetPartnerUserConfig(ctx, id)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			body := map[string]string{
				"error": err.Error(),
			}
			sendJson(w, http.StatusNotFound, body)
			return
		}
		if errors.Is(err, errs.ErrPartnerNotFound) {
			body := map[string]string{
				"error": err.Error(),
			}
			sendJson(w, http.StatusNotFound, body)
			return
		}
		slog.Error("Unexpected error", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
