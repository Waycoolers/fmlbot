package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Waycoolers/fmlbot/pkg/errs"
	"github.com/Waycoolers/fmlbot/services/bot/internal/domain"
	"github.com/Waycoolers/fmlbot/services/bot/internal/state"
	"github.com/Waycoolers/fmlbot/services/bot/internal/ui"
)

type botStub struct {
	domain.BotClient
	sent     map[int64][]string
	failWith map[int64]error
}

func (b *botStub) SendMessage(chatID int64, text string) error {
	if err := b.failWith[chatID]; err != nil {
		return err
	}
	b.sent[chatID] = append(b.sent[chatID], text)
	return nil
}
func (b *botStub) AnswerCallback(string) error { return nil }
func (b *botStub) SendKeyboard(chatID int64, text string, _ domain.Keyboard) (domain.Message, error) {
	b.sent[chatID] = append(b.sent[chatID], text)
	return domain.Message{}, nil
}
func (b *botStub) DeleteMessage(int64, int) error { return nil }
func (b *botStub) SendWithInlineKeyboard(chatID int64, text string, _ domain.InlineKeyboard) error {
	b.sent[chatID] = append(b.sent[chatID], text)
	return nil
}
func (b *botStub) DeleteMessageReplyMarkup(int64, int) error { return nil }

type apiStub struct {
	domain.ApiClient
	partnerID   int64
	deleted     bool
	received    string
	receiveErr  error
	pairVersion int64
	unpaired    []int64
	oldSeen     string
	newSeen     string
	reset       bool
	fromUser    int64
	answerErr   error
}

func (a *apiStub) AcceptPartnerRequest(context.Context, int64, int64) (int64, error) {
	return a.fromUser, a.answerErr
}
func (a *apiStub) DeclinePartnerRequest(context.Context, int64, int64) (int64, error) {
	return a.fromUser, a.answerErr
}
func (a *apiStub) GetPartner(context.Context, int64) (*domain.User, error) {
	return &domain.User{UserID: a.fromUser, Username: "from"}, nil
}

func (a *apiStub) ChangePassword(_ context.Context, _ int64, oldPassword string, password string) error {
	a.oldSeen, a.newSeen = oldPassword, password
	return nil
}
func (a *apiStub) ResetPassword(context.Context, int64) (string, error) {
	a.reset = true
	return "a<b", nil
}

func (a *apiStub) GetMe(context.Context, int64) (*domain.User, error) {
	return &domain.User{UserID: 1, PartnerID: a.partnerID, PairVersion: a.pairVersion}, nil
}
func (a *apiStub) TelegramChatID(_ context.Context, id int64) (int64, error) {
	return id, nil
}

func (a *apiStub) Unpair(_ context.Context, _ int64, version int64) error {
	a.unpaired = append(a.unpaired, version)
	return nil
}

func (a *apiStub) DeleteMe(context.Context, int64) error { a.deleted = true; return nil }
func (a *apiStub) UpdateUserConfig(context.Context, int64, *int) error {
	return nil
}
func (a *apiStub) ReceiveNextCompliment(context.Context, int64) (*domain.Compliment, error) {
	return &domain.Compliment{Text: a.received, AuthorID: a.partnerID}, a.receiveErr
}
func (a *apiStub) GetAllCompliments(context.Context, int64) ([]domain.Compliment, error) {
	return []domain.Compliment{{Text: a.received, IsSent: true}}, nil
}

func setup(api *apiStub) (*Handler, *botStub) {
	bot := &botStub{sent: map[int64][]string{}}
	return New(ui.New(bot), nil, nil, nil, api, state.New()), bot
}

type registrationStub struct {
	apiStub
	names []string
}

func (a *registrationStub) CreateUser(_ context.Context, _ int64, name string) (string, error) {
	a.names = append(a.names, name)
	if len(a.names) == 1 {
		return "", errs.ErrUsernameIsAlreadyTaken
	}
	return "test-password", nil
}

func TestRegistrationAllowsDifferentNameAfterConflict(t *testing.T) {
	api := &registrationStub{}
	bot := &botStub{sent: map[int64][]string{}}
	h := New(ui.New(bot), nil, nil, nil, api, state.New())
	ctx := context.Background()
	h.registerTelegram(ctx, &domain.Message{ChatID: 1}, "taken")
	if h.sm.GetStep(1) != state.AwaitingRegistrationUsername {
		t.Fatal("registration cannot recover from username conflict")
	}
	h.HandleMessage(ctx, &domain.Message{ChatID: 1, Text: "new_name"})
	if len(api.names) != 2 || api.names[1] != "new_name" || h.sm.GetStep(1) != state.Empty {
		t.Fatalf("alternative username was not registered: %v", api.names)
	}
}

