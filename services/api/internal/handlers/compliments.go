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

func (h *Handler) AddCompliment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value(jwtmiddleware.UserIDKey).(int64)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var compliment domain.ComplimentRequest
	err := json.NewDecoder(r.Body).Decode(&compliment)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if compliment.Text == "" {
		http.Error(w, "Text is required", http.StatusBadRequest)
		return
	}

	resp, err := h.uc.AddCompliment(ctx, userID, compliment.Text)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		slog.Error("Unexpected error", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sendJson(w, http.StatusCreated, resp)
}

func (h *Handler) GetAllCompliments(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value(jwtmiddleware.UserIDKey).(int64)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	compliments, err := h.uc.GetAllCompliments(ctx, userID)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		slog.Error("Unexpected error", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sendJson(w, http.StatusOK, compliments)
}

func (h *Handler) RemoveCompliment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value(jwtmiddleware.UserIDKey).(int64)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	complimentID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "URL path does not contain id", http.StatusBadRequest)
		return
	}
	if complimentID <= 0 {
		http.Error(w, "Compliment id is incorrect", http.StatusBadRequest)
		return
	}

	err = h.uc.RemoveCompliment(ctx, userID, int64(complimentID))
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		slog.Error("Unexpected error", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) UpdateCompliment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value(jwtmiddleware.UserIDKey).(int64)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	complimentID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "URL path does not contain id", http.StatusBadRequest)
		return
	}
	if complimentID <= 0 {
		http.Error(w, "Compliment id is incorrect", http.StatusBadRequest)
		return
	}
	var compliment domain.ComplimentRequest
	err = json.NewDecoder(r.Body).Decode(&compliment)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if compliment.Text == "" {
		http.Error(w, "Text is required", http.StatusBadRequest)
		return
	}

	err = h.uc.UpdateCompliment(ctx, userID, int64(complimentID), &compliment)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) || errors.Is(err, errs.ErrComplimentNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		slog.Error("Unexpected error", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) ReceiveCompliment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value(jwtmiddleware.UserIDKey).(int64)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	compliment, err := h.uc.AcquireCompliment(ctx, userID)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			body := map[string]string{
				"error": err.Error(),
			}
			sendJson(w, http.StatusNotFound, body)
			return
		}
		if errors.Is(err, errs.ErrNoCompliments) {
			body := map[string]string{
				"error": err.Error(),
			}
			sendJson(w, http.StatusGone, body)
			return
		}
		if errors.Is(err, errs.ErrLimitExceeded) {
			body := map[string]string{
				"error": err.Error(),
			}
			sendJson(w, http.StatusTooManyRequests, body)
			return
		}
		var errBucketEmpty *errs.ErrBucketEmpty
		if errors.As(err, &errBucketEmpty) {
			body := map[string]any{
				"error":   errBucketEmpty.Error(),
				"minutes": errBucketEmpty.Minutes,
			}
			sendJson(w, http.StatusTooManyRequests, body)
			return
		}
		slog.Error("Unexpected error", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sendJson(w, http.StatusOK, compliment)
}
