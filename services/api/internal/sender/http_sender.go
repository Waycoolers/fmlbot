package sender

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Waycoolers/fmlbot/services/api/internal/domain"
)

type HTTPSender struct {
	botURL string
	client *http.Client
	secret []byte
}

func NewHTTPSender(botURL string, secret []byte) domain.Sender {
	return &HTTPSender{
		botURL: botURL,
		client: &http.Client{},
		secret: secret,
	}
}

func (s *HTTPSender) SendMessage(ctx context.Context, path string, update any) error {
	data, err := json.Marshal(update)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.botURL+path, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Internal-Secret", string(s.secret))
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}
