package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Waycoolers/fmlbot/pkg/plans"
	"github.com/Waycoolers/fmlbot/services/bot/internal/domain"
	"github.com/Waycoolers/fmlbot/services/bot/internal/state"
	"github.com/Waycoolers/fmlbot/services/bot/internal/ui"
)

type planAPIStub struct {
	apiStub
	proposals []plans.Proposal
	current   plans.View
	changes   []plans.Change
}

func (s *planAPIStub) GetPlan(context.Context, int64, int64) (plans.View, error) {
	return s.current, nil
}

func (s *planAPIStub) ChangePlan(_ context.Context, _, _ int64, change plans.Change) (plans.View, error) {
	s.changes = append(s.changes, change)
	return s.current, nil
}

func (s *planAPIStub) CreatePlan(_ context.Context, _ int64, proposal plans.Proposal) (plans.View, error) {
	s.proposals = append(s.proposals, proposal)
	return plans.View{
		ViewerID: 1,
		Plan: plans.Plan{
			ID: 10, OwnerID: 1, PartnerID: proposal.Pair.PartnerID, ProposedBy: 1, Revision: 1,
			Status: plans.Pending, StartsAt: proposal.StartsAt, Card: proposal.Card,
		},
	}, nil
}

func planCardFixture(title string) plans.Card {
	return plans.Card{
		Title: title, DurationMinutes: 30, EstimatedCost: "500 ₽", Preparation: "Купить чай",
		Steps: []string{"Заварить чай", "Включить плейлист"}, Backup: "Взять какао",
	}
}

func TestPlanOfferUsesSelectedCardAndInlineMoscowTime(t *testing.T) {
	api := &planAPIStub{}
	bot := &planKeyboardStub{botStub: botStub{sent: map[int64][]string{}}}
	pair := plans.Pair{PartnerID: 42, Version: 11}
	draft := &ideaDraftStub{draft: &domain.IdeaDraft{
		Session: "0123456789abcdef",
		Step:    planCardsStep,
		Plan: &domain.PlanDraft{
			Pair: pair, Cards: []plans.Card{planCardFixture("Пикник"), planCardFixture("Чайная церемония")},
		},
	}}
	h := New(ui.New(bot), nil, nil, draft, api, state.New())
	ctx := context.Background()

	h.HandleCallback(ctx, &domain.CallbackQuery{ChatID: 1, Data: "plan:offer:old-session:0"})
	if len(bot.keyboard.Rows) != 0 {
		t.Fatal("an obsolete card opened the calendar")
	}
	h.HandleCallback(ctx, &domain.CallbackQuery{ChatID: 1, Data: "plan:offer:0123456789abcdef:1"})
	if len(bot.keyboard.Rows) == 0 || len(api.proposals) != 0 {
		t.Fatal("a card was offered without a time")
	}
	h.HandleMessage(ctx, &domain.Message{ChatID: 1, Text: "пятница после семи"})
	if len(api.proposals) != 0 {
		t.Fatal("text submitted a plan")
	}

	expected := choosePlanPickerTime(t, h, bot)
	if len(api.proposals) != 0 {
		t.Fatal("picking a time submitted a plan before confirmation")
	}
	send := clickPlanPickerButton(t, h, bot, ":send")
	h.HandleCallback(ctx, &domain.CallbackQuery{ChatID: 1, Data: send})
	if len(api.proposals) != 1 {
		t.Fatalf("expected one proposal, got %d", len(api.proposals))
	}
	proposal := api.proposals[0]
	if proposal.Pair != pair || proposal.DraftID != "0123456789abcdef" || proposal.Card.Title != "Чайная церемония" ||
		!proposal.StartsAt.Equal(expected) {
		t.Fatalf("wrong card, pair, or time: %+v", proposal)
	}
	if h.sm.GetStep(1) != state.Empty || draft.draft != nil {
		t.Fatal("successful offer did not clear its draft")
	}
}

