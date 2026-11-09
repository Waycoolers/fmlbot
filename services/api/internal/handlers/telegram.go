package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/Waycoolers/fmlbot/pkg/errs"
	"github.com/Waycoolers/fmlbot/pkg/middlewares"
)

func telegramError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, errs.ErrUserNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, errs.ErrTelegramLinkInvalid):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, errs.ErrTelegramAlreadyLinked):
		http.Error(w, err.Error(), http.StatusConflict)
	default:
		slog.Error("Telegram account request failed", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

func (h *Handler) CreateTelegramLink(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value(middlewares.UserIDKey).(int64)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	code, err := h.uc.CreateTelegramLink(ctx, userID)
	if err != nil {
		telegramError(w, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	sendJson(w, http.StatusCreated, map[string]string{"code": code, "command": "/link " + code})
}

func (h *Handler) LinkTelegram(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TelegramID int64  `json:"telegram_id"`
		Code       string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.TelegramID <= 0 {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if err := h.uc.LinkTelegram(r.Context(), req.TelegramID, req.Code); err != nil {
		telegramError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ResolveTelegram(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "invalid ID", http.StatusBadRequest)
		return
	}

	var key string
	if r.PathValue("kind") == "user" {
		id, err = h.uc.ResolveTelegramUser(r.Context(), id)
		key = "user_id"
	} else if r.PathValue("kind") == "chat" {
		id, err = h.uc.ResolveTelegramChat(r.Context(), id)
		key = "telegram_id"
	} else {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		telegramError(w, err)
		return
	}
	sendJson(w, http.StatusOK, map[string]int64{key: id})
}
