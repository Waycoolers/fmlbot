package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"

	"github.com/Waycoolers/fmlbot/pkg/errs"
	"log/slog"
	"strings"
	"sync"

	"github.com/Waycoolers/fmlbot/services/bot/internal/domain"
	"github.com/Waycoolers/fmlbot/services/bot/internal/redis_store"
	"github.com/Waycoolers/fmlbot/services/bot/internal/state"
	"github.com/Waycoolers/fmlbot/services/bot/internal/ui"
)

type Handler struct {
	ui                      *ui.MenuUI
	importantDateDrafts     *redis_store.ImportantDateDraftStore
	importantDateEditDrafts *redis_store.ImportantDateEditDraftStore
	ideaDrafts              ideaDraftStore
	ideaJobs                chan struct{}
	ideaInFlight            sync.Map
	api                     domain.ApiClient
	sm                      *state.Machine
	// Текущий пароль на время диалога смены пароля. Только в памяти, на диск не попадает
	oldPasswords               sync.Map
	accountDeleteConfirmations sync.Map
}

func New(ui *ui.MenuUI, importantDateDrafts *redis_store.ImportantDateDraftStore,
	importantDateEditDrafts *redis_store.ImportantDateEditDraftStore,
	ideaDrafts ideaDraftStore,
	api domain.ApiClient, sm *state.Machine) *Handler {
	return &Handler{
		ui:                      ui,
		importantDateDrafts:     importantDateDrafts,
		importantDateEditDrafts: importantDateEditDrafts,
		ideaDrafts:              ideaDrafts,
		ideaJobs:                make(chan struct{}, 8),
		api:                     api,
		sm:                      sm,
	}
}

func wizardSession() string {
	var token [8]byte
	_, _ = rand.Read(token[:])
	return hex.EncodeToString(token[:])
}

func (h *Handler) HandleMessage(ctx context.Context, msg *domain.Message) {
	chatID := msg.ChatID
	text := msg.Text

	if fields := strings.Fields(text); len(fields) > 0 && fields[0] == "/link" {
		if len(fields) != 2 {
			h.Reply(chatID, "Отправь /link и одноразовый код из своего аккаунта.")
			return
		}
		h.LinkTelegram(ctx, msg, fields[1])
		return
	}

	// Если введена команда, то сбрасываем state
	for _, command := range domain.Commands {
		if text == string(command) {
			h.sm.SetStep(chatID, state.Empty)
			h.oldPasswords.Delete(chatID)
		}
	}

	slog.Debug("Client wrote", "chatID", chatID)

	if text == string(domain.Start) {
		h.ShowStartMenu(ctx, chatID)
		return
	} else if text == string(domain.Register) {
		h.Register(ctx, msg)
		return
	}

	step := h.sm.GetStep(chatID)

	if !(step == state.Empty) {
		switch step {
		case state.AwaitingRegistrationUsername:
			h.RegisterWithUsername(ctx, msg)
		case state.AwaitingPartner:
			h.ProcessPartnerUsername(ctx, msg)
		case state.AwaitingCompliment:
			h.ProcessCompliment(ctx, msg)
		case state.AwaitingComplimentFrequency:
			h.ProcessComplimentFrequency(ctx, msg)
		case state.AwaitingTitleImportantDate:
			h.HandleTitleImportantDate(ctx, msg)
		case state.AwaitingEditTitleImportantDate:
			h.HandleEditTitleImportantDateText(ctx, msg)
		case state.AwaitingOldPassword:
			h.HandleOldPassword(ctx, msg)
		case state.AwaitingPassword:
			h.HandleChangePassword(ctx, msg)
		case state.AwaitingPlanBrief:
			h.processPlanBrief(ctx, msg)
		default:
			h.ReplyUnknownMessage(ctx, msg)
		}
	} else {
		switch text {
		case string(domain.Main):
			h.ShowMainMenu(ctx, msg)
		case string(domain.Account):
			h.ShowAccountMenu(ctx, msg)
		case string(domain.Partner):
			h.ShowPartnerMenu(ctx, msg)
		case string(domain.Compliments):
			h.ShowComplimentsMenu(ctx, msg)
		case string(domain.ImportantDates):
			h.ShowImportantDatesMenu(ctx, msg)
		case string(domain.DeleteAccount):
			h.DeleteAccount(ctx, msg)
		case string(domain.ChangePassword):
			h.ChangePassword(ctx, msg)
		case string(domain.AddPartner):
			h.SetPartner(ctx, msg)
		case string(domain.DeletePartner):
			h.DeletePartner(ctx, msg)
		case string(domain.AddCompliment):
			h.AddCompliment(ctx, msg)
		case string(domain.DeleteCompliment):
			h.DeleteCompliment(ctx, msg)
		case string(domain.GetCompliments):
			h.GetCompliments(ctx, msg)
		case string(domain.ReceiveCompliment):
			h.ReceiveCompliment(ctx, msg)
		case string(domain.EditComplimentFrequency):
			h.EditComplimentFrequency(ctx, msg)
		case string(domain.AddImportantDate):
			h.AddImportantDate(ctx, msg)
		case string(domain.GetImportantDates):
			h.GetImportantDates(ctx, msg)
		case string(domain.DeleteImportantDate):
			h.DeleteImportantDate(ctx, msg)
		case string(domain.EditImportantDate):
			h.EditImportantDate(ctx, msg)
		case string(domain.Ideas):
			h.ShowIdeasMenu(ctx, msg)
		case string(domain.GenerateLeisureIdea):
			h.GenerateLeisureIdea(ctx, msg)
		case string(domain.PlanForTwo):
			h.startPlan(ctx, msg)
		case string(domain.MyPlans):
			h.showMyPlans(ctx, msg)
		default:
			h.ReplyUnknownMessage(ctx, msg)
		}
	}
}

