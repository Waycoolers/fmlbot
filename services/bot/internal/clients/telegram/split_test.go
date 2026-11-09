package telegram

import (
	"errors"
	"strings"
	"testing"

	"github.com/Waycoolers/fmlbot/services/bot/internal/domain"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func TestClassifySendError(t *testing.T) {
	for code, final := range map[int]bool{400: true, 403: true, 429: false, 502: false} {
		err := classifySendError(&tgbotapi.Error{Code: code, Message: "x"})
		if errors.Is(err, domain.ErrRecipientUnavailable) != final {
			t.Errorf("code %d: %v", code, err)
		}
	}
	if err := classifySendError(errors.New("timeout")); errors.Is(err, domain.ErrRecipientUnavailable) {
		t.Errorf("network error must stay temporary: %v", err)
	}
}

func TestSplitMessage(t *testing.T) {
	if parts := splitMessage("короткий текст", 100); len(parts) != 1 {
		t.Fatalf("got %d parts", len(parts))
	}

	text := strings.Repeat("👉 строка\n\n", 100)
	parts := splitMessage(text, 50)
	if len(parts) < 2 {
		t.Fatalf("got %d parts", len(parts))
	}
	for _, part := range parts {
		if len([]rune(part)) > 50 {
			t.Errorf("part is too long: %d", len([]rune(part)))
		}
		if strings.Contains(part, "строк\n") {
			t.Errorf("line was cut: %q", part)
		}
	}
	if got := strings.Count(strings.Join(parts, "\n"), "👉"); got != 100 {
		t.Errorf("lost lines: %d", got)
	}

	long := splitMessage(strings.Repeat("я", 120), 50)
	if len(long) != 3 {
		t.Errorf("text without line breaks: got %d parts", len(long))
	}
}
