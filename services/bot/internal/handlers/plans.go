package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Waycoolers/fmlbot/pkg/errs"
	"github.com/Waycoolers/fmlbot/pkg/plans"
	"github.com/Waycoolers/fmlbot/services/bot/internal/domain"
	"github.com/Waycoolers/fmlbot/services/bot/internal/state"
	"github.com/Waycoolers/fmlbot/services/bot/internal/ui"
)

const (
	planBriefStep   = "plan_brief"
	planCardsStep   = "plan_cards"
	planOfferStep   = "plan_offer_time"
	planCounterStep = "plan_counter_time"
)

func (h *Handler) startPlan(ctx context.Context, msg *domain.Message) {
	chatID := msg.ChatID
	if h.planIdeaBusy(chatID) {
		return
	}
	user, err := h.api.GetMe(ctx, chatID)
	if err != nil || user == nil {
		h.replyPlanError(chatID, err)
		return
	}
	if user.PartnerID == 0 {
		h.replyPlanError(chatID, errs.ErrPartnerNotFound)
		return
	}
	draft := &domain.IdeaDraft{
		Session: wizardSession(),
		Step:    planBriefStep,
		Plan:    &domain.PlanDraft{Pair: plans.Pair{PartnerID: user.PartnerID, Version: user.PairVersion}},
	}
	if err = h.ideaDrafts.Save(ctx, chatID, draft); err != nil {
		h.replyPlanError(chatID, err)
		return
	}
	h.sm.SetStep(chatID, state.AwaitingPlanBrief)
	h.Reply(chatID, "💞 <b>План для двоих</b>\n\nРасскажи, чего хочется. Можно указать место, бюджет, сколько у вас времени и настроение.\n\n"+
		"Например, «В пятницу после семи, дома, до 2 000 ₽. У нас полтора часа, хочется чего-то уютного».\n\n"+
		"Подберу идеи с помощью AI. Партнёру отправится выбранная тобой карточка со временем.")
}

func (h *Handler) processPlanBrief(ctx context.Context, msg *domain.Message) {
	chatID := msg.ChatID
	brief := strings.TrimSpace(msg.Text)
	if !plans.ValidBrief(brief) {
		h.Reply(chatID, fmt.Sprintf("Опиши пожелания одним сообщением до %d символов.", plans.MaxBriefLength))
		return
	}
	draft, err := h.ideaDrafts.Get(ctx, chatID)
	if err != nil {
		h.replyPlanError(chatID, err)
		return
	}
	if draft == nil || draft.Plan == nil || draft.Step != planBriefStep {
		h.sm.SetStep(chatID, state.Empty)
		h.Reply(chatID, "Черновик истёк. Открой «План для двоих» заново.")
		return
	}
	draft.Plan.Brief = brief
	if err = h.ideaDrafts.Save(ctx, chatID, draft); err != nil {
		h.replyPlanError(chatID, err)
		return
	}
	h.generatePlanCards(ctx, chatID, draft, plans.Prompt{Brief: brief})
}

func (h *Handler) generatePlanCards(ctx context.Context, chatID int64, draft *domain.IdeaDraft, prompt plans.Prompt) {
	if !h.startIdeaJob(chatID, func() {
		h.Reply(chatID, "⏳ Подбираю два варианта для вас…")
		options, err := h.api.GetPlanOptions(ctx, chatID, plans.OptionsRequest{Pair: draft.Plan.Pair, Prompt: prompt})
		if ctx.Err() != nil {
			return
		}
		switch {
		case errors.Is(err, errs.ErrPairChanged), errors.Is(err, errs.ErrPartnerNotFound), errors.Is(err, errs.ErrLimitExceeded):
			h.replyPlanError(chatID, err)
			return
		case err != nil || !options.Valid():
			slog.Warn("Could not generate plan cards", "error", err)
			h.replyPlanGenerationRetry(chatID, draft)
			return
		}
		next := &domain.IdeaDraft{
			Session: wizardSession(),
			Step:    planCardsStep,
			Plan:    &domain.PlanDraft{Pair: draft.Plan.Pair, Brief: prompt.Brief, Cards: options.Cards},
		}
		if err = h.ideaDrafts.Save(ctx, chatID, next); err != nil {
			h.replyPlanGenerationRetry(chatID, draft)
			return
		}
		if err = h.ui.SendPlanOptions(chatID, next.Session, options.Cards); err != nil {
			slog.Warn("Could not show generated plan cards", "error", err)
			h.Reply(chatID, "Не удалось показать все карточки. Открой «План для двоих» ещё раз.")
		}
	}) {
		h.Reply(chatID, "Сейчас генератор занят. Попробуй ещё раз чуть позже.")
		return
	}
	h.sm.SetStep(chatID, state.Empty)
}

