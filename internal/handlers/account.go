package handlers

import (
	"context"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) ShowAccountMenu(_ context.Context, msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	text := "⚙️ Здесь можно управлять своим аккаунтом"
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

	exists, err := h.Store.Users.IsUserExists(ctx, userID)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при проверке пользователя", err)
		return
	}

	if !exists {
		if username == "" {
			h.Reply(chatID, "Сначала установи себе имя пользователя в настройках telegram")
			return
		}

		er := h.Store.Users.AddUser(ctx, userID, username)
		if er != nil {
			h.HandleErr(chatID, "Ошибка при регистрации", err)
			return
		}
	}

	h.ShowMainMenu(ctx, msg)
}

func (h *Handler) DeleteAccount(_ context.Context, msg *tgbotapi.Message) {
	chatID := msg.Chat.ID

	buttons := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("💔 Да, удалить", "account:delete:confirm"),
			tgbotapi.NewInlineKeyboardButtonData("↩️ Передумал(а)", "account:delete:cancel"),
		),
	)

	text := "💭 Ты уверен(а), что хочешь удалить аккаунт?\n\n" +
		"Все сохранённые данные и тёплые моменты будут удалены без возможности восстановления."

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
		partnerID, err := h.Store.Users.GetPartnerID(ctx, userID)
		if err != nil {
			h.ui.RemoveButtons(chatID, messageID)
			h.HandleErr(chatID, "Ошибка при попытке получить id партнера", err)
			return
		}

		if partnerID != 0 {
			err = h.Store.Users.RemovePartners(ctx, userID, partnerID)
			if err != nil {
				h.ui.RemoveButtons(chatID, messageID)
				h.HandleErr(chatID, "Ошибка при попытке удалить партнеров", err)
				return
			}

			err = h.Store.Users.DeleteUser(ctx, userID)
			if err != nil {
				h.ui.RemoveButtons(chatID, messageID)
				h.HandleErr(chatID, "Ошибка при попытке удалить юзера", err)
				return
			}

			err = h.Store.UserConfig.SetDefault(ctx, partnerID)
			if err != nil {
				h.ui.RemoveButtons(chatID, messageID)
				h.HandleErr(chatID, "Ошибка при сбросе конфига", err)
				return
			}

			h.Reply(partnerID, "Твой партнёр удалил свой аккаунт 💔")
		} else {
			err = h.Store.Users.DeleteUser(ctx, userID)
			if err != nil {
				h.ui.RemoveButtons(chatID, messageID)
				h.HandleErr(chatID, "Ошибка при попытке удалить юзера", err)
				return
			}
		}

		h.Reply(chatID, "🕊️ Аккаунт удалён\nЕсли захочешь — я всегда буду рад(а) начать заново")
		text := "✨ Хочешь вернуться?\nНажми кнопку ниже, чтобы начать сначала"
		err = h.ui.StartMenu(chatID, text)
		if err != nil {
			log.Printf("Ошибка при вызове стартового меню")
			h.Reply(chatID, "Попробуй перезапустить бота командой /start")
		}
	case "account:delete:cancel":
		h.Reply(chatID, "💛 Хорошо, ничего не удаляем")
	}
	_ = h.ui.Client.DeleteMessage(chatID, messageID)
}
