package app

import (
	"context"
	"log"
	"strings"

	"github.com/Waycoolers/fmlbot/internal/domain"
	"github.com/Waycoolers/fmlbot/internal/handlers"
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
	} else {
		switch {
		case strings.HasPrefix(text, string(domain.SetPartner)):
			err = r.h.Store.SetUserState(ctx, userID, domain.Empty)
			if err != nil {
				r.h.HandleErr(chatID, "Ошибка при сбросе состояния", err)
				return
			}
			r.h.SetPartner(ctx, msg)
			return
		case strings.HasPrefix(text, string(domain.DeletePartner)):
			err = r.h.Store.SetUserState(ctx, userID, domain.Empty)
			if err != nil {
				r.h.HandleErr(chatID, "Ошибка при сбросе состояния", err)
				return
			}
			r.h.DeletePartner(ctx, msg)
			return
		case strings.HasPrefix(text, string(domain.Cancel)):
			r.h.Cancel(ctx, msg)
			return
		case strings.HasPrefix(text, string(domain.DeleteAccount)):
			err = r.h.Store.SetUserState(ctx, userID, domain.Empty)
			if err != nil {
				r.h.HandleErr(chatID, "Ошибка при сбросе состояния", err)
				return
			}
			r.h.DeleteAccount(ctx, msg)
			return
		case strings.HasPrefix(text, string(domain.AddCompliment)):
			err = r.h.Store.SetUserState(ctx, userID, domain.Empty)
			if err != nil {
				r.h.HandleErr(chatID, "Ошибка при сбросе состояния", err)
				return
			}
			r.h.AddCompliment(ctx, msg)
			return
		case strings.HasPrefix(text, string(domain.GetCompliments)):
			err = r.h.Store.SetUserState(ctx, userID, domain.Empty)
			if err != nil {
				r.h.HandleErr(chatID, "Ошибка при сбросе состояния", err)
				return
			}
			r.h.GetCompliments(ctx, msg)
			return
		case strings.HasPrefix(text, string(domain.DeleteCompliment)):
			err = r.h.Store.SetUserState(ctx, userID, domain.Empty)
			if err != nil {
				r.h.HandleErr(chatID, "Ошибка при сбросе состояния", err)
				return
			}
			r.h.DeleteCompliment(ctx, msg)
			return
		case strings.HasPrefix(text, string(domain.ReceiveCompliment)):
			err = r.h.Store.SetUserState(ctx, userID, domain.Empty)
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

func (r *Router) handleCallback(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	chatID := cq.Message.Chat.ID
	data := cq.Data
	if data == "delete_confirm" || data == "delete_cancel" {
		err := r.h.HandleDeleteCallback(ctx, cq)
		if err != nil {
			r.h.HandleErr(chatID, "Ошибка при обработке callback на удаление аккаунта", err)
		}
	}
	if data == "delete_partner_confirm" || data == "delete_partner_cancel" {
		err := r.h.HandleDeletePartnerCallback(ctx, cq)
		if err != nil {
			r.h.HandleErr(chatID, "Ошибка при обработке callback на удаление партнера", err)
		}
	}
	if strings.HasPrefix(data, "delete_compliment:") || data == "cancel_deletion" {
		err := r.h.HandleDeleteComplimentCallback(ctx, cq)
		if err != nil {
			r.h.HandleErr(chatID, "Ошибка при обработке callback на удаление комплимента", err)
		}
	}
	return
}