func (h *Handler) replyPlanGenerationRetry(chatID int64, draft *domain.IdeaDraft) {
	if draft.Step != planBriefStep {
		h.Reply(chatID, "Не удалось доработать варианты. Попробуй нажать кнопку ещё раз.")
		return
	}
	keyboard := domain.InlineKeyboard{Rows: []domain.InlineKeyboardRow{
		{Buttons: []domain.InlineKeyboardButton{{Text: "Повторить", Data: "plan:retry:" + draft.Session}}},
	}}
	if err := h.ui.Client.SendWithInlineKeyboard(chatID, "Не удалось подобрать планы. Твои пожелания сохранены в черновике.", keyboard); err != nil {
		slog.Warn("Could not show plan retry button", "error", err)
	}
}

func (h *Handler) handlePlan(ctx context.Context, cq *domain.CallbackQuery, action, payload string) {
	switch action {
	case "offer", "simpler", "cheaper", "more_involved", "more_expensive", "retry":
		h.handlePlanDraft(ctx, cq, action, payload)
	case "time":
		h.handlePlanPicker(ctx, cq, payload)
	case "view":
		id, err := strconv.ParseInt(payload, 10, 64)
		if err != nil || id <= 0 {
			h.ReplyUnknownCallback(ctx, cq)
			return
		}
		view, err := h.api.GetPlan(ctx, cq.ChatID, id)
		if err != nil {
			h.replyPlanError(cq.ChatID, err)
			return
		}
		h.showPlan(cq.ChatID, view)
	case string(plans.Accept), string(plans.Reschedule), string(plans.Decline), string(plans.Cancel):
		h.handlePlanChange(ctx, cq, plans.Action(action), payload)
	default:
		h.ReplyUnknownCallback(ctx, cq)
	}
}

func (h *Handler) handlePlanDraft(ctx context.Context, cq *domain.CallbackQuery, action, payload string) {
	chatID := cq.ChatID
	if h.planIdeaBusy(chatID) {
		return
	}
	session, index, hasIndex := strings.Cut(payload, ":")
	draft, err := h.ideaDrafts.Get(ctx, chatID)
	if err != nil {
		h.replyPlanError(chatID, err)
		return
	}
	if draft == nil || draft.Plan == nil || draft.Session != session {
		h.ui.RemoveButtons(chatID, cq.MessageID)
		h.Reply(chatID, "Эта карточка устарела. Открой «План для двоих» заново.")
		return
	}
	if action == "retry" && !hasIndex && draft.Step == planBriefStep && plans.ValidBrief(draft.Plan.Brief) {
		h.generatePlanCards(ctx, chatID, draft, plans.Prompt{Brief: draft.Plan.Brief})
		return
	}
	selected, err := strconv.Atoi(index)
	if !hasIndex || err != nil || selected < 0 || selected >= len(draft.Plan.Cards) ||
		(draft.Step != planCardsStep && draft.Step != planOfferStep) {
		h.ReplyUnknownCallback(ctx, cq)
		return
	}
	card := draft.Plan.Cards[selected]
	if action == "offer" {
		draft.Step = planOfferStep
		draft.Plan.Selected = selected
		h.askPlanTime(ctx, chatID, draft, "Когда предложить партнёру «"+html.EscapeString(card.Title)+"»?")
		return
	}
	h.generatePlanCards(ctx, chatID, draft, plans.Prompt{Brief: draft.Plan.Brief, Adjustment: action, Reference: &card})
}

