package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Waycoolers/fmlbot/pkg/plans"
)

func (c *Client) GetPlanOptions(ctx context.Context, prompt plans.Prompt) (plans.Options, error) {
	body, err := json.Marshal(prompt)
	if err != nil {
		return plans.Options{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/ai/plan-options", bytes.NewReader(body))
	if err != nil {
		return plans.Options{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Internal-Secret", c.secret)

	resp, err := c.client.Do(req)
	if err != nil {
		return plans.Options{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return plans.Options{}, fmt.Errorf("plan options returned status %d", resp.StatusCode)
	}
	var options plans.Options
	if err = json.NewDecoder(io.LimitReader(resp.Body, 64<<10)).Decode(&options); err != nil {
		return plans.Options{}, err
	}
	return options, nil
}
