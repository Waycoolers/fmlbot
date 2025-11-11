package handlers

import (
	"context"
	"log"
	"os"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) Compliment(msg *tgbotapi.Message) {
	ctx := context.Background()

	limitStr := os.Getenv("LIMIT_COMPLIMENTS_PER_DAY")
	dailyLimit, err := strconv.Atoi(limitStr)
	if err != nil {
		dailyLimit = 3
	}

	userID := msg.Chat.ID
	canSend, err := h.Store.CanSendCompliment(ctx, userID, dailyLimit)
	if err != nil {
		log.Println(err)
		h.Reply(msg.Chat.ID, "Ошибка при проверке лимита 😔")
		return
	}

	if !canSend {
		h.Reply(msg.Chat.ID, "Комплименты на сегодня закончились 💐")
		return
	}

	complimentID, text, err := h.Store.GetNextCompliment(ctx)
	if err != nil {
		text = "😅 У меня сейчас нет комплиментов, но ты всё равно чудесная!"
	}

	err = h.Store.RecordCompliment(ctx, userID, complimentID)
	if err != nil {
		log.Printf("Ошибка при записи комплимента: %v", err)
	}

	h.Reply(msg.Chat.ID, text)
}
