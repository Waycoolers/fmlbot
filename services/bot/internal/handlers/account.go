package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"log/slog"
	"net/http"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/Waycoolers/fmlbot/pkg/errs"
	pwd "github.com/Waycoolers/fmlbot/pkg/password"
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

	h.sendUpdate(w, message.UserID, message.Text)
}

func (h *Handler) sendUpdate(w http.ResponseWriter, chatID int64, text string) {
	if err := h.ui.Client.SendMessage(chatID, text); err != nil {
		slog.Error("Error sending message", "error", err)
		if errors.Is(err, domain.ErrRecipientUnavailable) {
			w.WriteHeader(http.StatusGone)
			return
		}
		w.WriteHeader(http.StatusServiceUnavailable)
	}
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

	if exists {
		h.ShowMainMenu(ctx, msg)
		return
	}
	if username == "" {
		h.sm.SetStep(chatID, state.AwaitingRegistrationUsername)
		h.Reply(chatID, "Придумай имя для FMLBot и отправь его одним сообщением.")
		return
	}
	h.registerTelegram(ctx, msg, username)
}

func (h *Handler) RegisterWithUsername(ctx context.Context, msg *domain.Message) {
	username := strings.TrimPrefix(strings.TrimSpace(msg.Text), "@")
	if username == "" || utf8.RuneCountInString(username) > 32 || strings.IndexFunc(username, unicode.IsSpace) >= 0 {
		h.Reply(msg.ChatID, "Имя должно быть без пробелов и не длиннее 32 символов.")
		return
	}
	h.registerTelegram(ctx, msg, username)
}

func (h *Handler) registerTelegram(ctx context.Context, msg *domain.Message, username string) {
	chatID := msg.ChatID
	password, err := h.api.CreateUser(ctx, chatID, username)
	switch {
	case errors.Is(err, errs.ErrUsernameIsAlreadyTaken):
		h.sm.SetStep(chatID, state.AwaitingRegistrationUsername)
		h.Reply(chatID, "Имя уже занято. Придумай другое или привяжи свой существующий аккаунт командой /link КОД.")
		return
	case err != nil:
		h.HandleUnknownError(chatID, err)
		return
	}
	h.accountDeleteConfirmations.Delete(chatID)
	h.sm.SetStep(chatID, state.Empty)
	text := fmt.Sprintf("Ты успешно зарегистрировался в боте!\nТвоё имя в FMLBot: %s\nТвой пароль: %s\nОни пригодятся для входа в мобильное приложение.", html.EscapeString(username), html.EscapeString(password))
	h.Reply(chatID, text)
	h.ShowMainMenu(ctx, msg)
}

func (h *Handler) LinkTelegram(ctx context.Context, msg *domain.Message, code string) {
	switch err := h.api.LinkTelegram(ctx, msg.ChatID, code); {
	case errors.Is(err, errs.ErrTelegramLinkInvalid):
		h.Reply(msg.ChatID, "Код недействителен или истёк. Получи новый код в своём аккаунте.")
	case errors.Is(err, errs.ErrTelegramAlreadyLinked):
		h.Reply(msg.ChatID, "Этот аккаунт уже связан с Telegram.")
	case err != nil:
		h.HandleErr(msg.ChatID, "Error linking Telegram", err)
	default:
		h.accountDeleteConfirmations.Delete(msg.ChatID)
		h.sm.SetStep(msg.ChatID, state.Empty)
		h.Reply(msg.ChatID, "Telegram привязан к твоему аккаунту.")
		h.ShowMainMenu(ctx, msg)
	}
}

