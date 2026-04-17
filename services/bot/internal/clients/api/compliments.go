package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Waycoolers/fmlbot/common/errs"
	"github.com/Waycoolers/fmlbot/services/bot/internal/domain"
)

func (c *client) AddCompliment(ctx context.Context, chatID int64, text string) (*domain.Compliment, error) {
	reqBody := struct {
		Text string `json:"text"`
	}{Text: text}
	resp, err := c.doAuthRequest(ctx, http.MethodPost, "/compliments", reqBody, chatID)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	var comp domain.Compliment
	err = json.NewDecoder(resp.Body).Decode(&comp)
	if err != nil {
		return nil, err
	}
	return &comp, nil
}

func (c *client) GetAllCompliments(ctx context.Context, chatID int64) ([]domain.Compliment, error) {
	resp, err := c.doAuthRequest(ctx, http.MethodGet, "/compliments", nil, chatID)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	var comps []domain.Compliment
	err = json.NewDecoder(resp.Body).Decode(&comps)
	if err != nil {
		return nil, err
	}
	return comps, nil
}

func (c *client) UpdateCompliment(ctx context.Context, chatID int64, complimentID int64, text string, isSent bool) error {
	reqBody := struct {
		Text   string `json:"text"`
		IsSent bool   `json:"is_sent"`
	}{Text: text, IsSent: isSent}
	path := fmt.Sprintf("/compliments/%d", complimentID)
	resp, err := c.doAuthRequest(ctx, http.MethodPut, path, reqBody, chatID)
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
		return errs.ErrComplimentNotFound
	default:
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
}

func (c *client) DeleteCompliment(ctx context.Context, chatID int64, complimentID int64) error {
	path := fmt.Sprintf("/compliments/%d", complimentID)
	resp, err := c.doAuthRequest(ctx, http.MethodDelete, path, nil, chatID)
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
		return errs.ErrComplimentNotFound
	default:
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
}

func (c *client) ReceiveNextCompliment(ctx context.Context, chatID int64) (*domain.Compliment, error) {
	resp, err := c.doAuthRequest(ctx, http.MethodPost, "/compliments/next", nil, chatID)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	switch resp.StatusCode {
	case http.StatusOK:
		var comp domain.Compliment
		err = json.NewDecoder(resp.Body).Decode(&comp)
		if err != nil {
			return nil, err
		}
		return &comp, nil
	case http.StatusNotFound:
		return nil, errs.ErrUserNotFound
	case http.StatusGone:
		return nil, errs.ErrNoCompliments
	case http.StatusTooManyRequests:
		var errBody struct {
			Error   string `json:"error"`
			Minutes int    `json:"minutes"`
		}
		decodeErr := json.NewDecoder(resp.Body).Decode(&errBody)
		if decodeErr != nil {
			return nil, errs.ErrLimitExceeded
		}
		return nil, &errs.ErrBucketEmpty{Minutes: errBody.Minutes}
	default:
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
}
