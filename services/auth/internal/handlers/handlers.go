package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/Waycoolers/fmlbot/pkg/errs"
	"github.com/Waycoolers/fmlbot/pkg/password"
	"github.com/Waycoolers/fmlbot/pkg/ratelimit"
	"github.com/Waycoolers/fmlbot/services/auth/internal/config"
	"github.com/Waycoolers/fmlbot/services/auth/internal/domain"
	"github.com/golang-jwt/jwt/v5"
)

type Handler struct {
	repo   domain.TokensRepo
	cfg    *config.Config
	client domain.APIClient
	// Неудачные попытки входа по имени пользователя: защита от перебора пароля
	failedLogins *ratelimit.Limiter
}

func New(repo domain.TokensRepo, cfg *config.Config, client domain.APIClient) (*Handler, error) {
	return &Handler{
		repo:         repo,
		cfg:          cfg,
		client:       client,
		failedLogins: ratelimit.New(10, 15*time.Minute),
	}, nil
}

func sendJson(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		slog.Error("Error sending json", "error", err)
	}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*3)
	defer cancel()

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if req.Username == "" {
		http.Error(w, "invalid username", http.StatusBadRequest)
		return
	}
	if req.Password == "" {
		http.Error(w, "invalid password", http.StatusBadRequest)
		return
	}
	err := password.Validate(req.Password)
	if err != nil {
		if errors.Is(err, errs.ErrPasswordTooShort) || errors.Is(err, errs.ErrPasswordTooLong) ||
			errors.Is(err, errs.ErrPasswordWithoutLower) || errors.Is(err, errs.ErrPasswordWithoutUpper) ||
			errors.Is(err, errs.ErrPasswordWithoutDigit) || errors.Is(err, errs.ErrPasswordWithoutLetter) ||
			errors.Is(err, errs.ErrPasswordInvalidCharacter) {
			body := map[string]string{"error": err.Error()}
			sendJson(w, http.StatusBadRequest, body)
			return
		}
		slog.Error("Unexpected error", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	available, err := h.client.IsUsernameAvailable(ctx, req.Username)
	if err != nil {
		if errors.Is(err, errs.ErrUsernameIsAlreadyTaken) {
			body := map[string]string{"error": err.Error()}
			sendJson(w, http.StatusConflict, body)
			return
		}
		slog.Error("Unexpected error", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if !available {
		body := map[string]string{"error": "username is already taken"}
		sendJson(w, http.StatusConflict, body)
		return
	}

	userID := time.Now().UnixMicro()

	err = h.client.Register(ctx, userID, req.Username, req.Password)
	if err != nil {
		if errors.Is(err, errs.ErrUsernameIsAlreadyTaken) {
			body := map[string]string{"error": err.Error()}
			sendJson(w, http.StatusConflict, body)
			return
		}
		slog.Error("Unexpected error", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	accessToken, err := generateAccessToken(userID, h.cfg.JwtSecret, h.cfg.AccessTokenTTL)
	if err != nil {
		slog.Error("failed to generate access token", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	refreshToken, err := h.repo.Create(ctx, userID, h.cfg.RefreshTokenTTL)
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

func (h *Handler) Token(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*3)
	defer cancel()

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if req.Username == "" {
		http.Error(w, "invalid username", http.StatusBadRequest)
		return
	}
	if req.Password == "" {
		http.Error(w, "invalid password", http.StatusBadRequest)
		return
	}

	loginKey := strings.ToLower(req.Username)
	if h.failedLogins.Blocked(loginKey) {
		http.Error(w, "too many attempts", http.StatusTooManyRequests)
		return
	}

	userID, err := h.client.VerifyUser(ctx, req.Username, req.Password)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, errs.ErrWrongPassword) {
			h.failedLogins.Record(loginKey)
			http.Error(w, "wrong password", http.StatusUnauthorized)
			return
		}
		if errors.Is(err, errs.ErrBadRequest) {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		slog.Error("Unexpected error", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	h.failedLogins.Reset(loginKey)

	accessToken, err := generateAccessToken(userID, h.cfg.JwtSecret, h.cfg.AccessTokenTTL)
	if err != nil {
		slog.Error("failed to generate access token", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	refreshToken, err := h.repo.Create(ctx, userID, h.cfg.RefreshTokenTTL)
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

func (h *Handler) InternalToken(w http.ResponseWriter, r *http.Request) {
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

	accessToken, err := generateAccessToken(req.UserID, h.cfg.JwtSecret, h.cfg.AccessTokenTTL)
	if err != nil {
		slog.Error("failed to generate access token", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Внутренним клиентам нужен только access-токен: refresh здесь лишь копил бы строки в БД
	resp := struct {
		AccessToken string `json:"access_token"`
	}{
		AccessToken: accessToken,
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

	userID, err := h.repo.Consume(ctx, req.RefreshToken)
	if err != nil {
		if errors.Is(err, errs.ErrTokenNotFound) || errors.Is(err, errs.ErrTokenInvalid) {
			slog.Warn("refresh token rejected", "error", err)
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		slog.Error("failed to validate refresh token", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

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
		if errors.Is(err, errs.ErrTokenInvalid) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		slog.Error("failed to revoke refresh token", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) InternalRevokeAll(w http.ResponseWriter, r *http.Request) {
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

	if err := h.repo.RevokeAll(ctx, req.UserID); err != nil {
		slog.Error("failed to revoke user tokens", "error", err)
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

func generateAccessToken(userID int64, secret []byte, ttl time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(ttl).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}
