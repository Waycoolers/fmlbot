package telegram

import (
	"errors"
	"strings"
	"testing"

	"github.com/Waycoolers/fmlbot/services/bot/internal/domain"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func TestClassifySendError(t *testing.T) {
	cases := []struct {
		code    int
		message string
		final   bool
	}{
		{400, "Bad Request: chat not found", true},
		{400, "Bad Request: can't parse entities", false},
		{403, "Forbidden: bot was blocked by the user", true},
		{429, "Too Many Requests", false},
		{502, "Bad Gateway", false},
	}
	for _, c := range cases {
		err := classifySendError(&tgbotapi.Error{Code: c.code, Message: c.message})
		if errors.Is(err, domain.ErrRecipientUnavailable) != c.final {
			t.Errorf("code %d: %v", c.code, err)
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

func TestLongHTMLMessageDoesNotSplitTags(t *testing.T) {
	want := strings.Repeat("я", 4500) + " <3"
	parts, mode := messageParts("<b>" + strings.Repeat("я", 4500) + " &lt;3</b>")
	if mode != "" || len(parts) != 2 || strings.Join(parts, "") != want {
		t.Fatalf("invalid split (mode=%q, parts=%d)", mode, len(parts))
	}
}
