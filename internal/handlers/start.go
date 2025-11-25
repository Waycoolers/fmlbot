package handlers

import (
	"context"
	"log"

	"github.com/Waycoolers/fmlbot/internal/models"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) Start(msg *tgbotapi.Message) {
	ctx := context.Background()
	userID := msg.From.ID
	username := msg.From.UserName

	exists, err := h.Store.IsUserExists(ctx, userID)
	if err != nil {
		h.Reply(msg.Chat.ID, "Ошибка при проверке пользователя 😔")
		log.Printf("Ошибка при проверке пользователя: %v", err)
		return
	}

	if !exists {
		err := h.Store.AddUser(ctx, userID, username)
		if err != nil {
			h.Reply(msg.Chat.ID, "Ошибка при регистрации 😔")
			log.Printf("Ошибка при регистрации: %v", err)
			return
		}
		h.Reply(msg.Chat.ID, "Привет! 💖 Ты зарегистрирован в fmlbot. Добавь партнёра с помощью "+string(models.SetPartner)+"\n"+
			"(Не забудь, что партнер должен тоже зарегистрироваться в боте)")
	} else {
		partnerUsername, err := h.Store.GetPartnerUsername(ctx, userID)
		if err != nil {
			log.Printf("Ошибка при попытке получить username партнера: %v", err)
			return
		}

		if partnerUsername == "" {
			h.Reply(msg.Chat.ID, "Ты уже зарегистрирован! Используй "+string(models.SetPartner)+", чтобы добавить партнёра 💌")
		} else {
			text := "Ты уже зарегистрирован! Твой партнер - @" + partnerUsername
			h.Reply(msg.Chat.ID, text)
		}
	}
}
