package app

import (
	"context"
	"log"
	"strings"

	"github.com/Waycoolers/fmlbot/internal/domain"
	"github.com/Waycoolers/fmlbot/internal/handlers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

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

type Router struct {
	h *handlers.Handler
}

func NewRouter(h *handlers.Handler) *Router {
	return &Router{h: h}
}

func (r *Router) HandleUpdate(ctx context.Context, update tgbotapi.Update) {
	if update.CallbackQuery != nil {
		r.handleCallback(ctx, update.CallbackQuery)
		return
	}

	if update.Message != nil {
		r.handleMessage(ctx, update.Message)
		return
	}
}

func (r *Router) handleMessage(ctx context.Context, msg *tgbotapi.Message) {
	userID := msg.From.ID
	chatID := msg.Chat.ID
	text := msg.Text

	commands := []string{
		string(domain.Start),
		string(domain.Main),
		string(domain.Register),
		string(domain.Account),
		string(domain.Partner),
		string(domain.Compliments),
		string(domain.ImportantDates),
		string(domain.Register),
		string(domain.DeleteAccount),
		string(domain.AddPartner),
		string(domain.DeletePartner),
		string(domain.AddCompliment),
		string(domain.DeleteCompliment),
		string(domain.GetCompliments),
		string(domain.ReceiveCompliment),
		string(domain.EditComplimentFrequency),
		string(domain.AddImportantDate),
	}

	// Если введена команда, то сбрасываем state
	for _, command := range commands {
		if text == command {
			err := r.h.Store.SetUserState(ctx, userID, domain.Empty)
			if err != nil {
				r.h.HandleErr(chatID, "Ошибка при сбросе состояния", err)
				return
			}
		}
	}

	username, err := r.h.Store.GetUsername(ctx, userID)
	if err != nil {
		r.h.HandleErr(chatID, "Ошибка при получении юзернейма", err)
		return
	}
	log.Printf("Клиент %v написал: %v", username, text)

	if text == string(domain.Start) {
		r.h.ShowStartMenu(ctx, chatID)
		return
	} else if text == string(domain.Register) {
		r.h.Register(ctx, msg)
		r.h.ShowMainMenu(ctx, msg)
		return
	}

	state, err := r.h.Store.GetUserState(ctx, userID)
	if err != nil {
		r.h.HandleErr(chatID, "Ошибка при получении состояния", err)
		return
	}

	if !(state == domain.Empty) {
		switch state {
		case domain.AwaitingPartner:
			r.h.ProcessPartnerUsername(ctx, msg)
		case domain.AwaitingCompliment:
			r.h.ProcessCompliment(ctx, msg)
		case domain.AwaitingComplimentFrequency:
			r.h.ProcessComplimentFrequency(ctx, msg)
		case domain.AwaitingTitleImportantDate:
			r.h.HandleTitleImportantDate(ctx, msg)
		default:
			r.h.ReplyUnknownMessage(ctx, msg)
		}
	} else {
		switch text {
		case string(domain.Main):
			r.h.ShowMainMenu(ctx, msg)
		case string(domain.Account):
			r.h.ShowAccountMenu(ctx, msg)
		case string(domain.Partner):
			r.h.ShowPartnerMenu(ctx, msg)
		case string(domain.Compliments):
			r.h.ShowComplimentsMenu(ctx, msg)
		case string(domain.ImportantDates):
			r.h.ShowImportantDatesMenu(ctx, msg)
		case string(domain.DeleteAccount):
			r.h.DeleteAccount(ctx, msg)
		case string(domain.AddPartner):
			r.h.SetPartner(ctx, msg)
		case string(domain.DeletePartner):
			r.h.DeletePartner(ctx, msg)
		case string(domain.AddCompliment):
			r.h.AddCompliment(ctx, msg)
		case string(domain.DeleteCompliment):
			r.h.DeleteCompliment(ctx, msg)
		case string(domain.GetCompliments):
			r.h.GetCompliments(ctx, msg)
		case string(domain.ReceiveCompliment):
			r.h.ReceiveCompliment(ctx, msg)
		case string(domain.EditComplimentFrequency):
			r.h.EditComplimentFrequency(ctx, msg)
		case string(domain.AddImportantDate):
			r.h.AddImportantDate(ctx, msg)
		default:
			r.h.ReplyUnknownMessage(ctx, msg)
		}
	}
}

func (r *Router) handleCallback(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	data := cq.Data
	username := cq.From.UserName
	text := ""
	if cq.Message != nil {
		text = cq.Message.Text
	}
	log.Printf("Клиент %v написал: %v", username, text)

	section, action, payload := parseCallbackData(data)

	switch section {
	case "account":
		r.handleAccount(ctx, cq, action, payload)
	case "partner":
		r.handlePartner(ctx, cq, action, payload)
	case "compliments":
		r.handleCompliments(ctx, cq, action, payload)
	case "important_dates":
		r.handleImportantDates(ctx, cq, action, payload)
	default:
		r.h.ReplyUnknownCallback(ctx, cq)
	}
}

func (r *Router) handleAccount(ctx context.Context, cq *tgbotapi.CallbackQuery, action string, payload string) {
	switch action {
	case "delete":
		if strings.HasPrefix(payload, "confirm") || strings.HasPrefix(payload, "cancel") {
			r.h.HandleDeleteAccount(ctx, cq)
		}
	default:
		r.h.ReplyUnknownCallback(ctx, cq)
	}
}

func (r *Router) handlePartner(ctx context.Context, cq *tgbotapi.CallbackQuery, action string, payload string) {
	switch action {
	case "delete":
		if strings.HasPrefix(payload, "confirm") || strings.HasPrefix(payload, "cancel") {
			r.h.HandleDeletePartner(ctx, cq)
		}
	default:
		r.h.ReplyUnknownCallback(ctx, cq)
	}
}

func (r *Router) handleCompliments(ctx context.Context, cq *tgbotapi.CallbackQuery, action string, payload string) {
	switch action {
	case "delete":
		if strings.HasPrefix(payload, "confirm") || strings.HasPrefix(payload, "cancel") {
			r.h.HandleDeleteCompliment(ctx, cq)
		}
	default:
		r.h.ReplyUnknownCallback(ctx, cq)
	}
}

func (r *Router) handleImportantDates(ctx context.Context, cq *tgbotapi.CallbackQuery, action string, payload string) {
	switch action {
	case "add":
		switch {
		case strings.HasPrefix(payload, "partner"):
			r.h.HandlePartnerImportantDate(ctx, cq)
		case strings.HasPrefix(payload, "notify_before"):
			r.h.HandleNotifyBeforeImportantDate(ctx, cq)
		case strings.HasPrefix(payload, "year"):
			r.h.HandleYearImportantDate(ctx, cq)
		case strings.HasPrefix(payload, "month"):
			r.h.HandleMonthImportantDate(ctx, cq)
		case strings.HasPrefix(payload, "day"):
			r.h.HandleDayImportantDate(ctx, cq)
		default:
			r.h.ReplyUnknownCallback(ctx, cq)
		}
	default:
		r.h.ReplyUnknownCallback(ctx, cq)
	}
}
