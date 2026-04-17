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

func (c *client) ResetMyUserConfig(ctx context.Context, chatID int64) error {
	resp, err := c.doAuthRequest(ctx, http.MethodPost, "/user_config/reset/me", nil, chatID)
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
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
}

func (c *client) ResetPartnerUserConfig(ctx context.Context, chatID int64) error {
	resp, err := c.doAuthRequest(ctx, http.MethodPost, "/user_config/reset/partner", nil, chatID)
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
