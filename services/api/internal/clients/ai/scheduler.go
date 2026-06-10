package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func (c *Client) GetDateIdea(ctx context.Context, title string, days int, context string) (text string, err error) {
	body := struct {
		EventType string `json:"event_type"`
		DaysUntil int    `json:"days_until"`
		Context   string `json:"context"`
	}{
		EventType: title,
		DaysUntil: days,
		Context:   context,
	}
	reqBody, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/v1/ai/date-idea", bytes.NewReader(reqBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	var respBody struct {
		Idea string `json:"idea"`
	}
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		return "", err
	}
	return respBody.Idea, nil
}
