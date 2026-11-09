package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Waycoolers/fmlbot/services/bot/internal/domain"
)

func (c *client) GetMyUserConfig(ctx context.Context, chatID int64) (*domain.UserConfig, error) {
	resp, err := c.doAuthRequest(ctx, http.MethodGet, "/user_config/me", nil, chatID)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	var cfg domain.UserConfig
	err = json.NewDecoder(resp.Body).Decode(&cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *client) GetPartnerUserConfig(ctx context.Context, chatID int64) (*domain.UserConfig, error) {
	resp, err := c.doAuthRequest(ctx, http.MethodGet, "/user_config/partner", nil, chatID)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	var cfg domain.UserConfig
	err = json.NewDecoder(resp.Body).Decode(&cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *client) UpdateUserConfig(ctx context.Context, chatID int64, maxCount *int) error {
	reqBody := struct {
		MaxComplimentCount *int `json:"max_compliment_count"`
	}{MaxComplimentCount: maxCount}
	resp, err := c.doAuthRequest(ctx, http.MethodPatch, "/user_config/me", reqBody, chatID)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	return nil
}