func TestPlanCounterofferUsesInlineTimeAndOriginalRevision(t *testing.T) {
	api := &planAPIStub{current: plans.View{
		ViewerID: 1,
		Plan: plans.Plan{
			ID: 10, OwnerID: 2, PartnerID: 1, ProposedBy: 2, Revision: 7,
			Status: plans.Pending, StartsAt: time.Now().Add(time.Hour), Card: planCardFixture("Пикник"),
		},
	}}
	bot := &planKeyboardStub{botStub: botStub{sent: map[int64][]string{}}}
	h := New(ui.New(bot), nil, nil, &ideaDraftStub{}, api, state.New())
	h.HandleCallback(context.Background(), &domain.CallbackQuery{ChatID: 1, Data: "plan:reschedule:10:7"})
	expected := choosePlanPickerTime(t, h, bot)
	clickPlanPickerButton(t, h, bot, ":send")
	if len(api.changes) != 1 || api.changes[0].Action != plans.Reschedule || api.changes[0].Revision != 7 ||
		api.changes[0].StartsAt == nil || !api.changes[0].StartsAt.Equal(expected) {
		t.Fatalf("incorrect counteroffer: %+v", api.changes)
	}
}

func choosePlanPickerTime(t *testing.T, h *Handler, bot *planKeyboardStub) time.Time {
	t.Helper()
	moscow := time.FixedZone("MSK", 3*60*60)
	tomorrow := time.Now().In(moscow).AddDate(0, 0, 1)
	start := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 19, 17, 0, 0, moscow)
	if tomorrow.Month() != time.Now().In(moscow).Month() {
		nextMonth := time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, moscow)
		clickPlanPickerButton(t, h, bot, ":month:"+nextMonth.Format("200601021504"))
	}
	clickPlanPickerButton(t, h, bot, ":hours:"+start.Format("20060102")+"0000")
	clickPlanPickerButton(t, h, bot, ":minutes:"+start.Format("2006010215")+"00")
	clickPlanPickerButton(t, h, bot, ":exact:"+start.Format("2006010215")+"00")
	clickPlanPickerButton(t, h, bot, ":review:"+start.Format("200601021504"))
	return start
}

func clickPlanPickerButton(t *testing.T, h *Handler, bot *planKeyboardStub, suffix string) string {
	t.Helper()
	for _, row := range bot.keyboard.Rows {
		for _, button := range row.Buttons {
			if strings.HasPrefix(button.Data, "plan:time:") && strings.HasSuffix(button.Data, suffix) {
				h.HandleCallback(context.Background(), &domain.CallbackQuery{ChatID: 1, Data: button.Data})
				return button.Data
			}
		}
	}
	t.Fatalf("missing picker button ending in %s", suffix)
	return ""
}

type planKeyboardStub struct {
	botStub
	keyboard domain.InlineKeyboard
}

func (b *planKeyboardStub) SendWithInlineKeyboard(chatID int64, text string, keyboard domain.InlineKeyboard) error {
	b.keyboard = keyboard
	return b.botStub.SendWithInlineKeyboard(chatID, text, keyboard)
}

func TestPlanNoticeOffersActionsForCurrentRevision(t *testing.T) {
	bot := &planKeyboardStub{botStub: botStub{sent: map[int64][]string{}}}
	h := New(ui.New(bot), nil, nil, nil, nil, state.New())
	notice := plans.Notice{
		ChatID: 77,
		View: plans.View{
			ViewerID: 2,
			Plan: plans.Plan{
				ID: 5, OwnerID: 1, PartnerID: 2, ProposedBy: 1, Revision: 3, Status: plans.Pending,
				StartsAt: time.Now().Add(time.Hour), Card: planCardFixture("Чай <вместе>"),
			},
		},
	}
	body, err := json.Marshal(notice)
	if err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	h.NotifyPlan(response, httptest.NewRequest(http.MethodPost, "/updates/plans", bytes.NewReader(body)))
	messages := bot.sent[notice.ChatID]
	if response.Code != http.StatusOK || len(messages) != 1 ||
		!strings.Contains(messages[0], "другое время") || !strings.Contains(messages[0], "Чай &lt;вместе&gt;") {
		t.Fatalf("unexpected notice: status %d, messages %v", response.Code, messages)
	}
	var callbacks []string
	for _, row := range bot.keyboard.Rows {
		for _, button := range row.Buttons {
			callbacks = append(callbacks, button.Data)
		}
	}
	if strings.Join(callbacks, ",") != "plan:accept:5:3,plan:reschedule:5:3,plan:decline:5:3" {
		t.Fatalf("unexpected actions: %v", callbacks)
	}
}
