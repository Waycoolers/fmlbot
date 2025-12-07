package handlers

import (
	"context"
	"fmt"
	"strings"

	"github.com/Waycoolers/fmlbot/internal/domain"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) ShowPartnerMenu(_ context.Context, msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	err := h.ui.PartnerMenu(chatID)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при попытке отобразить меню партнеров", err)
		return
	}
}

func (h *Handler) SetPartner(ctx context.Context, msg *tgbotapi.Message) {
	userID := msg.From.ID
	chatID := msg.Chat.ID

	partnerID, err := h.Store.GetPartnerID(ctx, userID)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при попытке получить id партнёра", err)
		return
	}

	if partnerID == 0 {
		err := h.Store.SetUserState(ctx, userID, domain.AwaitingPartner)
		if err != nil {
			h.HandleErr(chatID, "Ошибка при установке состояния awaiting_partner", err)
			return
		}
		h.Reply(chatID, "Отправь username своей половинки\n(Напиши чтобы отменить это действие)")
	} else {
		partnerUsername, er := h.Store.GetUsername(ctx, partnerID)
		if er != nil {
			h.HandleErr(chatID, "Ошибка при попытке получить username партнёра", er)
			return
		}
		h.Reply(chatID, "Твой партнер - @"+partnerUsername+"\nЕсли хочешь изменить партнёра, "+
			"то сначала удали существующего")
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
		h.Reply(chatID, "Партнёр не найден. Попроси его сначала написать боту "+string(domain.Start)+" 😅"+
			"\n(Напиши чтобы отменить это действие)")
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
			err = h.Store.SetUserState(ctx, userID, domain.Empty)
			if err != nil {
				h.HandleErr(chatID, "Ошибка при сбросе состояния", err)
				return
			}
			return
		} else {
			h.Reply(chatID, "У данного пользователя уже есть партнёр 😔")
			err = h.Store.SetUserState(ctx, userID, domain.Empty)
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

	err = h.Store.SetUserState(ctx, partnerID, domain.Empty)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при сбросе состояния", err)
		return
	}

	err = h.Store.SetUserState(ctx, userID, domain.Empty)
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

func (h *Handler) DeletePartner(ctx context.Context, msg *tgbotapi.Message) {
	userID := msg.From.ID
	chatID := msg.Chat.ID
	partnerID, err := h.Store.GetPartnerID(ctx, userID)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при получении id партнера", err)
		return
	}

	if partnerID == 0 {
		h.Reply(userID, "У тебя ещё не добавлен партнер")
		return
	}

	buttons := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Да, удалить 💔", "partner:delete:confirm"),
			tgbotapi.NewInlineKeyboardButtonData("Отмена ❌", "partner:delete:cancel"),
		),
	)

	partnerUsername, err := h.Store.GetUsername(ctx, partnerID)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при попытке получить username партнера", err)
		return
	}

	text := "Вы уверены, что хотите удалить партнёра @" + partnerUsername + "?"

	err = h.ui.Client.SendWithInlineKeyboard(chatID, text, buttons)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при отправке подтверждения", err)
		return
	}
}

func (h *Handler) HandleDeletePartner(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	userID := cb.From.ID
	chatID := cb.Message.Chat.ID
	messageID := cb.Message.MessageID

	switch cb.Data {
	case "partner:delete:confirm":
		partnerID, err := h.Store.GetPartnerID(ctx, userID)
		if err != nil {
			h.ui.RemoveButtons(chatID, messageID)
			h.HandleErr(chatID, "Ошибка при попытке получить id партнера", err)
			return
		}

		err = h.Store.RemovePartners(ctx, userID, partnerID)
		if err != nil {
			h.ui.RemoveButtons(chatID, messageID)
			h.HandleErr(chatID, "Ошибка при попытке удалить партнеров", err)
			return
		}

		h.Reply(chatID, "Партнёр успешно удалён 💔")
		h.Reply(partnerID, "Твой партнёр отписался от тебя 💔")

	case "partner:delete:cancel":
		h.Reply(chatID, "Удаление партнёра отменено")
	}
	h.ui.RemoveButtons(chatID, messageID)
}