func (h *Handler) HandleCallback(ctx context.Context, cq *domain.CallbackQuery) {
	data := cq.Data

	// Без ответа Telegram продолжает показывать загрузку на кнопке
	if err := h.ui.Client.AnswerCallback(cq.ID); err != nil {
		slog.Warn("Error answering callback", "error", err)
	}

	section, action, payload := parseCallbackData(data)

	switch section {
	case "account":
		h.handleAccount(ctx, cq, action, payload)
	case "partner":
		h.handlePartner(ctx, cq, action, payload)
	case "compliments":
		h.handleCompliments(ctx, cq, action, payload)
	case "important_dates":
		h.handleImportantDates(ctx, cq, action, payload)
	case "idea":
		h.handleIdea(ctx, cq, action, payload)
	case "plan":
		h.handlePlan(ctx, cq, action, payload)
	default:
		h.ReplyUnknownCallback(ctx, cq)
	}
}

func (h *Handler) handleAccount(ctx context.Context, cq *domain.CallbackQuery, action string, payload string) {
	switch action {
	case "delete":
		if strings.HasPrefix(payload, "confirm") || strings.HasPrefix(payload, "cancel") {
			h.HandleDeleteAccount(ctx, cq)
		}
	case "password":
		if payload == "forgot" {
			h.HandleForgotPassword(ctx, cq)
		}
	default:
		h.ReplyUnknownCallback(ctx, cq)
	}
}

func (h *Handler) handlePartner(ctx context.Context, cq *domain.CallbackQuery, action string, payload string) {
	switch action {
	case "delete":
		if strings.HasPrefix(payload, "confirm") || strings.HasPrefix(payload, "cancel") {
			h.HandleDeletePartner(ctx, cq)
		}
	case "request":
		h.HandlePartnerRequest(ctx, cq, payload)
	default:
		h.ReplyUnknownCallback(ctx, cq)
	}
}

func (h *Handler) handleCompliments(ctx context.Context, cq *domain.CallbackQuery, action string, payload string) {
	switch action {
	case "delete":
		if strings.HasPrefix(payload, "confirm") || strings.HasPrefix(payload, "cancel") {
			h.HandleDeleteCompliment(ctx, cq)
		}
	default:
		h.ReplyUnknownCallback(ctx, cq)
	}
}

