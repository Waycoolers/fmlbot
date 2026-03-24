package handlers

import (
	"context"
	"fmt"
	"strings"

	"github.com/Waycoolers/fmlbot/internal/domain"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) ShowPartnerMenu(ctx context.Context, msg *domain.Message) {
	userID := msg.UserID
	chatID := msg.ChatID
	text := "👤 Партнёр"

	partnerID, err := h.Store.Users.GetPartnerID(ctx, userID)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при попытке получить id партнера", err)
		return
	}

	if partnerID == 0 {
		text = "🤍 У тебя пока нет партнёра"
	} else {
		partnerUsername, er := h.Store.Users.GetUsername(ctx, partnerID)
		if er != nil {
			h.HandleErr(chatID, "Ошибка при попытке получить username партнера", er)
			return
		}

		text = "💞 Твой партнёр: @" + partnerUsername
	}

	err = h.ui.PartnerMenu(chatID, text)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при попытке отобразить меню партнеров", err)
		return
	}
}

func (h *Handler) SetPartner(ctx context.Context, msg *domain.Message) {
	userID := msg.UserID
	chatID := msg.ChatID

	partnerID, err := h.Store.Users.GetPartnerID(ctx, userID)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при попытке получить id партнёра", err)
		return
	}

	if partnerID == 0 {
		er := h.Store.Users.SetUserState(ctx, userID, domain.AwaitingPartner)
		if er != nil {
			h.HandleErr(chatID, "Ошибка при установке состояния awaiting_partner", er)
			return
		}
		h.Reply(chatID, "💌 Отправь username партнёра")
	} else {
		partnerUsername, er := h.Store.Users.GetUsername(ctx, partnerID)
		if er != nil {
			h.HandleErr(chatID, "Ошибка при попытке получить username партнёра", er)
			return
		}
		h.Reply(
			chatID,
			"💞 Сейчас твой партнёр — @"+partnerUsername+
				"\nЕсли хочешь изменить выбор, сначала нужно удалить текущего партнёра",
		)
	}
}

func (h *Handler) ProcessPartnerUsername(ctx context.Context, msg *domain.Message) {
	userID := msg.UserID
	chatID := msg.ChatID
	partnerUsername := msg.Text
	userUsername := msg.UserName

	if strings.HasPrefix(partnerUsername, "@") {
		partnerUsername = partnerUsername[1:]
	}

	exists, err := h.Store.Users.IsUserExistsByUsername(ctx, partnerUsername)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при проверке партнёра", err)
		return
	}

	if strings.ToLower(partnerUsername) == strings.ToLower(userUsername) {
		h.Reply(chatID, "😅 Так не получится — себя добавить нельзя")
		return
	}

	if !exists {
		h.Reply(
			chatID,
			"🤔 Я не нашёл этого пользователя\n"+
				"Пусть он сначала напишет боту команду "+string(domain.Start)+"\n\n",
		)
		return
	}

	partnerID, err := h.Store.Users.GetUserIDByUsername(ctx, partnerUsername)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при получении id партнера", err)
		return
	}
	correctPartnerUsername, _ := h.Store.Users.GetUsername(ctx, partnerID)

	partnerExists, err := h.Store.Users.GetPartnerID(ctx, partnerID)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при проверке на существование партнёра", err)
		return
	}

	if partnerExists != 0 {
		if partnerExists == userID {
			h.Reply(chatID, "💛 @"+correctPartnerUsername+" и так ваш партнёр. Приятного времяпрепровождения!")
			err = h.Store.Users.SetUserState(ctx, userID, domain.Empty)
			if err != nil {
				h.HandleErr(chatID, "Ошибка при сбросе состояния", err)
				return
			}
			return
		} else {
			h.Reply(chatID, "😔 У этого пользователя уже есть партнёр")
			err = h.Store.Users.SetUserState(ctx, userID, domain.Empty)
			if err != nil {
				h.HandleErr(chatID, "Ошибка при сбросе состояния", err)
				return
			}
			return
		}
	}

	userPartnerExists, err := h.Store.Users.GetPartnerID(ctx, userID)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при проверке на существование партнёра", err)
		return
	}

	if userPartnerExists != 0 {
		err = h.Store.Users.SetPartner(ctx, userPartnerExists, 0)
		if err != nil {
			h.HandleErr(chatID, "Ошибка при сбросе партнера у партнера", err)
			return
		}
		h.Reply(userPartnerExists, "💔 Твой партнёр добавил другого партнёра")
	}

	err = h.Store.Users.SetUserState(ctx, partnerID, domain.Empty)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при сбросе состояния", err)
		return
	}

	err = h.Store.Users.SetUserState(ctx, userID, domain.Empty)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при сбросе состояния", err)
		return
	}

	err = h.Store.Users.SetPartners(ctx, userID, partnerID)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при связи партнеров", err)
		return
	}

	h.Reply(partnerID, "💞 У вас с @"+userUsername+" теперь есть общая история в боте ✨")
	h.Reply(chatID, fmt.Sprintf("✨ Готово! Партнёр @%s добавлен", correctPartnerUsername))
}