func (h *Handler) DeleteAccount(_ context.Context, msg *domain.Message) {
	chatID := msg.ChatID
	session := wizardSession()
	h.accountDeleteConfirmations.Store(chatID, session)

	keyboard := domain.InlineKeyboard{
		Rows: []domain.InlineKeyboardRow{
			{
				Buttons: []domain.InlineKeyboardButton{
					{Text: "💔 Да, удалить", Data: "account:delete:confirm:" + session},
					{Text: "↩️ Передумал(а)", Data: "account:delete:cancel:" + session},
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

	action, session, valid := strings.Cut(strings.TrimPrefix(cq.Data, "account:delete:"), ":")
	current, _ := h.accountDeleteConfirmations.Load(chatID)
	if !valid || session != current {
		h.ui.RemoveButtons(chatID, messageID)
		h.Reply(chatID, "Это подтверждение устарело. Открой настройки аккаунта заново.")
		return
	}
	switch action {
	case "confirm":
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

		// API сам разрывает пару и удаляет данные одной транзакцией
		err = h.api.DeleteMe(ctx, chatID)
		if err != nil {
			h.ui.RemoveButtons(chatID, messageID)
			h.HandleErr(chatID, "Error occurred while trying to delete a user", err)
			return
		}

		h.accountDeleteConfirmations.Delete(chatID)
		if user.PartnerID != 0 {
			h.ReplyToUser(ctx, user.PartnerID, "Твой партнёр удалил свой аккаунт 💔")
		}

		h.Reply(chatID, "🕊️ Аккаунт удалён\nЕсли захочешь — я всегда буду рад(а) начать заново")
		text := "✨ Хочешь вернуться?\nНажми кнопку ниже, чтобы начать сначала"
		err = h.ui.StartMenu(chatID, text)
		if err != nil {
			slog.Error("Error calling the start menu", "error", err)
			h.Reply(chatID, "Попробуй перезапустить бота командой /start")
		}
	case "cancel":
		h.accountDeleteConfirmations.Delete(chatID)
		h.Reply(chatID, "💛 Хорошо, ничего не удаляем")
	}
	_ = h.ui.Client.DeleteMessage(chatID, messageID)
}

func (h *Handler) ChangePassword(_ context.Context, msg *domain.Message) {
	chatID := msg.ChatID

	h.oldPasswords.Delete(chatID)
	h.sm.SetStep(chatID, state.AwaitingOldPassword)

	keyboard := domain.InlineKeyboard{
		Rows: []domain.InlineKeyboardRow{
			{
				Buttons: []domain.InlineKeyboardButton{
					{Text: "🔑 Не помню пароль", Data: "account:password:forgot"},
				},
			},
		},
	}
	err := h.ui.Client.SendWithInlineKeyboard(chatID, "Отправь текущий пароль", keyboard)
	if err != nil {
		h.sm.SetStep(chatID, state.Empty)
		h.HandleErr(chatID, "Error sending password request", err)
	}
}

func (h *Handler) HandleOldPassword(_ context.Context, msg *domain.Message) {
	chatID := msg.ChatID

	h.oldPasswords.Store(chatID, msg.Text)
	h.sm.SetStep(chatID, state.AwaitingPassword)
	h.Reply(chatID, "Отправь новый пароль")
}

func (h *Handler) HandleForgotPassword(ctx context.Context, cq *domain.CallbackQuery) {
	chatID := cq.ChatID

	h.ui.RemoveButtons(chatID, cq.MessageID)
	if h.sm.GetStep(chatID) != state.AwaitingOldPassword {
		h.ReplyUnknownCallback(ctx, cq)
		return
	}
	h.sm.SetStep(chatID, state.Empty)

	password, err := h.api.ResetPassword(ctx, chatID)
	if err != nil {
		h.HandleErr(chatID, "Error occurred while trying to reset password", err)
		return
	}
	h.Reply(chatID, fmt.Sprintf("Твой новый пароль: %s\nТы всегда можешь поменять пароль в настройках аккаунта.", html.EscapeString(password)))
}

func (h *Handler) HandleChangePassword(ctx context.Context, msg *domain.Message) {
	chatID := msg.ChatID
	password := msg.Text

	text := ""
	err := pwd.Validate(password)
	if err != nil {
		switch {
		case errors.Is(err, errs.ErrPasswordTooShort):
			text = "Пароль должен содержать минимум 8 символов"
		case errors.Is(err, errs.ErrPasswordTooLong):
			text = "Пароль не может быть длиннее 32 символов"
		case errors.Is(err, errs.ErrPasswordInvalidCharacter):
			text = "Пароль может состоять из латинских букв, цифр и знаков препинания"
		case errors.Is(err, errs.ErrPasswordWithoutLetter):
			text = "В пароле обязательно должна быть хотя бы одна буква"
		case errors.Is(err, errs.ErrPasswordWithoutUpper):
			text = "В пароле обязательно должен быть хотя бы один символ с верхним регистром"
		case errors.Is(err, errs.ErrPasswordWithoutLower):
			text = "В пароле обязательно должен быть хотя бы один символ с нижним регистром"
		case errors.Is(err, errs.ErrPasswordWithoutDigit):
			text = "В пароле обязательно должна быть хотя бы одна цифра"
		}
		h.Reply(chatID, text)
		return
	}

	oldPassword, _ := h.oldPasswords.LoadAndDelete(chatID)
	currentPassword, _ := oldPassword.(string)

	h.sm.SetStep(chatID, state.Empty)
	err = h.api.ChangePassword(ctx, chatID, currentPassword, password)
	if err != nil {
		if errors.Is(err, errs.ErrWrongPassword) {
			h.Reply(chatID, "Текущий пароль указан неверно. Попробуй сменить пароль ещё раз")
			return
		}
		h.HandleErr(chatID, "Error occurred while trying to change password", err)
		return
	}
	text = "Пароль успешно изменен!"
	h.sm.SetStep(chatID, state.Empty)
	h.Reply(chatID, text)
}
