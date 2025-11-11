package bot

import (
	"context"
	"log"
	"strings"

	"github.com/Waycoolers/fmlbot/internal/handlers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Router struct {
	h *handlers.Handler
}

func NewRouter(h *handlers.Handler) *Router {
	return &Router{h: h}
}

func (r *Router) HandleUpdate(update tgbotapi.Update) {
	if update.CallbackQuery != nil {
		data := update.CallbackQuery.Data
		if data == "delete_confirm" || data == "delete_cancel" {
			err := r.h.HandleDeleteCallback(update.CallbackQuery)
			if err != nil {
				log.Printf("Ошибка при обработке callback: %v", err)
			}
		}
		return
	}

	msg := update.Message
	text := msg.Text
	userID := msg.From.ID
	username, _ := r.h.Store.GetUsername(context.Background(), userID)

	log.Printf("Клиент %v написал: %v", username, text)

	if text == "/start" {
		_ = r.h.Store.SetUserState(context.Background(), userID, "")
		r.h.Start(msg)
		return
	}

	state, err := r.h.Store.GetUserState(context.Background(), userID)
	if err != nil {
		log.Printf("Ошибка при получении состояния: %v", err)
		r.h.Reply(msg.Chat.ID, "Произошла ошибка 😔")
		return
	}

	if state == "awaiting_partner" {
		r.h.ProcessPartnerUsername(msg)
		return
	}

	switch {
	case strings.HasPrefix(text, "/setpartner"):
		_ = r.h.Store.SetUserState(context.Background(), userID, "")
		r.h.SetPartner(msg)
	case strings.HasPrefix(text, "/delete"):
		_ = r.h.Store.SetUserState(context.Background(), userID, "")
		r.h.DeleteAccount(msg)
	case strings.HasPrefix(text, "/compliment"):
		_ = r.h.Store.SetUserState(context.Background(), userID, "")
		r.h.Compliment(msg)
	default:
		r.h.Reply(msg.Chat.ID, "Некорректный ввод")
	}
}
