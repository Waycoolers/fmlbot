package handlers

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Waycoolers/fmlbot/common/errs"
	"github.com/Waycoolers/fmlbot/services/bot/internal/domain"
	"github.com/Waycoolers/fmlbot/services/bot/internal/state"
)

func (h *Handler) ShowPartnerMenu(ctx context.Context, msg *domain.Message) {
	chatID := msg.ChatID
	text := "👤 Партнёр"

	user, err := h.api.GetMe(ctx, chatID)
	if err != nil || user == nil {
		h.HandleErr(chatID, "Error getting user", err)
		return
	}

	if user.PartnerID == 0 {
		text = "🤍 У тебя пока нет партнёра"
	} else {
		partner, err := h.api.GetPartner(ctx, chatID)
		if err != nil || partner == nil {
			h.HandleErr(chatID, "Error getting partner", err)
			return
		}

		text = "💞 Твой партнёр: @" + partner.Username
	}

	err = h.ui.PartnerMenu(chatID, text)
	if err != nil {
		h.HandleErr(chatID, "Error while trying to display the partners menu", err)
		return
	}
}

func (h *Handler) SetPartner(ctx context.Context, msg *domain.Message) {
	chatID := msg.ChatID

	user, err := h.api.GetMe(ctx, chatID)
	if err != nil || user == nil {
		h.HandleErr(chatID, "Error getting user", err)
		return
	}

	if user.PartnerID == 0 {
		h.sm.SetStep(state.AwaitingPartner)
		h.Reply(chatID, "💌 Отправь username партнёра")
	} else {
		partner, err := h.api.GetPartner(ctx, chatID)
		if err != nil || partner == nil {
			h.HandleErr(chatID, "Error getting partner", err)
			return
		}
		h.Reply(
			chatID,
			"💞 Сейчас твой партнёр — @"+partner.Username+
				"\nЕсли хочешь изменить выбор, сначала нужно удалить текущего партнёра",
		)
	}
}

func (h *Handler) ProcessPartnerUsername(ctx context.Context, msg *domain.Message) {
	chatID := msg.ChatID
	partnerUsername := msg.Text
	userUsername := msg.UserName

	if strings.HasPrefix(partnerUsername, "@") {
		partnerUsername = partnerUsername[1:]
	}

	partner, err := h.api.GetUserByUsername(ctx, chatID, partnerUsername)
	exists := true
	if err != nil || partner == nil {
		if !errors.Is(err, errs.ErrUserNotFound) {
			h.HandleErr(chatID, "Error getting user", err)
			return
		}
		exists = false
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

	user, err := h.api.GetMe(ctx, chatID)
	if err != nil || user == nil {
		h.HandleErr(chatID, "Error getting user", err)
		return
	}

	if user.PartnerID != 0 {
		if user.PartnerID == partner.UserID {
			h.Reply(chatID, "💛 @"+partner.Username+" и так ваш партнёр. Приятного времяпрепровождения!")
			h.sm.SetStep(state.Empty)
			return
		} else {
			h.Reply(chatID, "😔 У этого пользователя уже есть партнёр")
			h.sm.SetStep(state.Empty)
			return
		}
	}

	if user.PartnerID != 0 {
		err = h.api.Unpair(ctx, chatID)
		if err != nil {
			h.HandleErr(chatID, "Error when resetting partners", err)
			return
		}
		h.Reply(partner.UserID, "💔 Твой партнёр добавил другого партнёра")
	}

	h.sm.SetStep(state.Empty)

	err = h.api.PairUsers(ctx, chatID, partner.UserID)
	if errors.Is(err, errs.ErrPartnerNotFound) {
		h.HandleErr(chatID, "User with this ID not found", err)
		return
	} else if err != nil {
		h.HandleErr(chatID, "Error when resetting partner at partner", err)
		return
	}

	h.Reply(partner.UserID, "💞 У вас с @"+userUsername+" теперь есть общая история в боте ✨")
	h.Reply(chatID, fmt.Sprintf("✨ Готово! Партнёр @%s добавлен", partner.Username))
}

func (h *Handler) DeletePartner(ctx context.Context, msg *domain.Message) {
	chatID := msg.ChatID

	user, err := h.api.GetMe(ctx, chatID)
	if err != nil || user == nil {
		h.HandleErr(chatID, "Error getting user", err)
		return
	}

	if user.PartnerID == 0 {
		h.Reply(chatID, "🤍 У тебя сейчас не добавлен партнёр")
		return
	}

	keyboard := domain.InlineKeyboard{
		Rows: []domain.InlineKeyboardRow{
			{
				Buttons: []domain.InlineKeyboardButton{
					{Text: "💔 Да, удалить", Data: "partner:delete:confirm"},
					{Text: "↩️ Передумал(а)", Data: "partner:delete:cancel"},
				},
			},
		},
	}

	partner, err := h.api.GetPartner(ctx, chatID)
	if err != nil || partner == nil {
		h.HandleErr(chatID, "Error getting partner", err)
		return
	}

	text := "💭 Ты уверен(а), что хочешь удалить партнёра @" + partner.Username + "?\n" +
		"Все общие настройки будут сброшены."

	err = h.ui.Client.SendWithInlineKeyboard(chatID, text, keyboard)
	if err != nil {
		h.HandleErr(chatID, "Ошибка при отправке подтверждения", err)
		return
	}
}

func (h *Handler) HandleDeletePartner(ctx context.Context, cb *domain.CallbackQuery) {
	chatID := cb.ChatID
	messageID := cb.MessageID

	switch cb.Data {
	case "partner:delete:confirm":
		partner, err := h.api.GetPartner(ctx, chatID)
		if err != nil || partner == nil {
			h.ui.RemoveButtons(chatID, messageID)
			h.HandleErr(chatID, "Error getting partner", err)
			return
		}

		err = h.api.ResetMyUserConfig(ctx, chatID)
		if err != nil {
			h.ui.RemoveButtons(chatID, messageID)
			h.HandleErr(chatID, "Error resetting user config", err)
			return
		}
		err = h.api.ResetPartnerUserConfig(ctx, chatID)
		if err != nil {
			h.ui.RemoveButtons(chatID, messageID)
			h.HandleErr(chatID, "Error resetting partner user config", err)
			return
		}

		err = h.api.Unpair(ctx, chatID)
		if err != nil {
			h.ui.RemoveButtons(chatID, messageID)
			h.HandleErr(chatID, "Error while trying to delete partners", err)
			return
		}

		h.Reply(chatID, "🕊️ Партнёр удалён")
		h.Reply(partner.UserID, "💔 Твой партнёр больше не связан с тобой")

	case "partner:delete:cancel":
		h.Reply(chatID, "💛 Хорошо, ничего не меняем")
	}
	h.ui.RemoveButtons(chatID, messageID)
}
