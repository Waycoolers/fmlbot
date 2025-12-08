package handlers

import (
	"context"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) ShowAccountMenu(_ context.Context, msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	text := "Меню аккаунта"
	err := h.ui.AccountMenu(chatID, text)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при попытке отобразить меню аккаунтов", err)
		return
	}
}

func (h *Handler) Register(ctx context.Context, msg *tgbotapi.Message) {
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
	}
}

func (h *Handler) DeleteAccount(_ context.Context, msg *tgbotapi.Message) {
	chatID := msg.Chat.ID

	buttons := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Да, удалить 💔", "account:delete:confirm"),
			tgbotapi.NewInlineKeyboardButtonData("Отмена ❌", "account:delete:cancel"),
		),
	)

	text := "Ты уверен, что хочешь удалить аккаунт? Все твои пользовательские данные тоже будут удалены."

	err := h.ui.Client.SendWithInlineKeyboard(chatID, text, buttons)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при отправке подтверждения", err)
		return
	}
}

func (h *Handler) HandleDeleteAccount(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	userID := cq.From.ID
	chatID := cq.Message.Chat.ID
	messageID := cq.Message.MessageID

	switch cq.Data {
	case "account:delete:confirm":
		partnerID, err := h.Store.GetPartnerID(ctx, userID)
		if err != nil {
			h.ui.RemoveButtons(chatID, messageID)
			h.HandleErr(chatID, "Ошибка при попытке получить id партнера", err)
			return
		}

		if partnerID != 0 {
			err = h.Store.RemovePartners(ctx, userID, partnerID)
			if err != nil {
				h.ui.RemoveButtons(chatID, messageID)
				h.HandleErr(chatID, "Ошибка при попытке удалить партнеров", err)
				return
			}

			err = h.Store.DeleteUser(ctx, userID)
			if err != nil {
				h.ui.RemoveButtons(chatID, messageID)
				h.HandleErr(chatID, "Ошибка при попытке удалить юзера", err)
				return
			}
			h.Reply(partnerID, "Твой партнёр удалил свой аккаунт 💔")
		} else {
			err = h.Store.DeleteUser(ctx, userID)
			if err != nil {
				h.ui.RemoveButtons(chatID, messageID)
				h.HandleErr(chatID, "Ошибка при попытке удалить юзера", err)
				return
			}
		}

		h.Reply(chatID, "Твой аккаунт успешно удалён 💔")
		text := "Чтобы разбудить бота, зарегистрируйся по кнопке ниже"
		err = h.ui.StartMenu(chatID, text)
		if err != nil {
			log.Printf("Ошибка при вызове стартового меню")
			h.Reply(chatID, "Перезапусти бота с помощью /start")
		}
	case "account:delete:cancel":
		h.Reply(chatID, "Удаление аккаунта отменено ✅")
	}
	h.ui.RemoveButtons(chatID, messageID)
}
