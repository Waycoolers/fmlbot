package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Waycoolers/fmlbot/common/errs"
	"github.com/Waycoolers/fmlbot/services/bot/internal/domain"
)

func (c *client) CreateUser(ctx context.Context, chatID int64, username string) error {
	reqBody := struct {
		Username string `json:"username"`
	}{
		Username: username,
	}

	resp, err := c.doAuthRequest(ctx, http.MethodPost, "/users", reqBody, chatID)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	switch resp.StatusCode {
	case http.StatusCreated:
		return nil
	case http.StatusConflict:
		return errs.ErrUserExists
	default:
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
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

func (c *client) UpdateMe(ctx context.Context, chatID int64, username string, partnerID int64) error {
	reqBody := struct {
		Username  string `json:"username"`
		PartnerID int64  `json:"partnerID"`
	}{
		Username:  username,
		PartnerID: partnerID,
	}
	resp, err := c.doAuthRequest(ctx, http.MethodPut, "/users/me", reqBody, chatID)
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
		return errs.ErrUserNotFound
	default:
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}

func (c *client) UpdatePartner(ctx context.Context, requesterID int64, username string, partnerID int64) error {
	reqBody := struct {
		Username  string `json:"username"`
		PartnerID int64  `json:"partner_id"`
	}{
		Username:  username,
		PartnerID: partnerID,
	}
	resp, err := c.doAuthRequest(ctx, http.MethodPut, "/users/partner", reqBody, requesterID)
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
		decodeErr := json.NewDecoder(resp.Body).Decode(&errBody)
		if decodeErr != nil {
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

func (c *client) PairUsers(ctx context.Context, requesterID int64, partnerID int64) error {
	reqBody := struct {
		PartnerID int64 `json:"partner_id"`
	}{
		PartnerID: partnerID,
	}
	resp, err := c.doAuthRequest(ctx, http.MethodPost, "/users/pair", reqBody, requesterID)
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
		return errs.ErrPartnerNotFound
	default:
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
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
	path := fmt.Sprintf("/users/by-username/%s", username)
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
