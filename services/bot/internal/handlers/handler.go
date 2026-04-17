package handlers

import (
	"context"
	"log/slog"
	"strings"

	"github.com/Waycoolers/fmlbot/services/bot/internal/domain"
	"github.com/Waycoolers/fmlbot/services/bot/internal/redis_store"
	"github.com/Waycoolers/fmlbot/services/bot/internal/state"
	"github.com/Waycoolers/fmlbot/services/bot/internal/ui"
)

type Handler struct {
	ui                      *ui.MenuUI
	importantDateDrafts     *redis_store.ImportantDateDraftStore
	importantDateEditDrafts *redis_store.ImportantDateEditDraftStore
	api                     domain.ApiClient
	sm                      *state.Machine
}

func New(ui *ui.MenuUI, importantDateDrafts *redis_store.ImportantDateDraftStore,
	importantDateEditDrafts *redis_store.ImportantDateEditDraftStore,
	api domain.ApiClient, sm *state.Machine) *Handler {
	return &Handler{
		ui:                      ui,
		importantDateDrafts:     importantDateDrafts,
		importantDateEditDrafts: importantDateEditDrafts,
		api:                     api,
		sm:                      sm,
	}
}

func (h *Handler) HandleMessage(ctx context.Context, msg *domain.Message) {
	chatID := msg.ChatID
	text := msg.Text

	// Если введена команда, то сбрасываем state
	for _, command := range domain.Commands {
		if text == string(command) {
			h.sm.SetStep(state.Empty)
		}
	}

	slog.Info("Client wrote", "chatID", chatID, "text", text)

	if text == string(domain.Start) {
		h.ShowStartMenu(ctx, chatID)
		return
	} else if text == string(domain.Register) {
		h.Register(ctx, msg)
		return
	}

	step := h.sm.GetStep()

	if !(step == state.Empty) {
		switch step {
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
		default:
			h.ReplyUnknownMessage(ctx, msg)
		}
	}
}

func (h *Handler) HandleCallback(ctx context.Context, cq *domain.CallbackQuery) {
	data := cq.Data

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
