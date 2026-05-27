package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/Waycoolers/fmlbot/pkg/errs"
	"github.com/Waycoolers/fmlbot/services/auth/internal/config"
	"github.com/Waycoolers/fmlbot/services/auth/internal/domain"
)

type client struct {
	baseURL    string
	httpClient *http.Client
	secret     []byte
}

func New(cfg *config.APIConfig, internalSecret []byte) domain.APIClient {
	url := fmt.Sprintf("http://%s:%d", cfg.Host, cfg.Port)
	httpClient := &http.Client{
		Timeout: cfg.HTTPTimeout,
	}
	return &client{
		baseURL:    url,
		httpClient: httpClient,
		secret:     internalSecret,
	}
}

func (c *client) VerifyUser(ctx context.Context, username string, password string) (int64, error) {
	body := struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{
		Username: username,
		Password: password,
	}
	reqBody, err := json.Marshal(body)
	if err != nil {
		return 0, err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/auth/verify", bytes.NewBuffer(reqBody))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Internal-Secret", string(c.secret))
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusUnauthorized {
			return 0, errs.ErrWrongPassword
		}
		if resp.StatusCode == http.StatusNotFound {
			return 0, errs.ErrUserNotFound
		}
		if resp.StatusCode == http.StatusBadRequest {
			return 0, errs.ErrBadRequest
		}
		return 0, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	var respBody struct {
		UserID int64 `json:"user_id"`
	}
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		return 0, err
	}
	return respBody.UserID, nil
}

func (c *client) IsUsernameAvailable(ctx context.Context, username string) (bool, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+fmt.Sprintf("/users/by-username/%s", username), nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("X-Internal-Secret", string(c.secret))
	resp, err := c.httpClient.Do(req)
	if errors.Is(err, errs.ErrUserNotFound) {
		return true, nil
	}
	if err == nil && resp.StatusCode == http.StatusOK {
		defer func(Body io.ReadCloser) {
			_ = Body.Close()
		}(resp.Body)
		return false, errs.ErrUsernameIsAlreadyTaken
	}
	return false, err
}

func (c *client) Register(ctx context.Context, userID int64, username string, password string) error {
	body := struct {
		UserID   int64  `json:"user_id"`
		Username string `json:"username"`
		Password string `json:"password"`
	}{
		UserID:   userID,
		Username: username,
		Password: password,
	}
	reqBody, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/auth/register", bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Internal-Secret", string(c.secret))
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	return nil
}