func (h *Handler) handlePlanChange(ctx context.Context, cq *domain.CallbackQuery, action plans.Action, payload string) {
	chatID := cq.ChatID
	idText, revisionText, found := strings.Cut(payload, ":")
	id, idErr := strconv.ParseInt(idText, 10, 64)
	revision, revErr := strconv.ParseInt(revisionText, 10, 64)
	if !found || idErr != nil || revErr != nil || id <= 0 || revision <= 0 {
		h.ReplyUnknownCallback(ctx, cq)
		return
	}
	if action == plans.Reschedule {
		if h.planIdeaBusy(chatID) {
			return
		}
		view, err := h.api.GetPlan(ctx, chatID, id)
		if err != nil {
			h.replyPlanError(chatID, err)
			return
		}
		if view.Revision != revision || !view.StartsAt.After(time.Now()) || !view.Can(view.ViewerID, action) {
			h.ui.RemoveButtons(chatID, cq.MessageID)
			h.replyPlanError(chatID, errs.ErrPlanChanged)
			return
		}
		draft := &domain.IdeaDraft{Step: planCounterStep, Plan: &domain.PlanDraft{TargetID: id, TargetRevision: revision}}
		h.askPlanTime(ctx, chatID, draft, "На какое время предложить перенести «"+html.EscapeString(view.Card.Title)+"»?")
		return
	}
	view, err := h.api.ChangePlan(ctx, chatID, id, plans.Change{Action: action, Revision: revision})
	if err != nil {
		if isStalePlan(err) {
			h.ui.RemoveButtons(chatID, cq.MessageID)
		}
		h.replyPlanError(chatID, err)
		return
	}
	h.ui.RemoveButtons(chatID, cq.MessageID)
	h.showPlan(chatID, view)
}

func (h *Handler) askPlanTime(ctx context.Context, chatID int64, draft *domain.IdeaDraft, question string) {
	if err := h.showPlanPicker(ctx, chatID, draft, "month", time.Time{}, question); err != nil {
		h.replyPlanError(chatID, err)
		return
	}
	h.sm.SetStep(chatID, state.Empty)
}

func (h *Handler) showPlanPicker(ctx context.Context, chatID int64, draft *domain.IdeaDraft, screen string, at time.Time, question string) error {
	session := wizardSession()
	if err := h.ui.SendPlanPicker(chatID, session, screen, at, time.Now(), question); err != nil {
		return err
	}
	draft.Plan.PickerSession = session
	draft.Plan.PickerTime = time.Time{}
	if screen == "review" {
		draft.Plan.PickerTime = at
	}
	return h.ideaDrafts.Save(ctx, chatID, draft)
}

func (h *Handler) handlePlanPicker(ctx context.Context, cq *domain.CallbackQuery, payload string) {
	chatID := cq.ChatID
	if h.planIdeaBusy(chatID) {
		return
	}
	session, rest, found := strings.Cut(payload, ":")
	if !found || session == "" {
		h.ReplyUnknownCallback(ctx, cq)
		return
	}
	draft, err := h.ideaDrafts.Get(ctx, chatID)
	if err != nil {
		h.replyPlanError(chatID, err)
		return
	}
	if draft == nil || draft.Plan == nil || draft.Plan.PickerSession != session ||
		(draft.Step != planOfferStep && draft.Step != planCounterStep) {
		h.ui.RemoveButtons(chatID, cq.MessageID)
		h.Reply(chatID, "Этот календарь устарел. Выбери время на последней карточке или открой план заново.")
		return
	}
	if rest == "cancel" {
		if draft.Step == planOfferStep {
			draft.Step = planCardsStep
			draft.Plan.PickerSession = ""
			draft.Plan.PickerTime = time.Time{}
			err = h.ideaDrafts.Save(ctx, chatID, draft)
		} else {
			err = h.ideaDrafts.Delete(ctx, chatID)
		}
		if err != nil {
			h.replyPlanError(chatID, err)
			return
		}
		h.ui.RemoveButtons(chatID, cq.MessageID)
		h.Reply(chatID, "Выбор времени отменён.")
		return
	}
	if rest == "send" {
		startsAt := draft.Plan.PickerTime
		if startsAt.IsZero() {
			h.ReplyUnknownCallback(ctx, cq)
			return
		}
		h.submitPlanTime(ctx, cq, draft, startsAt)
		return
	}
	screen, value, found := strings.Cut(rest, ":")
	at, err := ui.ParsePlanPickerTime(value)
	if !found || err != nil {
		h.ReplyUnknownCallback(ctx, cq)
		return
	}
	switch screen {
	case "month", "hours", "minutes", "exact", "review":
	default:
		h.ReplyUnknownCallback(ctx, cq)
		return
	}
	if screen == "review" && !plans.ValidStart(at, time.Now()) {
		h.replyPlanError(chatID, errs.ErrPlanTime)
		screen = "hours"
	}
	if err = h.showPlanPicker(ctx, chatID, draft, screen, at, ""); err != nil {
		h.replyPlanError(chatID, err)
		return
	}
	h.ui.RemoveButtons(chatID, cq.MessageID)
}

