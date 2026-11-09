package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/Waycoolers/fmlbot/pkg/errs"
	"github.com/Waycoolers/fmlbot/pkg/middlewares"
	"github.com/Waycoolers/fmlbot/pkg/password"
	"github.com/Waycoolers/fmlbot/services/api/internal/domain"
)

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req struct {
		UserID   int64  `json:"user_id"`
		Username string `json:"username"`
		Password string `json:"password"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if !isValidUsername(req.Username) {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}
	if req.Password == "" {
		http.Error(w, "Password is required", http.StatusBadRequest)
		return
	}

	err = h.uc.RegisterUser(ctx, req.UserID, req.Username, req.Password)
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
	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) AddUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req struct {
		Username   string `json:"username"`
		TelegramID int64  `json:"telegram_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	userID := req.TelegramID
	if r.URL.Path == "/users" {
		userID, _ = ctx.Value(middlewares.UserIDKey).(int64)
	}
	if userID <= 0 || userID >= 1_000_000_000_000_000 {
		http.Error(w, "invalid Telegram ID", http.StatusBadRequest)
		return
	}
	if !isValidUsername(req.Username) {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}

	generatedPassword, err := h.uc.AddUserWithRandomPassword(ctx, userID, req.Username)
	if err != nil {
		if errors.Is(err, errs.ErrUserExists) {
			sendJson(w, http.StatusConflict, map[string]string{"error": err.Error()})
			return
		}
		if errors.Is(err, errs.ErrUsernameIsAlreadyTaken) {
			sendJson(w, http.StatusConflict, map[string]string{"error": err.Error()})
			return
		}
		slog.Error("Unexpected error", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	resp := struct {
		Password string `json:"password"`
	}{
		Password: generatedPassword,
	}
	sendJson(w, http.StatusCreated, resp)
}

func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value(middlewares.UserIDKey).(int64)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	err := h.uc.RemoveUser(ctx, userID)
	if err != nil {
		if errors.Is(err, errs.ErrPairChanged) {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		if errors.Is(err, errs.ErrUserNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		slog.Error("Unexpected error", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, ok := ctx.Value(middlewares.UserIDKey).(int64)
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
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	sendJson(w, http.StatusOK, user)
}

func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, ok := ctx.Value(middlewares.UserIDKey).(int64)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var req struct {
		OldPassword string `json:"old_password"`
		Password    string `json:"password"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err = password.Validate(req.Password); err != nil {
		sendJson(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	err = h.uc.ChangePassword(ctx, id, req.OldPassword, req.Password)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		// Не 401: токен действителен, а неверен только пароль
		if errors.Is(err, errs.ErrWrongPassword) {
			sendJson(w, http.StatusForbidden, map[string]string{"error": err.Error()})
			return
		}
		slog.Error("Unexpected error", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// ResetPassword - внутренний метод для бота, закрыт X-Internal-Secret.
func (h *Handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, ok := ctx.Value(middlewares.UserIDKey).(int64)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	generatedPassword, err := h.uc.ResetPassword(ctx, id)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		slog.Error("Unexpected error", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	sendJson(w, http.StatusOK, map[string]string{"password": generatedPassword})
}

func (h *Handler) GetPartner(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, ok := ctx.Value(middlewares.UserIDKey).(int64)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.uc.GetPartner(ctx, id)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) || errors.Is(err, errs.ErrPartnerNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		slog.Error("Unexpected error", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	sendJson(w, http.StatusOK, user)
}

func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value(middlewares.UserIDKey).(int64)
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

	if !isValidUsername(req.Username) {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}

	// partner_id из запроса намеренно игнорируется: партнёр меняется только через /users/pair и /users/unpair
	err = h.uc.UpdateUser(ctx, userID, req.Username)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if errors.Is(err, errs.ErrUsernameIsAlreadyTaken) {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		slog.Error("Unexpected error", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// UpdatePartner закрыт: пользователь не может менять имя и связи другого пользователя.
func (h *Handler) UpdatePartner(w http.ResponseWriter, _ *http.Request) {
	sendJson(w, http.StatusForbidden, map[string]string{"error": errs.ErrForbidden.Error()})
}

func writePairError(w http.ResponseWriter, err error) {
	body := map[string]string{"error": err.Error()}
	switch {
	case errors.Is(err, errs.ErrUserNotFound), errors.Is(err, errs.ErrPartnerRequestNotFound):
		sendJson(w, http.StatusNotFound, body)
	case errors.Is(err, errs.ErrCannotPartnerYourself):
		sendJson(w, http.StatusBadRequest, body)
	case errors.Is(err, errs.ErrAlreadyHasPartner), errors.Is(err, errs.ErrPartnerAlreadyHasPartner):
		sendJson(w, http.StatusForbidden, body)
	default:
		slog.Error("Unexpected error", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

func (h *Handler) AddPartners(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value(middlewares.UserIDKey).(int64)
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
	if req.PartnerID == 0 {
		http.Error(w, "partner_id is required", http.StatusBadRequest)
		return
	}
	if !h.pairLimiter.Allow(strconv.FormatInt(userID, 10)) {
		http.Error(w, "too many requests", http.StatusTooManyRequests)
		return
	}

	requestID, err := h.uc.RequestPartner(ctx, userID, req.PartnerID)
	if err != nil {
		writePairError(w, err)
		return
	}
	sendJson(w, http.StatusAccepted, map[string]int64{"request_id": requestID})
}

func (h *Handler) GetPartnerRequests(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value(middlewares.UserIDKey).(int64)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	requests, err := h.uc.GetPartnerRequests(ctx, userID)
	if err != nil {
		writePairError(w, err)
		return
	}
	sendJson(w, http.StatusOK, requests)
}

func (h *Handler) answerPartnerRequest(w http.ResponseWriter, r *http.Request, answer func(ctx context.Context, requestID, userID int64) (int64, error)) {
	ctx := r.Context()
	userID, ok := ctx.Value(middlewares.UserIDKey).(int64)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	requestID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || requestID <= 0 {
		http.Error(w, "Request id is incorrect", http.StatusBadRequest)
		return
	}

	fromUserID, err := answer(ctx, requestID, userID)
	if err != nil {
		writePairError(w, err)
		return
	}
	sendJson(w, http.StatusOK, map[string]int64{"from_user_id": fromUserID})
}

func (h *Handler) AcceptPartnerRequest(w http.ResponseWriter, r *http.Request) {
	h.answerPartnerRequest(w, r, h.uc.AcceptPartnerRequest)
}

func (h *Handler) DeclinePartnerRequest(w http.ResponseWriter, r *http.Request) {
	h.answerPartnerRequest(w, r, h.uc.DeclinePartnerRequest)
}

func (h *Handler) DeletePartners(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, ok := ctx.Value(middlewares.UserIDKey).(int64)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var expectedVersion *int64
	if raw := r.URL.Query().Get("pair_version"); raw != "" {
		version, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || version < 0 {
			http.Error(w, "invalid pair version", http.StatusBadRequest)
			return
		}
		expectedVersion = &version
	}
	err := h.uc.RemovePartners(ctx, id, expectedVersion)
	if err != nil {
		if errors.Is(err, errs.ErrPairChanged) {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
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
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
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

func (h *Handler) SaveFCMToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := ctx.Value(middlewares.UserIDKey).(int64)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var req struct {
		SCMToken string `json:"fcm_token"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.SCMToken == "" {
		http.Error(w, "fcm_token is required", http.StatusBadRequest)
		return
	}

	err = h.uc.ProcessFCMToken(ctx, userID, req.SCMToken)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		slog.Error("Unexpected error", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
