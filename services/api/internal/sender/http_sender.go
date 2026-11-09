package sender

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Waycoolers/fmlbot/pkg/errs"
	"github.com/Waycoolers/fmlbot/pkg/plans"
	"github.com/Waycoolers/fmlbot/services/api/internal/domain"
)

type HTTPSender struct {
	botURL string
	client *http.Client
	secret []byte
	users  domain.UsersRepo
}

func NewHTTPSender(botURL string, secret []byte, users domain.UsersRepo) *HTTPSender {
	return &HTTPSender{
		botURL: botURL,
		client: &http.Client{Timeout: 30 * time.Second},
		secret: secret,
		users:  users,
	}
}

func (s *HTTPSender) SendPlanNotice(ctx context.Context, notice plans.Notice) error {
	return s.send(ctx, "/updates/plans", notice)
}

func (s *HTTPSender) SendMessage(ctx context.Context, update domain.MessageRequest) error {
	chatID, err := s.telegramChat(ctx, update.UserID)
	if err != nil {
		return err
	}
	update.UserID = chatID
	return s.send(ctx, "/updates/message", update)
}

func (s *HTTPSender) SendImportantDatesNotification(ctx context.Context, update domain.ImportantDateMessage) error {
	if len(update.UserIDs) != 1 {
		return fmt.Errorf("expected one Telegram recipient")
	}
	chatID, err := s.telegramChat(ctx, update.UserIDs[0])
	if err != nil {
		return err
	}
	update.UserIDs = []int64{chatID}
	return s.send(ctx, "/updates/important_dates", update)
}

func (s *HTTPSender) telegramChat(ctx context.Context, userID int64) (int64, error) {
	chatID, err := s.users.GetTelegramID(ctx, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, errs.ErrRecipientUnavailable
	}
	return chatID, err
}

func (s *HTTPSender) send(ctx context.Context, path string, update any) error {
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

	if resp.StatusCode == http.StatusGone {
		return errs.ErrRecipientUnavailable
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}
