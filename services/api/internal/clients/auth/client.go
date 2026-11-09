package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Waycoolers/fmlbot/services/api/internal/config"
)

type Client struct {
	baseURL string
	client  *http.Client
	secret  string
}

func New(cfg *config.AuthConfig, internalSecret []byte) *Client {
	return &Client{
		baseURL: fmt.Sprintf("http://%s:%d", cfg.Host, cfg.Port),
		client:  &http.Client{Timeout: time.Duration(cfg.HTTPTimeout) * time.Second},
		secret:  string(internalSecret),
	}
}

func (c *Client) RevokeAllTokens(ctx context.Context, userID int64) error {
	reqBody, err := json.Marshal(struct {
		UserID int64 `json:"user_id"`
	}{UserID: userID})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/internal/auth/revoke-all", bytes.NewReader(reqBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Internal-Secret", c.secret)
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	return nil
}