func (h *Handler) DeletePartner(ctx context.Context, msg *domain.Message) {
	userID := msg.UserID
	chatID := msg.ChatID
	partnerID, err := h.Store.Users.GetPartnerID(ctx, userID)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при получении id партнера", err)
		return
	}

	if partnerID == 0 {
		h.Reply(chatID, "🤍 У тебя сейчас не добавлен партнёр")
		return
	}

	buttons := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("💔 Да, удалить", "partner:delete:confirm"),
			tgbotapi.NewInlineKeyboardButtonData("↩️ Передумал(а)", "partner:delete:cancel"),
		),
	)

	partnerUsername, err := h.Store.Users.GetUsername(ctx, partnerID)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при попытке получить username партнера", err)
		return
	}

	text := "💭 Ты уверен(а), что хочешь удалить партнёра @" + partnerUsername + "?\n" +
		"Все общие настройки будут сброшены."

	err = h.ui.Client.SendWithInlineKeyboard(chatID, text, buttons)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при отправке подтверждения", err)
		return
	}
}

func (h *Handler) HandleDeletePartner(ctx context.Context, cb *domain.CallbackQuery) {
	userID := cb.UserID
	chatID := cb.ChatID
	messageID := cb.MessageID

	switch cb.Data {
	case "partner:delete:confirm":
		partnerID, err := h.Store.Users.GetPartnerID(ctx, userID)
		if err != nil {
			h.ui.RemoveButtons(chatID, messageID)
			h.HandleErr(chatID, "Ошибка при попытке получить id партнера", err)
			return
		}

		err = h.Store.UserConfig.SetDefault(ctx, userID)
		if err != nil {
			h.ui.RemoveButtons(chatID, messageID)
			h.HandleErr(chatID, "Ошибка при сбросе конфига", err)
			return
		}
		err = h.Store.UserConfig.SetDefault(ctx, partnerID)
		if err != nil {
			h.ui.RemoveButtons(chatID, messageID)
			h.HandleErr(chatID, "Ошибка при сбросе конфига", err)
			return
		}

		err = h.Store.Users.RemovePartners(ctx, userID, partnerID)
		if err != nil {
			h.ui.RemoveButtons(chatID, messageID)
			h.HandleErr(chatID, "Ошибка при попытке удалить партнеров", err)
			return
		}

		h.Reply(chatID, "🕊️ Партнёр удалён")
		h.Reply(partnerID, "💔 Твой партнёр больше не связан с тобой")

	case "partner:delete:cancel":
		h.Reply(chatID, "💛 Хорошо, ничего не меняем")
	}
	h.ui.RemoveButtons(chatID, messageID)
}
