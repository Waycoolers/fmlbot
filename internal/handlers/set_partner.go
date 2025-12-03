package handlers

import (
	"context"
	"fmt"
	"strings"

	"github.com/Waycoolers/fmlbot/internal/models"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) SetPartner(ctx context.Context, msg *tgbotapi.Message) {
	userID := msg.From.ID
	chatID := msg.Chat.ID

	err := h.Store.SetUserState(ctx, userID, models.AwaitingPartner)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при установке состояния awaiting_partner", err)
		return
	}

	partnerID, err := h.Store.GetPartnerID(ctx, userID)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при попытке получить id партнёра", err)
		return
	}

	if partnerID == 0 {
		h.Reply(chatID, "Отправь username своей половинки\n(Напиши "+string(models.Cancel)+" чтобы отменить это действие)")
	} else {
		partnerUsername, er := h.Store.GetUsername(ctx, partnerID)
		if er != nil {
			h.HandleErr(chatID, "Ошибка при попытке получить username партнёра", er)
			return
		}
		h.Reply(chatID, "Твой партнер - @"+partnerUsername+"\nЕсли хочешь изменить аккаунт партнёра, "+
			"то отправь username своей половинки\n(Напиши "+string(models.Cancel)+" чтобы отменить это действие)")
	}
}

func (h *Handler) ProcessPartnerUsername(ctx context.Context, msg *tgbotapi.Message) {
	userID := msg.From.ID
	chatID := msg.Chat.ID
	partnerUsername := msg.Text
	userUsername := msg.From.UserName

	if strings.HasPrefix(partnerUsername, "@") {
		partnerUsername = partnerUsername[1:]
	}

	exists, err := h.Store.IsUserExistsByUsername(ctx, partnerUsername)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при проверке партнёра", err)
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
		h.HandleErr(chatID, "Ошибка при получении id партнера", err)
		return
	}
	correctPartnerUsername, _ := h.Store.GetUsername(ctx, partnerID)

	partnerExists, err := h.Store.GetPartnerID(ctx, partnerID)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при проверке на существование партнёра", err)
		return
	}

	if partnerExists != 0 {
		if partnerExists == userID {
			h.Reply(chatID, "@"+correctPartnerUsername+" и так ваш партнёр. Приятного времяпрепровождения!")
			err = h.Store.SetUserState(ctx, userID, models.Empty)
			if err != nil {
				h.HandleErr(chatID, "Ошибка при сбросе состояния", err)
				return
			}
			return
		} else {
			h.Reply(chatID, "У данного пользователя уже есть партнёр 😔")
			err = h.Store.SetUserState(ctx, userID, models.Empty)
			if err != nil {
				h.HandleErr(chatID, "Ошибка при сбросе состояния", err)
				return
			}
			return
		}
	}

	userPartnerExists, err := h.Store.GetPartnerID(ctx, userID)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при проверке на существование партнёра", err)
		return
	}

	if userPartnerExists != 0 {
		err = h.Store.SetPartner(ctx, userPartnerExists, 0)
		if err != nil {
			h.HandleErr(chatID, "Ошибка при сбросе партнера у партнера", err)
			return
		}
		h.Reply(userPartnerExists, "Твой партнёр добавил другого партнёра 💔")
	}

	err = h.Store.SetUserState(ctx, partnerID, models.Empty)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при сбросе состояния", err)
		return
	}

	err = h.Store.SetUserState(ctx, userID, models.Empty)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при сбросе состояния", err)
		return
	}

	err = h.Store.SetPartners(ctx, userID, partnerID)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при связи партнеров", err)
		return
	}

	h.Reply(partnerID, "💞 Ура! Теперь вы и @"+userUsername+" — официально пара в боте 💌")
	h.Reply(chatID, fmt.Sprintf("Партнёр успешно добавлен! 💖 (@%s)", correctPartnerUsername))
}
