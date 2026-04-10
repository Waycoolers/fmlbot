package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/Waycoolers/fmlbot/common/errs"
	"github.com/Waycoolers/fmlbot/common/jwtmiddleware"
	"github.com/Waycoolers/fmlbot/services/api/internal/domain"
)

func (h *Handler) AddImportantDate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value(jwtmiddleware.UserIDKey).(int64)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req domain.ImportantDateRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := h.uc.AddImportantDate(ctx, userID, req)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) || errors.Is(err, errs.ErrPartnerNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		slog.Error("Unexpected error", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sendJson(w, http.StatusCreated, resp)
}

func (h *Handler) RemoveImportantDate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value(jwtmiddleware.UserIDKey).(int64)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	importantDateID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "URL path does not contain id", http.StatusBadRequest)
		return
	}
	if importantDateID <= 0 {
		http.Error(w, "Important date id is incorrect", http.StatusBadRequest)
	}

	err = h.uc.RemoveImportantDate(ctx, userID, int64(importantDateID))
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) || errors.Is(err, errs.ErrImportantDateNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		slog.Error("Unexpected error", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) UpdateImportantDate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value(jwtmiddleware.UserIDKey).(int64)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	importantDateID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "URL path does not contain id", http.StatusBadRequest)
		return
	}
	if importantDateID <= 0 {
		http.Error(w, "Important date id is incorrect", http.StatusBadRequest)
	}
	var req domain.ImportantDateRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.uc.UpdateImportantDate(ctx, userID, int64(importantDateID), req)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) || errors.Is(err, errs.ErrImportantDateNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		slog.Error("Unexpected error", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) GetImportantDate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value(jwtmiddleware.UserIDKey).(int64)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	importantDateID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "URL path does not contain id", http.StatusBadRequest)
		return
	}
	if importantDateID <= 0 {
		http.Error(w, "Important date id is incorrect", http.StatusBadRequest)
	}

	resp, err := h.uc.GetImportantDate(ctx, userID, int64(importantDateID))
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) || errors.Is(err, errs.ErrImportantDateNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		slog.Error("Unexpected error", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sendJson(w, http.StatusOK, resp)
}

func (h *Handler) GetAllImportantDates(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value(jwtmiddleware.UserIDKey).(int64)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	resp, err := h.uc.GetAllImportantDates(ctx, userID)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		slog.Error("Unexpected error", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sendJson(w, http.StatusOK, resp)
}

func (h *Handler) UpdateImportantDateSharing(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value(jwtmiddleware.UserIDKey).(int64)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	importantDateID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "URL path does not contain id", http.StatusBadRequest)
		return
	}
	if importantDateID <= 0 {
		http.Error(w, "Important date id is incorrect", http.StatusBadRequest)
	}

	var req struct {
		MakeShared bool `json:"make_shared"`
	}
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = h.uc.UpdateImportantDateSharing(ctx, int64(importantDateID), userID, req.MakeShared)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) || errors.Is(err, errs.ErrImportantDateNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if errors.Is(err, errs.ErrPartnerNotFound) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		slog.Error("Unexpected error", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