func TestDeleteAccountWithPartner(t *testing.T) {
	api := &apiStub{partnerID: 42}
	h, bot := setup(api)
	h.HandleDeleteAccount(context.Background(), &domain.CallbackQuery{ChatID: 1, MessageID: 1, Data: "account:delete:confirm"})
	if api.deleted {
		t.Fatal("old button deleted account")
	}
	h.accountDeleteConfirmations.Store(int64(1), "current")
	h.HandleDeleteAccount(context.Background(), &domain.CallbackQuery{ChatID: 1, MessageID: 2, Data: "account:delete:confirm:current"})
	if !api.deleted {
		t.Fatal("account was not deleted")
	}
	if len(bot.sent[42]) != 1 {
		t.Fatalf("partner messages: %q", bot.sent[42])
	}
}

func TestComplimentTextIsEscaped(t *testing.T) {
	api := &apiStub{partnerID: 42, received: "ты <3 & я"}
	h, bot := setup(api)
	h.ReceiveCompliment(context.Background(), &domain.Message{ChatID: 1})
	for _, chatID := range []int64{42} {
		if len(bot.sent[chatID]) != 1 || !strings.Contains(bot.sent[chatID][0], "ты &lt;3 &amp; я") {
			t.Errorf("chat %d: %q", chatID, bot.sent[chatID])
		}
	}

	h.GetCompliments(context.Background(), &domain.Message{ChatID: 7})
	if !strings.Contains(bot.sent[7][0], "ты &lt;3 &amp; я") {
		t.Errorf("list: %q", bot.sent[7])
	}
}

func TestFrequencyResetsState(t *testing.T) {
	h, _ := setup(&apiStub{})
	h.sm.SetStep(1, state.AwaitingComplimentFrequency)
	h.ProcessComplimentFrequency(context.Background(), &domain.Message{ChatID: 1, Text: "3"})
	if step := h.sm.GetStep(1); step != state.Empty {
		t.Fatalf("step %q", step)
	}
}

func TestChangePasswordAsksOldPasswordFirst(t *testing.T) {
	api := &apiStub{}
	h, _ := setup(api)
	ctx := context.Background()

	h.ChangePassword(ctx, &domain.Message{ChatID: 1})
	h.HandleMessage(ctx, &domain.Message{ChatID: 1, Text: "OldPassw0rd"})
	h.HandleMessage(ctx, &domain.Message{ChatID: 1, Text: "NewPassw0rd"})

	if api.oldSeen != "OldPassw0rd" || api.newSeen != "NewPassw0rd" {
		t.Fatalf("old=%q new=%q", api.oldSeen, api.newSeen)
	}
	if step := h.sm.GetStep(1); step != state.Empty {
		t.Fatalf("step %q", step)
	}
	if _, ok := h.oldPasswords.Load(int64(1)); ok {
		t.Fatal("old password must not stay in memory")
	}
}

func TestForgotPassword(t *testing.T) {
	api := &apiStub{}
	h, bot := setup(api)
	ctx := context.Background()

	stale := &domain.CallbackQuery{ChatID: 1, MessageID: 1, Data: "account:password:forgot"}
	h.HandleForgotPassword(ctx, stale)
	if api.reset {
		t.Fatal("stale button must not reset the password")
	}

	h.ChangePassword(ctx, &domain.Message{ChatID: 1})
	h.HandleForgotPassword(ctx, stale)
	if !api.reset {
		t.Fatal("password was not reset")
	}
	last := bot.sent[1][len(bot.sent[1])-1]
	if !strings.Contains(last, "a&lt;b") {
		t.Fatalf("got %q", last)
	}
}

func TestPartnerRequestAnswers(t *testing.T) {
	ctx := context.Background()
	cq := &domain.CallbackQuery{ChatID: 1, MessageID: 1, UserName: "target"}

	api := &apiStub{fromUser: 2}
	h, bot := setup(api)
	h.HandlePartnerRequest(ctx, cq, "accept:5")
	if len(bot.sent[1]) != 1 || len(bot.sent[2]) != 1 || !strings.Contains(bot.sent[2][0], "@target") {
		t.Fatalf("accept: %q %q", bot.sent[1], bot.sent[2])
	}

	h, bot = setup(api)
	h.HandlePartnerRequest(ctx, cq, "decline:5")
	if len(bot.sent[2]) != 1 || !strings.Contains(bot.sent[2][0], "отклонил") {
		t.Fatalf("decline: %q", bot.sent[2])
	}

	h, bot = setup(&apiStub{answerErr: errs.ErrPartnerRequestNotFound})
	h.HandlePartnerRequest(ctx, cq, "accept:5")
	if len(bot.sent[1]) != 1 || !strings.Contains(bot.sent[1][0], "неактуальна") {
		t.Fatalf("stale: %q", bot.sent[1])
	}
}

