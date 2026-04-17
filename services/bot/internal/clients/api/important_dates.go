package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Waycoolers/fmlbot/services/bot/internal/domain"
)

func (c *client) AddImportantDate(ctx context.Context, chatID int64, req domain.ImportantDateRequest) (*domain.ImportantDate, error) {
	resp, err := c.doAuthRequest(ctx, http.MethodPost, "/important_dates", req, chatID)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	var date domain.ImportantDate
	if err := json.NewDecoder(resp.Body).Decode(&date); err != nil {
		return nil, err
	}
	return &date, nil
}

func (c *client) GetImportantDate(ctx context.Context, chatID int64, dateID int64) (*domain.ImportantDate, error) {
	path := fmt.Sprintf("/important_dates/%d", dateID)
	resp, err := c.doAuthRequest(ctx, http.MethodGet, path, nil, chatID)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	var date domain.ImportantDate
	if err := json.NewDecoder(resp.Body).Decode(&date); err != nil {
		return nil, err
	}
	return &date, nil
}

func (c *client) GetAllImportantDates(ctx context.Context, chatID int64) ([]domain.ImportantDate, error) {
	resp, err := c.doAuthRequest(ctx, http.MethodGet, "/important_dates", nil, chatID)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	var dates []domain.ImportantDate
	if err := json.NewDecoder(resp.Body).Decode(&dates); err != nil {
		return nil, err
	}
	return dates, nil
}

func (c *client) UpdateImportantDate(ctx context.Context, chatID int64, dateID int64, req domain.ImportantDateRequest) error {
	path := fmt.Sprintf("/important_dates/%d", dateID)
	resp, err := c.doAuthRequest(ctx, http.MethodPatch, path, req, chatID)
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

func (c *client) UpdateImportantDateSharing(ctx context.Context, chatID int64, dateID int64, makeShared bool) error {
	path := fmt.Sprintf("/important_dates/%d/sharing", dateID)
	reqBody := struct {
		MakeShared bool `json:"make_shared"`
	}{MakeShared: makeShared}
	resp, err := c.doAuthRequest(ctx, http.MethodPatch, path, reqBody, chatID)
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

func (c *client) DeleteImportantDate(ctx context.Context, chatID int64, dateID int64) error {
	path := fmt.Sprintf("/important_dates/%d", dateID)
	resp, err := c.doAuthRequest(ctx, http.MethodDelete, path, nil, chatID)
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
