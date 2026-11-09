package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/Waycoolers/fmlbot/pkg/errs"
	"github.com/Waycoolers/fmlbot/services/bot/internal/domain"
)

func (c *client) CreateUser(ctx context.Context, chatID int64, username string) (string, error) {
	reqBody := struct {
		Username string `json:"username"`
	}{
		Username: username,
	}

	resp, err := c.doAuthRequest(ctx, http.MethodPost, "/users", reqBody, chatID)
	if err != nil {
		return "", err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	switch resp.StatusCode {
	case http.StatusCreated:
		var respBody struct {
			Password string `json:"password"`
		}
		err = json.NewDecoder(resp.Body).Decode(&respBody)
		if err != nil {
			return "", err
		}
		return respBody.Password, nil
	case http.StatusConflict:
		return "", errs.ErrUserExists
	default:
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}

func (c *client) GetMe(ctx context.Context, chatID int64) (*domain.User, error) {
	resp, err := c.doAuthRequest(ctx, http.MethodGet, "/users/me", nil, chatID)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	switch resp.StatusCode {
	case http.StatusOK:
		var user domain.User
		err = json.NewDecoder(resp.Body).Decode(&user)
		if err != nil {
			return nil, err
		}
		return &user, nil
	case http.StatusNotFound:
		return nil, errs.ErrUserNotFound
	default:
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}

func (c *client) GetPartner(ctx context.Context, chatID int64) (*domain.User, error) {
	resp, err := c.doAuthRequest(ctx, http.MethodGet, "/users/partner", nil, chatID)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	switch resp.StatusCode {
	case http.StatusOK:
		var user domain.User
		err = json.NewDecoder(resp.Body).Decode(&user)
		if err != nil {
			return nil, err
		}
		return &user, nil
	case http.StatusNotFound:
		return nil, errs.ErrUserNotFound
	default:
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}

func (c *client) DeleteMe(ctx context.Context, chatID int64) error {
	resp, err := c.doAuthRequest(ctx, http.MethodDelete, "/users/me", nil, chatID)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	switch resp.StatusCode {
	case http.StatusNoContent:
		return nil
	case http.StatusNotFound:
		return errs.ErrUserNotFound
	default:
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}

func decodePairError(resp *http.Response) error {
	var errBody struct {
		Error string `json:"error"`
	}
	if decodeErr := json.NewDecoder(resp.Body).Decode(&errBody); decodeErr != nil {
		return fmt.Errorf("unexpected %d response: %w", resp.StatusCode, decodeErr)
	}
	if errBody.Error == errs.ErrPartnerAlreadyHasPartner.Error() {
		return errs.ErrPartnerAlreadyHasPartner
	}
	return errs.ErrAlreadyHasPartner
}

func (c *client) RequestPartner(ctx context.Context, requesterID int64, partnerID int64) (int64, error) {
	reqBody := struct {
		PartnerID int64 `json:"partner_id"`
	}{
		PartnerID: partnerID,
	}
	resp, err := c.doAuthRequest(ctx, http.MethodPost, "/users/pair", reqBody, requesterID)
	if err != nil {
		return 0, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	switch resp.StatusCode {
	case http.StatusAccepted:
		var respBody struct {
			RequestID int64 `json:"request_id"`
		}
		if err = json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
			return 0, err
		}
		return respBody.RequestID, nil
	case http.StatusNotFound:
		return 0, errs.ErrPartnerNotFound
	case http.StatusForbidden:
		return 0, decodePairError(resp)
	default:
		return 0, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
}

func (c *client) answerPartnerRequest(ctx context.Context, chatID int64, requestID int64, answer string) (int64, error) {
	path := fmt.Sprintf("/users/pair/requests/%d/%s", requestID, answer)
	resp, err := c.doAuthRequest(ctx, http.MethodPost, path, nil, chatID)
	if err != nil {
		return 0, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	switch resp.StatusCode {
	case http.StatusOK:
		var respBody struct {
			FromUserID int64 `json:"from_user_id"`
		}
		if err = json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
			return 0, err
		}
		return respBody.FromUserID, nil
	case http.StatusNotFound:
		return 0, errs.ErrPartnerRequestNotFound
	case http.StatusForbidden:
		return 0, decodePairError(resp)
	default:
		return 0, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
}

func (c *client) AcceptPartnerRequest(ctx context.Context, chatID int64, requestID int64) (int64, error) {
	return c.answerPartnerRequest(ctx, chatID, requestID, "accept")
}

func (c *client) DeclinePartnerRequest(ctx context.Context, chatID int64, requestID int64) (int64, error) {
	return c.answerPartnerRequest(ctx, chatID, requestID, "decline")
}

func (c *client) Unpair(ctx context.Context, chatID int64) error {
	resp, err := c.doAuthRequest(ctx, http.MethodPatch, "/users/unpair", nil, chatID)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusNotFound:
		var errBody struct {
			Error string `json:"error"`
		}
		if decodeErr := json.NewDecoder(resp.Body).Decode(&errBody); decodeErr != nil {
			return fmt.Errorf("unexpected 404 response: %w", decodeErr)
		}
		if strings.Contains(errBody.Error, "partner") {
			return errs.ErrPartnerNotFound
		}
		return errs.ErrUserNotFound
	default:
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
}

func (c *client) GetUserByUsername(ctx context.Context, requesterID int64, username string) (*domain.User, error) {
	path := "/users/by-username/" + url.PathEscape(username)
	resp, err := c.doAuthRequest(ctx, http.MethodGet, path, nil, requesterID)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	switch resp.StatusCode {
	case http.StatusOK:
		var user domain.User
		err = json.NewDecoder(resp.Body).Decode(&user)
		if err != nil {
			return nil, err
		}
		return &user, nil
	case http.StatusNotFound:
		return nil, errs.ErrUserNotFound
	default:
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
}

func (c *client) ChangePassword(ctx context.Context, userID int64, oldPassword string, password string) error {
	path := "/users/me/password"
	reqBody := struct {
		OldPassword string `json:"old_password"`
		Password    string `json:"password"`
	}{
		OldPassword: oldPassword,
		Password:    password,
	}
	resp, err := c.doAuthRequest(ctx, http.MethodPatch, path, reqBody, userID)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusForbidden:
		return errs.ErrWrongPassword
	case http.StatusNotFound:
		return errs.ErrUserNotFound
	default:
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
}

func (c *client) ResetPassword(ctx context.Context, userID int64) (string, error) {
	resp, err := c.doAuthRequest(ctx, http.MethodPost, "/users/me/password/reset", nil, userID)
	if err != nil {
		return "", err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	var respBody struct {
		Password string `json:"password"`
	}
	if err = json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return "", err
	}
	return respBody.Password, nil
}
