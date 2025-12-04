package handlers

import (
	"context"

	"github.com/Waycoolers/fmlbot/internal/domain"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) ShowAccountMenu(_ context.Context, cq *tgbotapi.CallbackQuery) {
	chatID := cq.Message.Chat.ID
	err := h.ui.AccountMenu(chatID)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при попытке отобразить меню аккаунтов", err)
		return
	}
}

func (h *Handler) Register(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	userID := cq.From.ID
	chatID := cq.Message.Chat.ID
	messageID := cq.Message.MessageID
	username := cq.From.UserName

	exists, err := h.Store.IsUserExists(ctx, userID)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при проверке пользователя", err)
		h.ui.RemoveButtons(chatID, messageID)
		return
	}

	if !exists {
		er := h.Store.AddUser(ctx, userID, username)
		if er != nil {
			h.HandleErr(chatID, "Ошибка при регистрации", err)
			h.ui.RemoveButtons(chatID, messageID)
			return
		}
		h.Reply(chatID, "Привет! 💖 Ты зарегистрирован в fmlbot. Добавь партнёра с помощью "+string(domain.SetPartner)+"\n"+
			"(Не забудь, что партнер должен тоже зарегистрироваться в боте)")
		h.ui.RemoveButtons(chatID, messageID)
	} else {
		partnerID, er := h.Store.GetPartnerID(ctx, userID)
		if er != nil {
			h.HandleErr(chatID, "Ошибка при попытке получить id партнера", err)
			h.ui.RemoveButtons(chatID, messageID)
			return
		}

		if partnerID == 0 {
			h.Reply(chatID, "Ты уже зарегистрирован! Используй "+string(domain.SetPartner)+", чтобы добавить партнёра 💌")
		} else {
			partnerUsername, err2 := h.Store.GetUsername(ctx, partnerID)
			if err2 != nil {
				h.HandleErr(chatID, "Ошибка при попытке получить username партнера", err2)
				h.ui.RemoveButtons(chatID, messageID)
				return
			}
			text := "Ты уже зарегистрирован! Твой партнер - @" + partnerUsername
			h.Reply(chatID, text)
		}
		h.ui.RemoveButtons(chatID, messageID)
	}
}

func (h *Handler) DeleteAccount(_ context.Context, cq *tgbotapi.CallbackQuery) {
	chatID := cq.Message.Chat.ID
	messageID := cq.Message.MessageID

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
		h.ui.RemoveButtons(chatID, messageID)
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

	case "account:delete:cancel":
		h.Reply(chatID, "Удаление аккаунта отменено ✅")
	}
	h.ui.RemoveButtons(chatID, messageID)
}
