package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Waycoolers/fmlbot/services/bot/internal/config"
	"github.com/Waycoolers/fmlbot/services/bot/internal/domain"
)

type client struct {
	baseURL    string
	httpClient *http.Client
}

type tokenRequest struct {
	UserID int64 `json:"user_id"`
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
}

func New(cfg *config.AuthConfig) domain.AuthClient {
	url := fmt.Sprintf("http://%s:%d", cfg.Host, cfg.Port)
	timeout := time.Duration(cfg.HTTPTimeout) * time.Second
	c := &http.Client{
		Timeout: timeout,
	}
	return &client{
		baseURL:    url,
		httpClient: c,
	}
}

func (c *client) GetAccessToken(ctx context.Context, userID int64) (string, error) {
	reqBody := tokenRequest{UserID: userID}
	data, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("%s/auth/token", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("auth service returned %d", resp.StatusCode)
	}

	var tokenResp tokenResponse
	err = json.NewDecoder(resp.Body).Decode(&tokenResp)
	if err != nil {
		return "", err
	}
	return tokenResp.AccessToken, nil
}