func (h *Handler) submitPlanTime(ctx context.Context, cq *domain.CallbackQuery, draft *domain.IdeaDraft, startsAt time.Time) {
	chatID := cq.ChatID
	var view plans.View
	var err error
	if !plans.ValidStart(startsAt, time.Now()) {
		err = errs.ErrPlanTime
	}
	if err == nil && draft.Step == planOfferStep {
		selected := draft.Plan.Selected
		if selected < 0 || selected >= len(draft.Plan.Cards) {
			h.Reply(chatID, "Выбери карточку плана заново.")
			return
		}
		view, err = h.api.CreatePlan(ctx, chatID, plans.Proposal{
			Pair: draft.Plan.Pair, DraftID: draft.Session, StartsAt: startsAt, Card: draft.Plan.Cards[selected],
		})
	} else if err == nil {
		view, err = h.api.ChangePlan(ctx, chatID, draft.Plan.TargetID, plans.Change{
			Action: plans.Reschedule, Revision: draft.Plan.TargetRevision, StartsAt: &startsAt,
		})
	}
	if err != nil {
		if isStalePlan(err) {
			h.ui.RemoveButtons(chatID, cq.MessageID)
		}
		h.replyPlanError(chatID, err)
		if errors.Is(err, errs.ErrPlanTime) {
			if err := h.showPlanPicker(ctx, chatID, draft, "month", startsAt, ""); err == nil {
				h.ui.RemoveButtons(chatID, cq.MessageID)
			} else {
				h.replyPlanError(chatID, err)
			}
		}
		return
	}
	h.ui.RemoveButtons(chatID, cq.MessageID)
	if err = h.ideaDrafts.Delete(ctx, chatID); err != nil {
		slog.Warn("Could not clear plan draft", "error", err)
	}
	h.showPlan(chatID, view)
}

func (h *Handler) showMyPlans(ctx context.Context, msg *domain.Message) {
	items, err := h.api.ListPlans(ctx, msg.ChatID)
	if err == nil {
		err = h.ui.SendPlansList(msg.ChatID, items)
	}
	if err != nil {
		h.replyPlanError(msg.ChatID, err)
	}
}

func (h *Handler) showPlan(chatID int64, view plans.View) {
	if err := h.ui.SendPlan(chatID, view, false); err != nil {
		h.replyPlanError(chatID, err)
	}
}

func (h *Handler) planIdeaBusy(chatID int64) bool {
	_, busy := h.ideaInFlight.Load(chatID)
	if busy {
		h.Reply(chatID, "⏳ Идеи уже готовятся.")
	}
	return busy
}

func isStalePlan(err error) bool {
	return errors.Is(err, errs.ErrPlanNotFound) || errors.Is(err, errs.ErrPlanChanged) ||
		errors.Is(err, errs.ErrPairChanged) || errors.Is(err, errs.ErrForbidden)
}

func (h *Handler) replyPlanError(chatID int64, err error) {
	switch {
	case errors.Is(err, errs.ErrUserNotFound):
		h.Reply(chatID, "Сначала открой /start и войди в FMLBot.")
	case errors.Is(err, errs.ErrPartnerNotFound):
		h.Reply(chatID, "🤍 Сначала добавь партнёра в меню «Мой партнёр».")
	case errors.Is(err, errs.ErrPlanPartnerTelegram):
		h.Reply(chatID, "У партнёра пока не привязан Telegram. Когда он подключит аккаунт, нажми «Отправить» ещё раз.")
	case errors.Is(err, errs.ErrPairChanged):
		h.Reply(chatID, "Пара изменилась. Открой «План для двоих» заново.")
	case isStalePlan(err):
		h.Reply(chatID, "Предложение уже изменилось или недоступно. Открой «Мои планы».")
	case errors.Is(err, errs.ErrPlanTime):
		h.Reply(chatID, "Выбери другое время в календаре, в пределах года.")
	case errors.Is(err, errs.ErrLimitExceeded):
		h.Reply(chatID, "Слишком много запросов. Попробуй чуть позже.")
	default:
		h.HandleErr(chatID, "Plan operation failed", err)
	}
}

func (h *Handler) NotifyPlan(w http.ResponseWriter, r *http.Request) {
	var notice plans.Notice
	if err := json.NewDecoder(r.Body).Decode(&notice); err != nil || notice.ChatID <= 0 ||
		(notice.View.ViewerID != notice.View.OwnerID && notice.View.ViewerID != notice.View.PartnerID) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := h.ui.SendPlan(notice.ChatID, notice.View, notice.Reminder); err != nil {
		if errors.Is(err, domain.ErrRecipientUnavailable) {
			w.WriteHeader(http.StatusGone)
			return
		}
		slog.Warn("Could not deliver plan card", "error", err)
		w.WriteHeader(http.StatusServiceUnavailable)
	}
}
