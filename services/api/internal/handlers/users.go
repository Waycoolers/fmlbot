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

func (h *Handler) AddUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req domain.UserRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	userID, ok := ctx.Value(jwtmiddleware.UserIDKey).(int64)
	if !ok {
		http.Error(w, "unauthorized", http.StatusBadRequest)
		return
	}
	if req.Username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}

	err = h.uc.AddUser(ctx, userID, req.Username)
	if err != nil {
		if errors.Is(err, errs.ErrUserExists) {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		slog.Error("Unexpected error", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value(jwtmiddleware.UserIDKey).(int64)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	err := h.uc.RemoveUser(ctx, userID)
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

func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, ok := ctx.Value(jwtmiddleware.UserIDKey).(int64)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.uc.GetMe(ctx, id)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		slog.Error("Unexpected error", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sendJson(w, http.StatusOK, user)
}

func (h *Handler) GetPartner(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, ok := ctx.Value(jwtmiddleware.UserIDKey).(int64)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.uc.GetPartner(ctx, id)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		slog.Error("Unexpected error", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sendJson(w, http.StatusOK, user)
}

func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value(jwtmiddleware.UserIDKey).(int64)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var req domain.UserRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.uc.UpdateUser(ctx, userID, req.Username, req.PartnerID)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		slog.Error("Unexpected error", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) UpdatePartner(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value(jwtmiddleware.UserIDKey).(int64)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var req domain.UserRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.uc.UpdatePartner(ctx, userID, req.Username, req.PartnerID)
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) AddPartners(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value(jwtmiddleware.UserIDKey).(int64)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var req struct {
		PartnerID int64 `json:"partner_id"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.uc.AddPartners(ctx, userID, req.PartnerID)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		slog.Error("Unexpected error", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) DeletePartners(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, ok := ctx.Value(jwtmiddleware.UserIDKey).(int64)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	err := h.uc.RemovePartners(ctx, id)
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) GetUserByUsername(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	username := r.PathValue("username")
	if username == "" {
		http.Error(w, "username is required", http.StatusBadRequest)
		return
	}

	user, err := h.uc.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			body := map[string]string{
				"error": err.Error(),
			}
			sendJson(w, http.StatusNotFound, body)
			return
		}
		slog.Error("Unexpected error", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	sendJson(w, http.StatusOK, user)
}
