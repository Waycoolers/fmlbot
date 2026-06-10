package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Waycoolers/fmlbot/services/bot/internal/domain"
)

func (c *client) GetLeisureIdea(ctx context.Context, chatID int64, location, level, budget, context string) (string, error) {
	body := domain.LeisureIdeaRequest{
		Location:      location,
		ActivityLevel: level,
		Budget:        budget,
		ExtraContext:  context,
	}
	resp, err := c.doAuthRequest(ctx, http.MethodPost, "/ideas/leisure", body, chatID)
	if err != nil {
		return "", err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	var idea struct {
		Idea string `json:"idea"`
	}
	err = json.NewDecoder(resp.Body).Decode(&idea)
	if err != nil {
		return "", err
	}
	return idea.Idea, nil
}
