package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/Waycoolers/fmlbot/pkg/errs"
)

func (h *Handler) VerifyUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.Username == "" {
		http.Error(w, "username is required", http.StatusBadRequest)
		return
	}
	if req.Password == "" {
		http.Error(w, "password is required", http.StatusBadRequest)
		return
	}

	userID, err := h.uc.VerifyUser(ctx, req.Username, req.Password)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if errors.Is(err, errs.ErrWrongPassword) {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		slog.Error("Unexpected error", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	resp := struct {
		UserID int64 `json:"user_id"`
	}{
		UserID: userID,
	}
	sendJson(w, http.StatusOK, resp)
}
