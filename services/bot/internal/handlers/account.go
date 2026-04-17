package handlers

import (
	"context"
	"errors"
	"log/slog"

	"github.com/Waycoolers/fmlbot/common/errs"
	"github.com/Waycoolers/fmlbot/services/bot/internal/domain"
)

func (h *Handler) ShowAccountMenu(_ context.Context, msg *domain.Message) {
	chatID := msg.ChatID
	text := "⚙️ Здесь можно управлять своим аккаунтом"
	err := h.ui.AccountMenu(chatID, text)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при попытке отобразить меню аккаунтов", err)
		return
	}
}

func (h *Handler) Register(ctx context.Context, msg *domain.Message) {
	chatID := msg.ChatID
	username := msg.UserName

	_, err := h.api.GetMe(ctx, chatID)
	exists := true
	if err != nil {
		if !errors.Is(err, errs.ErrUserNotFound) {
			h.HandleUnknownError(chatID, err)
			return
		}
		exists = false
	}

	if !exists {
		if username == "" {
			h.Reply(chatID, "Сначала установи себе имя пользователя в настройках telegram")
			return
		}

		err = h.api.CreateUser(ctx, chatID, username)
		if err != nil {
			if errors.Is(err, errs.ErrUserExists) {
				h.HandleErr(chatID, "Error user exists", err)
				return
			}
			h.HandleUnknownError(chatID, err)
			return
		}
	}

	h.ShowMainMenu(ctx, msg)
}

func (h *Handler) DeleteAccount(_ context.Context, msg *domain.Message) {
	chatID := msg.ChatID

	keyboard := domain.InlineKeyboard{
		Rows: []domain.InlineKeyboardRow{
			{
				Buttons: []domain.InlineKeyboardButton{
					{Text: "💔 Да, удалить", Data: "account:delete:confirm"},
					{Text: "↩️ Передумал(а)", Data: "account:delete:cancel"},
				},
			},
		},
	}

	text := "💭 Ты уверен(а), что хочешь удалить аккаунт?\n\n" +
		"Все сохранённые данные и тёплые моменты будут удалены без возможности восстановления."

	err := h.ui.Client.SendWithInlineKeyboard(chatID, text, keyboard)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при отправке подтверждения", err)
		return
	}
}

func (h *Handler) HandleDeleteAccount(ctx context.Context, cq *domain.CallbackQuery) {
	chatID := cq.ChatID
	messageID := cq.MessageID

	switch cq.Data {
	case "account:delete:confirm":
		user, err := h.api.GetMe(ctx, chatID)
		if err != nil || user == nil {
			h.ui.RemoveButtons(chatID, messageID)
			if errors.Is(err, errs.ErrUserNotFound) {
				h.HandleErr(chatID, "Error while trying to get user", err)
				return
			}
			h.HandleUnknownError(chatID, err)
			return
		}

		if user.PartnerID != 0 {
			err = h.api.Unpair(ctx, chatID)
			if err != nil {
				h.ui.RemoveButtons(chatID, messageID)
				h.HandleErr(chatID, "Error while trying to delete partners", err)
				return
			}

			err = h.api.ResetPartnerUserConfig(ctx, chatID)
			if err != nil {
				h.ui.RemoveButtons(chatID, messageID)
				h.HandleErr(chatID, "Error resetting config", err)
				return
			}

			h.Reply(user.PartnerID, "Твой партнёр удалил свой аккаунт 💔")
		}

		err = h.api.DeleteMe(ctx, chatID)
		if err != nil {
			h.ui.RemoveButtons(chatID, messageID)
			h.HandleErr(chatID, "Error occurred while trying to delete a user", err)
			return
		}

		h.Reply(chatID, "🕊️ Аккаунт удалён\nЕсли захочешь — я всегда буду рад(а) начать заново")
		text := "✨ Хочешь вернуться?\nНажми кнопку ниже, чтобы начать сначала"
		err = h.ui.StartMenu(chatID, text)
		if err != nil {
			slog.Error("Error calling the start menu", "error", err)
			h.Reply(chatID, "Попробуй перезапустить бота командой /start")
		}
	case "account:delete:cancel":
		h.Reply(chatID, "💛 Хорошо, ничего не удаляем")
	}
	_ = h.ui.Client.DeleteMessage(chatID, messageID)
}