func TestNotificationStatus(t *testing.T) {
	cases := []struct {
		name string
		fail error
		want int
	}{
		{"delivered", nil, http.StatusOK},
		{"blocked bot is final", domain.ErrRecipientUnavailable, http.StatusGone},
		{"temporary failure", errors.New("telegram is down"), http.StatusServiceUnavailable},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			h, bot := setup(&apiStub{})
			bot.failWith = map[int64]error{2: c.fail}
			body := `{"important_date_id":1,"user_ids":[2],"message":"x"}`
			w := httptest.NewRecorder()
			h.NotifyAllImportantDates(w, httptest.NewRequest(http.MethodPost, "/updates/important_dates", strings.NewReader(body)))
			if w.Code != c.want {
				t.Errorf("got %d, want %d", w.Code, c.want)
			}
		})
	}
}

func TestDeletePartnerRejectsOldConfirmation(t *testing.T) {
	api := &apiStub{partnerID: 42, pairVersion: 12}
	h, _ := setup(api)
	for _, data := range []string{"partner:delete:confirm", "partner:delete:confirm:41:11", "partner:delete:confirm:42:11"} {
		h.HandleDeletePartner(context.Background(), &domain.CallbackQuery{ChatID: 1, Data: data})
	}
	if len(api.unpaired) != 0 {
		t.Fatalf("stale confirmations called API: %v", api.unpaired)
	}
	h.HandleDeletePartner(context.Background(), &domain.CallbackQuery{ChatID: 1, Data: "partner:delete:confirm:42:12"})
	if len(api.unpaired) != 1 || api.unpaired[0] != 12 {
		t.Fatalf("current confirmation did not call API with expected version: %v", api.unpaired)
	}
}

func TestComplimentDeliveryFailureDoesNotNotifyAuthor(t *testing.T) {
	h, bot := setup(&apiStub{partnerID: 42, receiveErr: errs.ErrComplimentDelivery})
	h.ReceiveCompliment(context.Background(), &domain.Message{ChatID: 1})
	if len(bot.sent[42]) != 0 || len(bot.sent[1]) != 1 || !strings.Contains(bot.sent[1][0], "ещё раз") {
		t.Fatalf("unexpected messages: %v", bot.sent)
	}
}

type ideaDraftStub struct {
	draft *domain.IdeaDraft
}

func (s *ideaDraftStub) Get(context.Context, int64) (*domain.IdeaDraft, error) { return s.draft, nil }
func (s *ideaDraftStub) Save(_ context.Context, _ int64, draft *domain.IdeaDraft) error {
	s.draft = draft
	return nil
}
func (s *ideaDraftStub) Delete(context.Context, int64) error { s.draft = nil; return nil }

func TestOldIdeaButtonsCannotChangeCurrentDraft(t *testing.T) {
	h, _ := setup(&apiStub{})
	store := &ideaDraftStub{draft: &domain.IdeaDraft{Session: "current", Step: "loc"}}
	h.ideaDrafts = store
	for _, callback := range []string{"idea:bud:previous:0", "idea:cancel:previous:loc", "idea:bud:current:0"} {
		h.HandleCallback(context.Background(), &domain.CallbackQuery{ChatID: 1, Data: callback})
	}
	if store.draft == nil || store.draft.Session != "current" || store.draft.Step != "loc" {
		t.Fatalf("draft was modified: %+v", store.draft)
	}
}

func TestFormatMarkdownToHTML(t *testing.T) {
	got := formatMarkdownToHTML("**План** <3\n* пункт & ещё")
	want := "<b>План</b> &lt;3\n• пункт &amp; ещё"
	if got != want {
		t.Fatalf("got %q", got)
	}
}

func TestNextOccurrence(t *testing.T) {
	loc := time.UTC
	date := time.Date(2020, 5, 25, 0, 0, 0, 0, loc)

	onTheDay := time.Date(2026, 5, 25, 15, 0, 0, 0, loc)
	if got := nextOccurrence(date, onTheDay); got.Year() != 2026 {
		t.Errorf("event day: %v", got)
	}

	after := time.Date(2026, 5, 26, 0, 0, 0, 0, loc)
	if got := nextOccurrence(date, after); got.Year() != 2027 {
		t.Errorf("after event: %v", got)
	}

	leap := time.Date(2024, 2, 29, 0, 0, 0, 0, loc)
	got := nextOccurrence(leap, time.Date(2026, 1, 1, 0, 0, 0, 0, loc))
	if got.Month() != time.February || got.Day() != 28 {
		t.Errorf("29 february: %v", got)
	}
}
