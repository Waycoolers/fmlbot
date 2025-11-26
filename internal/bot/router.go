package bot

import (
	"context"
	"log"
	"strings"

	"github.com/Waycoolers/fmlbot/internal/handlers"
	"github.com/Waycoolers/fmlbot/internal/models"
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
				log.Printf("Ошибка при обработке callback на удаление аккаунта: %v", err)
			}
		}
		if data == "delete_partner_confirm" || data == "delete_partner_cancel" {
			err := r.h.HandleDeletePartnerCallback(update.CallbackQuery)
			if err != nil {
				log.Printf("Ошибка при обработке callback на удаление партнера: %v", err)
			}
		}
		return
	}

	ctx := context.Background()
	msg := update.Message
	text := msg.Text
	userID := msg.From.ID
	username, _ := r.h.Store.GetUsername(ctx, userID)

	log.Printf("Клиент %v написал: %v", username, text)

	if text == string(models.Start) {
		_ = r.h.Store.SetUserState(ctx, userID, models.Empty)
		r.h.Start(msg)
		return
	}

	state, err := r.h.Store.GetUserState(ctx, userID)
	if err != nil {
		log.Printf("Ошибка при получении состояния: %v", err)
		r.h.Reply(msg.Chat.ID, "Произошла ошибка 😔")
		return
	}

	if !strings.HasPrefix(text, "/") {
		switch state {
		case models.AwaitingPartner:
			r.h.ProcessPartnerUsername(msg)
			return
		case models.AwaitingCompliment:
			r.h.ProcessCompliment(msg)
			return
		default:
			r.h.Reply(msg.Chat.ID, "Я жду от тебя команду")
			return
		}
	} else {
		switch {
		case strings.HasPrefix(text, string(models.SetPartner)):
			_ = r.h.Store.SetUserState(ctx, userID, models.Empty)
			r.h.SetPartner(msg)
			return
		case strings.HasPrefix(text, string(models.DeletePartner)):
			_ = r.h.Store.SetUserState(ctx, userID, models.Empty)
			r.h.DeletePartner(msg)
			return
		case strings.HasPrefix(text, string(models.Cancel)):
			r.h.Cancel(msg)
			return
		case strings.HasPrefix(text, string(models.Delete)):
			_ = r.h.Store.SetUserState(ctx, userID, models.Empty)
			r.h.DeleteAccount(msg)
			return
		case strings.HasPrefix(text, string(models.AddCompliment)):
			_ = r.h.Store.SetUserState(ctx, userID, models.Empty)
			r.h.AddCompliment(msg)
			return
		default:
			r.h.Reply(msg.Chat.ID, "Я не знаю такую команду")
			return
		}
	}
}
