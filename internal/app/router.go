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
	username, err := r.h.Store.GetUsername(ctx, userID)
	if err != nil {
		r.h.HandleErr(chatID, "Ошибка при получении юзернейма", err)
		return
	}
	log.Printf("Клиент %v написал: %v", username, text)

	if text == string(domain.Start) {
		err = r.h.Store.SetUserState(ctx, userID, domain.Empty)
		if err != nil {
			r.h.HandleErr(chatID, "Ошибка при сбросе состояния", err)
			return
		}
		r.h.ShowMainMenu(ctx, chatID)
		return
	}

	state, err := r.h.Store.GetUserState(ctx, userID)
	if err != nil {
		r.h.HandleErr(chatID, "Ошибка при получении состояния", err)
		return
	}

	if !strings.HasPrefix(text, "/") {
		switch state {
		case domain.AwaitingPartner:
			r.h.ProcessPartnerUsername(ctx, msg)
			return
		case domain.AwaitingCompliment:
			r.h.ProcessCompliment(ctx, msg)
			return
		default:
			r.h.Reply(chatID, "Я жду от тебя команду")
			return
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
	case "menu":
		r.handleMenu(ctx, cq, action)
	case "account":
		r.handleAccount(ctx, cq, action, payload)
	case "partner":
		r.handlePartner(ctx, cq, action, payload)
	case "compliments":
		r.handleCompliments(ctx, cq, action, payload)
	default:
		r.h.ReplyUnknownCallback(ctx, cq)
	}
}

func (r *Router) handleMenu(ctx context.Context, cq *tgbotapi.CallbackQuery, action string) {
	switch action {
	case "main":
		r.h.ShowMainMenu(ctx, cq.Message.Chat.ID)
	case "account":
		r.h.ShowAccountMenu(ctx, cq)
	case "partner":
		r.h.ShowPartnerMenu(ctx, cq)
	case "compliments":
		r.h.ShowComplimentsMenu(ctx, cq)
	default:
		r.h.ReplyUnknownCallback(ctx, cq)
	}
}

func (r *Router) handleAccount(ctx context.Context, cq *tgbotapi.CallbackQuery, action string, payload string) {
	switch action {
	case "register":
		r.h.Register(ctx, cq)
	case "delete":
		if strings.HasPrefix(payload, "confirm") || strings.HasPrefix(payload, "cancel") {
			r.h.HandleDeleteAccount(ctx, cq)
		} else {
			r.h.DeleteAccount(ctx, cq)
		}
	default:
		r.h.ReplyUnknownCallback(ctx, cq)
	}
}

func (r *Router) handlePartner(ctx context.Context, cq *tgbotapi.CallbackQuery, action string, payload string) {
	switch action {
	case "add":
		r.h.SetPartner(ctx, cq)
	case "delete":
		if strings.HasPrefix(payload, "confirm") || strings.HasPrefix(payload, "cancel") {
			r.h.HandleDeletePartner(ctx, cq)
		} else {
			r.h.DeletePartner(ctx, cq)
		}
	default:
		r.h.ReplyUnknownCallback(ctx, cq)
	}
}

func (r *Router) handleCompliments(ctx context.Context, cq *tgbotapi.CallbackQuery, action string, payload string) {
	switch action {
	case "add":
		r.h.AddCompliment(ctx, cq)
	case "delete":
		if strings.HasPrefix(payload, "confirm") || strings.HasPrefix(payload, "cancel") {
			r.h.HandleDeleteCompliment(ctx, cq)
		} else {
			r.h.DeleteCompliment(ctx, cq)
		}
	case "all":
		r.h.GetCompliments(ctx, cq)
	case "receive":
		r.h.ReceiveCompliment(ctx, cq)
	default:
		r.h.ReplyUnknownCallback(ctx, cq)
	}
}