func (h *Handler) handleImportantDates(ctx context.Context, cq *domain.CallbackQuery, action string, payload string) {
	switch action {
	case "add":
		payloadEnd := strings.LastIndexByte(payload, ':')
		if payloadEnd < 0 {
			h.ReplyUnknownCallback(ctx, cq)
			return
		}
		draft, err := h.importantDateDrafts.Get(ctx, cq.ChatID)
		if err != nil {
			h.HandleErr(cq.ChatID, "Error loading date draft", err)
			return
		}
		if draft == nil || draft.Session == "" || draft.Session != payload[payloadEnd+1:] ||
			!matchesDateDraftStep(h.sm.GetStep(cq.ChatID), draft, payload[:payloadEnd]) {
			h.ui.RemoveButtons(cq.ChatID, cq.MessageID)
			h.Reply(cq.ChatID, "Эта кнопка устарела. Создай дату заново.")
			return
		}
		payload = payload[:payloadEnd]
		callback := *cq
		callback.Data = "important_dates:add:" + payload
		callback.WizardSession = draft.Session
		cq = &callback
		switch {
		case strings.HasPrefix(payload, "partner"):
			h.HandlePartnerImportantDate(ctx, cq)
		case strings.HasPrefix(payload, "notify_before"):
			h.HandleNotifyBeforeImportantDate(ctx, cq)
		case strings.HasPrefix(payload, "year"):
			h.HandleYearImportantDateUniversal(ctx, cq)
		case strings.HasPrefix(payload, "month"):
			h.HandleMonthImportantDateUniversal(ctx, cq)
		case strings.HasPrefix(payload, "day"):
			h.HandleDayImportantDateUniversal(ctx, cq)
		default:
			h.ReplyUnknownCallback(ctx, cq)
		}
	case "delete":
		if strings.HasPrefix(payload, "confirm") || strings.HasPrefix(payload, "cancel") {
			h.HandleDeleteImportantDate(ctx, cq)
		}
	case "update_menu":
		h.HandleEditImportantDate(ctx, cq)
	case "update":
		switch {
		case strings.HasPrefix(payload, "title"):
			h.HandleEditTitleImportantDate(ctx, cq)
		case strings.HasPrefix(payload, "date"):
			h.HandleEditDateImportantDate(ctx, cq)
		case strings.HasPrefix(payload, "partner"):
			h.HandleEditPartnerImportantDate(ctx, cq)
		case strings.HasPrefix(payload, "notify_before"):
			h.HandleEditNotifyBeforeImportantDate(ctx, cq)
		case strings.HasPrefix(payload, "is_active"):
			h.HandleEditIsActiveImportantDate(ctx, cq)
		case strings.HasPrefix(payload, "cancel"):
			h.CancelCallbackImportantDate(ctx, cq)
		default:
			h.ReplyUnknownCallback(ctx, cq)
		}
	case "edit":
		payloadEnd := strings.LastIndexByte(payload, ':')
		if payloadEnd < 0 {
			h.ReplyUnknownCallback(ctx, cq)
			return
		}
		draft, err := h.importantDateEditDrafts.Get(ctx, cq.ChatID)
		if err != nil {
			h.HandleErr(cq.ChatID, "Error loading date edit draft", err)
			return
		}
		if draft == nil || draft.Session == "" || draft.Session != payload[payloadEnd+1:] ||
			!matchesDateEditStep(draft, payload[:payloadEnd]) {
			h.ui.RemoveButtons(cq.ChatID, cq.MessageID)
			h.Reply(cq.ChatID, "Эта кнопка устарела. Открой настройки даты заново.")
			return
		}
		payload = payload[:payloadEnd]
		callback := *cq
		callback.Data = "important_dates:edit:" + payload
		callback.WizardSession = draft.Session
		cq = &callback
		switch {
		case strings.HasPrefix(payload, "year"):
			h.HandleYearImportantDateUniversal(ctx, cq)
		case strings.HasPrefix(payload, "month"):
			h.HandleMonthImportantDateUniversal(ctx, cq)
		case strings.HasPrefix(payload, "day"):
			h.HandleDayImportantDateUniversal(ctx, cq)
		case strings.HasPrefix(payload, "partner"):
			h.HandleEditPartnerImportantDateSelect(ctx, cq)
		case strings.HasPrefix(payload, "notify_before"):
			h.HandleEditNotifyBeforeImportantDateSelect(ctx, cq)
		default:
			h.ReplyUnknownCallback(ctx, cq)
		}
	default:
		h.ReplyUnknownCallback(ctx, cq)
	}
}

