package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/Waycoolers/fmlbot/pkg/errs"
	"github.com/Waycoolers/fmlbot/services/bot/internal/domain"
	"github.com/Waycoolers/fmlbot/services/bot/internal/state"
)

func (h *Handler) SendMessage(w http.ResponseWriter, r *http.Request) {
	var message domain.MessageRequest
	err := json.NewDecoder(r.Body).Decode(&message)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	h.Reply(message.UserID, message.Text)
}

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

		password, err := h.api.CreateUser(ctx, chatID, username)
		if err != nil {
			if errors.Is(err, errs.ErrUserExists) {
				h.HandleErr(chatID, "Error user exists", err)
				return
			}
			h.HandleUnknownError(chatID, err)
			return
		}
		text := fmt.Sprintf("Ты успешно зарегистрировался в боте!\nТвой пароль: %s\nP.S. Он тебе может пригодится для входа в мобильное приложение. Ты всегда можешь поменять пароль в настройках аккаунта.", password)
		h.Reply(chatID, text)
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

func (h *Handler) ChangePassword(_ context.Context, msg *domain.Message) {
	chatID := msg.ChatID

	h.sm.SetStep(chatID, state.AwaitingPassword)

	text := "Отправь новый пароль"
	h.Reply(chatID, text)
}

func (h *Handler) HandleChangePassword(ctx context.Context, msg *domain.Message) {
	chatID := msg.ChatID
	password := msg.Text

	text := ""
	err := validatePassword(password)
	if err != nil {
		switch {
		case errors.Is(err, ErrPasswordTooShort):
			text = "Пароль должен содержать минимум 8 символов"
		case errors.Is(err, ErrPasswordTooLong):
			text = "Пароль не может быть длиннее 32 символов"
		case errors.Is(err, ErrPasswordInvalidCharacter):
			text = "Пароль может состоять из латинских букв, цифр и знаков препинания"
		case errors.Is(err, ErrPasswordWithoutLetter):
			text = "В пароле обязательно должна быть хотя бы одна буква"
		case errors.Is(err, ErrPasswordWithoutUpper):
			text = "В пароле обязательно должен быть хотя бы один символ с верхним регистром"
		case errors.Is(err, ErrPasswordWithoutLower):
			text = "В пароле обязательно должен быть хотя бы один символ с нижним регистром"
		case errors.Is(err, ErrPasswordWithoutDigit):
			text = "В пароле обязательно должна быть хотя бы одна цифра"
		}
		h.Reply(chatID, text)
		return
	}

	err = h.api.ChangePassword(ctx, chatID, password)
	if err != nil {
		h.sm.SetStep(chatID, state.Empty)
		h.HandleErr(chatID, "Error occurred while trying to change password", err)
		return
	}
	text = "Пароль успешно изменен!"
	h.sm.SetStep(chatID, state.Empty)
	h.Reply(chatID, text)
}

var (
	ErrPasswordTooShort         = errors.New("password too short")
	ErrPasswordTooLong          = errors.New("password too long")
	ErrPasswordInvalidCharacter = errors.New("password invalid")
	ErrPasswordWithoutLetter    = errors.New("password without letter")
	ErrPasswordWithoutUpper     = errors.New("password without uppercase")
	ErrPasswordWithoutLower     = errors.New("password without lowercase")
	ErrPasswordWithoutDigit     = errors.New("password without digit")
)

func validatePassword(password string) error {
	if len([]rune(password)) < 8 {
		return ErrPasswordTooShort
	}
	if len([]rune(password)) > 32 {
		return ErrPasswordTooLong
	}

	var hasLetter, hasUpper, hasLower, hasNumber bool

	for _, char := range password {
		if char < '!' || char > '~' {
			return ErrPasswordInvalidCharacter
		}

		switch {
		case char >= 'A' && char <= 'Z':
			hasLetter = true
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLetter = true
			hasLower = true
		case char >= '0' && char <= '9':
			hasNumber = true
		}
	}

	if !hasLetter {
		return ErrPasswordWithoutLetter
	}
	if !hasUpper {
		return ErrPasswordWithoutUpper
	}
	if !hasLower {
		return ErrPasswordWithoutLower
	}
	if !hasNumber {
		return ErrPasswordWithoutDigit
	}
	return nil
}
