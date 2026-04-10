package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/Waycoolers/fmlbot/services/auth/internal/config"
	"github.com/Waycoolers/fmlbot/services/auth/internal/domain"
	"github.com/golang-jwt/jwt/v5"
)

type Handler struct {
	repo domain.TokensRepo
	cfg  *config.Config
}

func New(repo domain.TokensRepo, cfg *config.Config) (*Handler, error) {
	return &Handler{repo: repo, cfg: cfg}, nil
}

func (h *Handler) Token(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*3)
	defer cancel()

	var req struct {
		UserID int64 `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if req.UserID <= 0 {
		http.Error(w, "invalid user_id", http.StatusBadRequest)
		return
	}

	// Генерируем access token
	accessToken, err := generateAccessToken(req.UserID, h.cfg.JwtSecret, h.cfg.AccessTokenTTL)
	if err != nil {
		slog.Error("failed to generate access token", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Создаём refresh token
	refreshToken, err := h.repo.Create(ctx, req.UserID, h.cfg.RefreshTokenTTL)
	if err != nil {
		slog.Error("failed to create refresh token", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	resp := struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		return
	}
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*3)
	defer cancel()

	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if req.RefreshToken == "" {
		http.Error(w, "refresh_token required", http.StatusBadRequest)
		return
	}

	userID, err := h.repo.Validate(ctx, req.RefreshToken)
	if err != nil {
		if errors.Is(err, domain.ErrTokenNotFound) {
			slog.Warn("refresh token not found", "error", err)
			http.Error(w, domain.ErrTokenNotFound.Error(), http.StatusUnauthorized)
			return
		}
		slog.Error("failed to validate refresh token", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Генерируем новый access token
	newAccessToken, err := generateAccessToken(userID, h.cfg.JwtSecret, h.cfg.AccessTokenTTL)
	if err != nil {
		slog.Error("failed to generate new access token", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	newRefreshToken, err := h.repo.Create(ctx, userID, h.cfg.RefreshTokenTTL)
	if err != nil {
		slog.Error("failed to create new refresh token", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	resp := struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		return
	}
}

func (h *Handler) Revoke(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*3)
	defer cancel()

	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if req.RefreshToken == "" {
		http.Error(w, "refresh_token required", http.StatusBadRequest)
		return
	}

	if err := h.repo.Revoke(ctx, req.RefreshToken); err != nil {
		slog.Error("failed to revoke refresh token", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Health(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("OK"))
	if err != nil {
		return
	}
}

// generateAccessToken создаёт JWT access token
func generateAccessToken(userID int64, secret []byte, ttl time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(ttl).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}