func matchesDateDraftStep(step state.State, draft *domain.ImportantDateDraft, action string) bool {
	switch {
	case strings.HasPrefix(action, "year:"):
		return step == state.AwaitingDateImportantDate && draft.Year == 0
	case strings.HasPrefix(action, "month:"):
		return step == state.AwaitingDateImportantDate && draft.Year != 0 && draft.Month == 0
	case strings.HasPrefix(action, "day:"):
		return step == state.AwaitingDateImportantDate && draft.Month != 0 && draft.Day == 0
	case strings.HasPrefix(action, "partner:"):
		return step == state.AwaitingPartnerImportantDate
	case strings.HasPrefix(action, "notify_before:"):
		return step == state.AwaitingNotifyBeforeImportantDate
	default:
		return false
	}
}

func matchesDateEditStep(draft *domain.ImportantDateEditDraft, payload string) bool {
	action, _, _ := strings.Cut(payload, ":")
	switch action {
	case "year":
		return draft.Action == "date" && draft.Year == 0
	case "month":
		return draft.Action == "date" && draft.Year != 0 && draft.Month == 0
	case "day":
		return draft.Action == "date" && draft.Month != 0
	default:
		return action == draft.Action
	}
}

func (h *Handler) ShowStartMenu(_ context.Context, chatID int64) {
	text := "✨ Чтобы разбудить бота, зарегистрируйся по кнопке ниже"
	err := h.ui.StartMenu(chatID, text)
	if err != nil {
		h.HandleErr(chatID, "An error occurred while trying to display the start menu", err)
		return
	}
}

func (h *Handler) ShowMainMenu(_ context.Context, msg *domain.Message) {
	chatID := msg.ChatID
	msgText := msg.Text
	text := "🌿 Выбери, что хочешь сделать"

	if msgText == string(domain.Register) {
		text = "fmlbot приветствует тебя! 💖"
	}

	err := h.ui.MainMenu(chatID, text)
	if err != nil {
		h.HandleErr(chatID, "Error while trying to display the main menu", err)
		return
	}
}

func (h *Handler) ReplyToUser(ctx context.Context, userID int64, text string) {
	chatID, err := h.api.TelegramChatID(ctx, userID)
	if errors.Is(err, errs.ErrUserNotFound) {
		return
	}
	if err != nil {
		slog.Error("Error resolving Telegram recipient", "error", err)
		return
	}
	h.Reply(chatID, text)
}

func (h *Handler) Reply(chatID int64, text string) {
	err := h.ui.Client.SendMessage(chatID, text)
	if err != nil {
		slog.Error("Error sending message", "error", err)
	}
}

func (h *Handler) ReplyUnknownCallback(_ context.Context, cq *domain.CallbackQuery) {
	chatID := cq.ChatID
	h.Reply(chatID, "🤍 Лучше воспользуйся кнопками - так будет проще")
}

func (h *Handler) ReplyUnknownMessage(_ context.Context, msg *domain.Message) {
	chatID := msg.ChatID
	h.Reply(chatID, "🤔 Я пока не понимаю это сообщение\nПопробуй выбрать действие с кнопок ниже")
}

func (h *Handler) HandleErr(chatID int64, msg string, err error) {
	h.Reply(chatID, "😔 Что-то пошло не так\nЯ уже стараюсь всё исправить")
	slog.Error("Error with message", "chatID", chatID, "message", msg, "error", err)
}

func (h *Handler) HandleUnknownError(chatID int64, err error) {
	h.Reply(chatID, "😔 Что-то пошло не так\nЯ уже стараюсь всё исправить")
	slog.Error("Unknown error", "chatID", chatID, "error", err)
}

func parseCallbackData(data string) (section, action, payload string) {
	parts := strings.Split(data, ":")
	switch len(parts) {
	case 1:
		return parts[0], "", ""
	case 2:
		return parts[0], parts[1], ""
	default:
		return parts[0], parts[1], strings.Join(parts[2:], ":")
	}
}
