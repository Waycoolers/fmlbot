package handlers

import (
	"context"
	"fmt"
	"strings"

	"github.com/Waycoolers/fmlbot/internal/models"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) SetPartner(msg *tgbotapi.Message) {
	ctx := context.Background()
	userID := msg.From.ID
	chatID := msg.Chat.ID

	err := h.Store.SetUserState(ctx, userID, models.AwaitingPartner)
	if err != nil {
		h.handleErr(chatID, "Ошибка при установке состояния awaiting_partner", err)
		return
	}

	partnerUsername, err := h.Store.GetPartnerUsername(ctx, userID)
	if err != nil {
		h.handleErr(chatID, "Ошибка при получении информации о партнёре", err)
		return
	}

	if partnerUsername == "" {
		h.Reply(chatID, "Отправь username своей половинки\n(Напиши "+string(models.Cancel)+" чтобы отменить это действие)")
	} else {
		h.Reply(chatID, "Твой партнер - @"+partnerUsername+"\nЕсли хочешь изменить аккаунт партнёра, "+
			"то отправь username своей половинки\n(Напиши "+string(models.Cancel)+" чтобы отменить это действие)")
	}
}

func (h *Handler) ProcessPartnerUsername(msg *tgbotapi.Message) {
	ctx := context.Background()
	userID := msg.From.ID
	chatID := msg.Chat.ID
	partnerUsername := msg.Text
	userUsername := msg.From.UserName

	if strings.HasPrefix(partnerUsername, "@") {
		partnerUsername = partnerUsername[1:]
	}

	exists, err := h.Store.IsUserExistsByUsername(ctx, partnerUsername)
	if err != nil {
		h.handleErr(chatID, "Ошибка при проверке партнёра", err)
		return
	}

	if strings.ToLower(partnerUsername) == strings.ToLower(userUsername) {
		h.Reply(chatID, "Ты не можешь добавить самого себя 😅")
		return
	}

	if !exists {
		h.Reply(chatID, "Партнёр не найден. Попроси его сначала написать боту "+string(models.Start)+" 😅"+
			"\n(Напиши "+string(models.Cancel)+" чтобы отменить это действие)")
		return
	}

	partnerID, err := h.Store.GetUserIDByUsername(ctx, partnerUsername)
	if err != nil {
		h.handleErr(chatID, "Ошибка при получении ID партнера", err)
		return
	}
	correctPartnerUsername, _ := h.Store.GetUsername(ctx, partnerID)

	partnerExists, err := h.Store.GetPartnerUsername(ctx, partnerID)
	if err != nil {
		h.handleErr(chatID, "Ошибка при проверке на существование партнёра", err)
		return
	}

	if partnerExists != "" {
		if partnerExists == userUsername {
			h.Reply(chatID, "@"+correctPartnerUsername+" и так ваш партнёр. Приятного времяпрепровождения!")
			err = h.Store.SetUserState(ctx, userID, models.Empty)
			if err != nil {
				h.handleErr(chatID, "Ошибка при сбросе состояния", err)
				return
			}
			return
		} else {
			h.Reply(chatID, "У данного пользователя уже есть партнёр 😔")
			err = h.Store.SetUserState(ctx, userID, models.Empty)
			if err != nil {
				h.handleErr(chatID, "Ошибка при сбросе состояния", err)
				return
			}
			return
		}
	}

	userPartnerExists, err := h.Store.GetPartnerUsername(ctx, userID)
	if err != nil {
		h.handleErr(chatID, "Ошибка при проверке на существование партнёра", err)
		return
	}

	if userPartnerExists != "" {
		userPartnerID, er := h.Store.GetUserIDByUsername(ctx, userPartnerExists)
		if er != nil {
			h.handleErr(chatID, "Ошибка при получении ID партнёра", er)
			return
		}

		er = h.Store.SetPartner(ctx, userPartnerID, "")
		h.Reply(userPartnerID, "Твой партнёр добавил другого партнёра 💔")
	}

	err = h.Store.SetUserState(ctx, partnerID, models.Empty)
	if err != nil {
		h.handleErr(chatID, "Ошибка при сбросе состояния", err)
		return
	}

	err = h.Store.SetUserState(ctx, userID, models.Empty)
	if err != nil {
		h.handleErr(chatID, "Ошибка при сбросе состояния", err)
		return
	}

	err = h.Store.SetPartners(ctx, userID, partnerID, userUsername, correctPartnerUsername)
	if err != nil {
		h.handleErr(chatID, "Ошибка при связи партнеров", err)
		return
	}

	h.Reply(partnerID, "💞 Ура! Теперь вы и @"+userUsername+" — официально пара в боте 💌")
	h.Reply(chatID, fmt.Sprintf("Партнёр успешно добавлен! 💖 (@%s)", correctPartnerUsername))
}
