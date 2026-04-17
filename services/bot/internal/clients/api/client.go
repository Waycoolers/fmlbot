package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/Waycoolers/fmlbot/services/bot/internal/clients/auth"
	"github.com/Waycoolers/fmlbot/services/bot/internal/config"
	"github.com/Waycoolers/fmlbot/services/bot/internal/domain"
)

type client struct {
	baseURL    string
	httpClient *http.Client
	authClient domain.AuthClient
	tokenCache sync.Map
}

func New(cfg *config.Config) domain.ApiClient {
	apiUrl := fmt.Sprintf("http://%s:%d", cfg.Api.Host, cfg.Api.Port)
	apiTimeout := time.Duration(cfg.Api.HTTPTimeout) * time.Second
	c := &http.Client{
		Timeout: apiTimeout,
	}
	authClient := auth.New(cfg.Auth)
	return &client{
		baseURL:    apiUrl,
		httpClient: c,
		authClient: authClient,
	}
}

func (c *client) getToken(ctx context.Context, userID int64) (string, error) {
	cached, ok := c.tokenCache.Load(userID)
	if ok {
		return cached.(string), nil
	}
	token, err := c.authClient.GetAccessToken(ctx, userID)
	if err != nil {
		return "", err
	}
	c.tokenCache.Store(userID, token)
	return token, nil
}

func (c *client) doAuthRequest(ctx context.Context, method, path string, body any, chatID int64) (*http.Response, error) {
	execute := func(token string) (*http.Response, error) {
		var reqBody []byte
		if body != nil {
			var err error
			reqBody, err = json.Marshal(body)
			if err != nil {
				return nil, err
			}
		}
		url := c.baseURL + path
		req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(reqBody))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		return c.httpClient.Do(req)
	}

	token, err := c.getToken(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	resp, err := execute(token)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusUnauthorized {
		_ = resp.Body.Close()

		c.tokenCache.Delete(chatID)

		newToken, err := c.authClient.GetAccessToken(ctx, chatID)
		if err != nil {
			return nil, fmt.Errorf("failed to refresh token: %w", err)
		}
		c.tokenCache.Store(chatID, newToken)

		return execute(newToken)
	}

	return resp, nil
}
