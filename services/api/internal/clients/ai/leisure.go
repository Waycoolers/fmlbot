package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Waycoolers/fmlbot/services/api/internal/domain"
)

func (c *Client) GetLeisureIdea(ctx context.Context, location, activity, budget, extra string) (string, error) {
	body := domain.LeisureIdeaRequest{
		Location:      location,
		ActivityLevel: activity,
		Budget:        budget,
		ExtraContext:  extra,
	}

	reqBody, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/v1/ai/leisure-idea", bytes.NewReader(reqBody))
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
