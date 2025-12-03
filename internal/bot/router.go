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

func (r *Router) HandleUpdate(ctx context.Context, update tgbotapi.Update) {
	if update.CallbackQuery != nil {
		chatID := update.CallbackQuery.Message.Chat.ID
		data := update.CallbackQuery.Data
		if data == "delete_confirm" || data == "delete_cancel" {
			err := r.h.HandleDeleteCallback(ctx, update.CallbackQuery)
			if err != nil {
				r.h.HandleErr(chatID, "Ошибка при обработке callback на удаление аккаунта", err)
			}
		}
		if data == "delete_partner_confirm" || data == "delete_partner_cancel" {
			err := r.h.HandleDeletePartnerCallback(ctx, update.CallbackQuery)
			if err != nil {
				r.h.HandleErr(chatID, "Ошибка при обработке callback на удаление партнера", err)
			}
		}
		if strings.HasPrefix(data, "delete_compliment:") || data == "cancel_deletion" {
			err := r.h.HandleDeleteComplimentCallback(ctx, update.CallbackQuery)
			if err != nil {
				r.h.HandleErr(chatID, "Ошибка при обработке callback на удаление комплимента", err)
			}
		}
		return
	}

	msg := update.Message
	userID := msg.From.ID
	chatID := msg.Chat.ID
	text := msg.Text

	username, err := r.h.Store.GetUsername(ctx, userID)
	if err != nil {
		log.Printf("Ошибка при получении юзернейма: %v", err)
	}

	log.Printf("Клиент %v написал: %v", username, text)

	if text == string(models.Start) {
		err = r.h.Store.SetUserState(ctx, userID, models.Empty)
		if err != nil {
			r.h.HandleErr(chatID, "Ошибка при сбросе состояния", err)
			return
		}
		r.h.Start(ctx, msg)
		return
	}

	state, err := r.h.Store.GetUserState(ctx, userID)
	if err != nil {
		r.h.HandleErr(chatID, "Ошибка при получении состояния", err)
		return
	}

	if !strings.HasPrefix(text, "/") {
		switch state {
		case models.AwaitingPartner:
			r.h.ProcessPartnerUsername(ctx, msg)
			return
		case models.AwaitingCompliment:
			r.h.ProcessCompliment(ctx, msg)
			return
		default:
			r.h.Reply(chatID, "Я жду от тебя команду")
			return
		}
	} else {
		switch {
		case strings.HasPrefix(text, string(models.SetPartner)):
			err = r.h.Store.SetUserState(ctx, userID, models.Empty)
			if err != nil {
				r.h.HandleErr(chatID, "Ошибка при сбросе состояния", err)
				return
			}
			r.h.SetPartner(ctx, msg)
			return
		case strings.HasPrefix(text, string(models.DeletePartner)):
			err = r.h.Store.SetUserState(ctx, userID, models.Empty)
			if err != nil {
				r.h.HandleErr(chatID, "Ошибка при сбросе состояния", err)
				return
			}
			r.h.DeletePartner(ctx, msg)
			return
		case strings.HasPrefix(text, string(models.Cancel)):
			r.h.Cancel(ctx, msg)
			return
		case strings.HasPrefix(text, string(models.DeleteAccount)):
			err = r.h.Store.SetUserState(ctx, userID, models.Empty)
			if err != nil {
				r.h.HandleErr(chatID, "Ошибка при сбросе состояния", err)
				return
			}
			r.h.DeleteAccount(ctx, msg)
			return
		case strings.HasPrefix(text, string(models.AddCompliment)):
			err = r.h.Store.SetUserState(ctx, userID, models.Empty)
			if err != nil {
				r.h.HandleErr(chatID, "Ошибка при сбросе состояния", err)
				return
			}
			r.h.AddCompliment(ctx, msg)
			return
		case strings.HasPrefix(text, string(models.GetCompliments)):
			err = r.h.Store.SetUserState(ctx, userID, models.Empty)
			if err != nil {
				r.h.HandleErr(chatID, "Ошибка при сбросе состояния", err)
				return
			}
			r.h.GetCompliments(ctx, msg)
			return
		case strings.HasPrefix(text, string(models.DeleteCompliment)):
			err = r.h.Store.SetUserState(ctx, userID, models.Empty)
			if err != nil {
				r.h.HandleErr(chatID, "Ошибка при сбросе состояния", err)
				return
			}
			r.h.DeleteCompliment(ctx, msg)
			return
		case strings.HasPrefix(text, string(models.ReceiveCompliment)):
			err = r.h.Store.SetUserState(ctx, userID, models.Empty)
			if err != nil {
				r.h.HandleErr(chatID, "Ошибка при сбросе состояния", err)
				return
			}
			r.h.ReceiveCompliment(ctx, msg)
			return
		default:
			r.h.Reply(chatID, "Я не знаю такую команду")
			return
		}
	}
}
