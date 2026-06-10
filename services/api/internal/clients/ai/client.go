package ai

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Waycoolers/fmlbot/services/api/internal/config"
)

type Client struct {
	baseURL string
	client  *http.Client
}

func New(cfg *config.AIConfig) *Client {
	aiUrl := fmt.Sprintf("http://%s:%d", cfg.Host, cfg.Port)
	client := &http.Client{
		Timeout: time.Duration(cfg.HTTPTimeout) * time.Second,
	}
	return &Client{
		baseURL: aiUrl,
		client:  client,
	}
}
