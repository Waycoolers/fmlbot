package handlers

import (
	"context"

	"github.com/Waycoolers/fmlbot/internal/models"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) Start(msg *tgbotapi.Message) {
	ctx := context.Background()
	userID := msg.From.ID
	chatID := msg.Chat.ID
	username := msg.From.UserName

	exists, err := h.Store.IsUserExists(ctx, userID)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при проверке пользователя", err)
		return
	}

	if !exists {
		er := h.Store.AddUser(ctx, userID, username)
		if er != nil {
			h.HandleErr(chatID, "Ошибка при регистрации", err)
			return
		}
		h.Reply(chatID, "Привет! 💖 Ты зарегистрирован в fmlbot. Добавь партнёра с помощью "+string(models.SetPartner)+"\n"+
			"(Не забудь, что партнер должен тоже зарегистрироваться в боте)")
	} else {
		partnerID, er := h.Store.GetPartnerID(ctx, userID)
		if er != nil {
			h.HandleErr(chatID, "Ошибка при попытке получить id партнера", err)
			return
		}

		if partnerID == 0 {
			h.Reply(chatID, "Ты уже зарегистрирован! Используй "+string(models.SetPartner)+", чтобы добавить партнёра 💌")
		} else {
			partnerUsername, err2 := h.Store.GetUsername(ctx, partnerID)
			if err2 != nil {
				h.HandleErr(chatID, "Ошибка при попытке получить username партнера", err2)
				return
			}
			text := "Ты уже зарегистрирован! Твой партнер - @" + partnerUsername
			h.Reply(chatID, text)
		}
	}
}
