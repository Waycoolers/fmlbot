package handlers

import (
	"context"
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) SetPartner(msg *tgbotapi.Message) {
	ctx := context.Background()
	userID := msg.From.ID

	err := h.Store.SetUserState(ctx, userID, "awaiting_partner")
	if err != nil {
		h.Reply(msg.Chat.ID, "Ошибка при установке состояния")
		return
	}

	partnerUsername, err := h.Store.GetPartnerUsername(ctx, userID)
	if err != nil {
		h.Reply(msg.Chat.ID, "Ошибка при получении информации о партнёре 😔")
		log.Printf("Ошибка при получении информации о партнёре: %v", err)
		return
	}

	if partnerUsername == "" {
		h.Reply(msg.Chat.ID, "Отправь username своей половинки")
	} else {
		h.Reply(msg.Chat.ID, "Твой партнер - @"+partnerUsername+"\nЕсли хочешь изменить аккаунт партнёра, "+
			"то отправь username своей половинки")
	}
}

func (h *Handler) ProcessPartnerUsername(msg *tgbotapi.Message) {
	ctx := context.Background()
	userID := msg.From.ID
	partnerUsername := msg.Text
	userUsername := msg.From.UserName

	if strings.HasPrefix(partnerUsername, "@") {
		partnerUsername = partnerUsername[1:]
	}

	exists, err := h.Store.IsUserExistsByUsername(ctx, partnerUsername)
	if err != nil {
		h.Reply(msg.Chat.ID, "Ошибка при проверке партнёра 😔")
		log.Printf("Ошибка при проверке партнёра: %v", err)
		return
	}

	if !exists {
		h.Reply(msg.Chat.ID, "Партнёр не найден. Попросите его сначала написать боту /start 😅")
		log.Printf("Ошибка. Партнёр не найден: %v", err)
		return
	}

	partnerID, err := h.Store.GetUserIDByUsername(ctx, partnerUsername)
	if err != nil {
		h.Reply(msg.Chat.ID, "Ошибка при проверке партнёра 😔")
		log.Printf("Ошибка при получении ID партнера: %v", err)
	}
	correctPartnerUsername, _ := h.Store.GetUsername(ctx, partnerID)

	// Сохраняем связь user → partner
	err = h.Store.SetPartner(ctx, userID, correctPartnerUsername)
	if err != nil {
		h.Reply(msg.Chat.ID, "Не удалось сохранить партнёра 😔")
		log.Printf("Ошибка при попытке сохранения связи user → partner: %v", err)
		return
	}

	// Сохраняем связь partner → user
	err = h.Store.SetPartner(ctx, partnerID, userUsername)
	if err != nil {
		h.Reply(msg.Chat.ID, "Не удалось сохранить партнёра 😔")
		log.Printf("Ошибка при попытке сохранения связи partner → user: %v", err)
		return
	}
	h.Reply(partnerID, "💞 Ура! Теперь вы и @"+userUsername+" — официально пара в боте 💌")

	_ = h.Store.SetUserState(ctx, userID, "")

	h.Reply(msg.Chat.ID, fmt.Sprintf("Партнёр успешно добавлен! 💖 (@%s)", correctPartnerUsername))
}
