package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Waycoolers/fmlbot/pkg/errs"
)

func (c *client) doInternalRequest(ctx context.Context, method, path string, body any) (*http.Response, error) {
	var data []byte
	if body != nil {
		var err error
		data, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Internal-Secret", c.secret)
	return c.httpClient.Do(req)
}

func (c *client) resolveTelegramID(ctx context.Context, kind string, id int64, key string) (int64, error) {
	path := fmt.Sprintf("/auth/telegram/%s/%d", kind, id)
	resp, err := c.doInternalRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return 0, errs.ErrUserNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("Telegram lookup returned status %d", resp.StatusCode)
	}
	var result map[string]int64
	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}
	if result[key] <= 0 {
		return 0, fmt.Errorf("Telegram lookup omitted %s", key)
	}
	return result[key], nil
}

func (c *client) TelegramChatID(ctx context.Context, userID int64) (int64, error) {
	return c.resolveTelegramID(ctx, "chat", userID, "telegram_id")
}

func (c *client) LinkTelegram(ctx context.Context, chatID int64, code string) error {
	body := map[string]any{"telegram_id": chatID, "code": code}
	resp, err := c.doInternalRequest(ctx, http.MethodPost, "/auth/telegram/link", body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusNoContent:
		c.tokenCache.Delete(chatID)
		return nil
	case http.StatusBadRequest:
		return errs.ErrTelegramLinkInvalid
	case http.StatusConflict:
		return errs.ErrTelegramAlreadyLinked
	default:
		return fmt.Errorf("Telegram linking returned status %d", resp.StatusCode)
	}
}
