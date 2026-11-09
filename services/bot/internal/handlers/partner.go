package handlers

import (
	"context"
	"errors"
	"fmt"
	"html"
	"log/slog"
	"strconv"
	"strings"

	"github.com/Waycoolers/fmlbot/pkg/errs"
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

		text = "💞 Твой партнёр: @" + html.EscapeString(partner.Username)
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
		h.sm.SetStep(chatID, state.AwaitingPartner)
		h.Reply(chatID, "💌 Отправь имя аккаунта партнёра в FMLBot")
	} else {
		partner, err := h.api.GetPartner(ctx, chatID)
		if err != nil || partner == nil {
			h.HandleErr(chatID, "Error getting partner", err)
			return
		}
		h.Reply(
			chatID,
			"💞 Сейчас твой партнёр — @"+html.EscapeString(partner.Username)+
				"\nЕсли хочешь изменить выбор, сначала нужно удалить текущего партнёра",
		)
	}
}

func (h *Handler) ProcessPartnerUsername(ctx context.Context, msg *domain.Message) {
	chatID := msg.ChatID
	partnerUsername := msg.Text

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

	if user.UserID == partner.UserID {
		h.Reply(chatID, "😅 Так не получится — себя добавить нельзя")
		return
	}
	if user.PartnerID != 0 {
		if user.PartnerID == partner.UserID {
			h.Reply(chatID, "💛 @"+html.EscapeString(partner.Username)+" и так ваш партнёр. Приятного времяпрепровождения!")
			h.sm.SetStep(chatID, state.Empty)
			return
		} else {
			h.Reply(chatID, "😔 У этого пользователя уже есть партнёр")
			h.sm.SetStep(chatID, state.Empty)
			return
		}
	}

	h.sm.SetStep(chatID, state.Empty)

	requestID, err := h.api.RequestPartner(ctx, chatID, partner.UserID)
	if errors.Is(err, errs.ErrPartnerNotFound) {
		h.HandleErr(chatID, "User with this ID not found", err)
		return
	} else if errors.Is(err, errs.ErrPartnerAlreadyHasPartner) {
		h.Reply(chatID, "😔 У этого пользователя уже есть партнёр")
		return
	} else if err != nil {
		h.HandleErr(chatID, "Error when requesting partner", err)
		return
	}

	keyboard := domain.InlineKeyboard{
		Rows: []domain.InlineKeyboardRow{
			{
				Buttons: []domain.InlineKeyboardButton{
					{Text: "💞 Принять", Data: fmt.Sprintf("partner:request:accept:%d", requestID)},
					{Text: "💔 Отклонить", Data: fmt.Sprintf("partner:request:decline:%d", requestID)},
				},
			},
		},
	}
	text := "💌 @" + html.EscapeString(user.Username) + " хочет стать твоим партнёром. Принять заявку?"
	partnerChatID, err := h.api.TelegramChatID(ctx, partner.UserID)
	if err == nil {
		err = h.ui.Client.SendWithInlineKeyboard(partnerChatID, text, keyboard)
	}
	if err != nil && !errors.Is(err, errs.ErrUserNotFound) {
		slog.Error("Error sending partner request", "error", err)
	}

	h.Reply(chatID, fmt.Sprintf("💌 Заявка отправлена @%s. Как только партнёр её примет, я напишу тебе", html.EscapeString(partner.Username)))
}

func (h *Handler) HandlePartnerRequest(ctx context.Context, cq *domain.CallbackQuery, payload string) {
	chatID := cq.ChatID

	answer, idText, found := strings.Cut(payload, ":")
	requestID, err := strconv.ParseInt(idText, 10, 64)
	if !found || err != nil || (answer != "accept" && answer != "decline") {
		h.ReplyUnknownCallback(ctx, cq)
		return
	}

	username := html.EscapeString(cq.UserName)
	if answer == "decline" {
		fromUserID, err := h.api.DeclinePartnerRequest(ctx, chatID, requestID)
		if errors.Is(err, errs.ErrPartnerRequestNotFound) {
			h.ui.RemoveButtons(chatID, cq.MessageID)
			h.Reply(chatID, "🤍 Эта заявка уже неактуальна")
			return
		} else if err != nil {
			h.HandleErr(chatID, "Error when declining partner request", err)
			return
		}
		h.ui.RemoveButtons(chatID, cq.MessageID)
		h.Reply(chatID, "🕊️ Заявка отклонена")
		h.ReplyToUser(ctx, fromUserID, "💔 @"+username+" отклонил(а) твою заявку")
		return
	}

	fromUserID, err := h.api.AcceptPartnerRequest(ctx, chatID, requestID)
	switch {
	case errors.Is(err, errs.ErrPartnerRequestNotFound):
		h.ui.RemoveButtons(chatID, cq.MessageID)
		h.Reply(chatID, "🤍 Эта заявка уже неактуальна")
		return
	case errors.Is(err, errs.ErrPartnerAlreadyHasPartner):
		h.ui.RemoveButtons(chatID, cq.MessageID)
		h.Reply(chatID, "😔 У этого пользователя уже есть партнёр")
		return
	case errors.Is(err, errs.ErrAlreadyHasPartner):
		h.ui.RemoveButtons(chatID, cq.MessageID)
		h.Reply(chatID, "💛 У тебя уже есть партнёр")
		return
	case err != nil:
		h.HandleErr(chatID, "Error when accepting partner request", err)
		return
	}

	h.ui.RemoveButtons(chatID, cq.MessageID)
	partner, err := h.api.GetPartner(ctx, chatID)
	if err != nil || partner == nil {
		h.HandleErr(chatID, "Error getting partner", err)
		return
	}
	h.Reply(chatID, "💞 У вас с @"+html.EscapeString(partner.Username)+" теперь есть общая история в боте ✨")
	h.ReplyToUser(ctx, fromUserID, "✨ Готово! Партнёр @"+username+" добавлен")
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
					{Text: "💔 Да, удалить", Data: fmt.Sprintf("partner:delete:confirm:%d:%d", user.PartnerID, user.PairVersion)},
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

	text := "💭 Ты уверен(а), что хочешь удалить партнёра @" + html.EscapeString(partner.Username) + "?\n" +
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

	if cb.Data == "partner:delete:cancel" {
		h.ui.RemoveButtons(chatID, messageID)
		h.Reply(chatID, "💛 Хорошо, ничего не меняем")
		return
	}

	payload, valid := strings.CutPrefix(cb.Data, "partner:delete:confirm:")
	partnerText, versionText, split := strings.Cut(payload, ":")
	partnerID, partnerErr := strconv.ParseInt(partnerText, 10, 64)
	version, versionErr := strconv.ParseInt(versionText, 10, 64)
	if !valid || !split || partnerErr != nil || versionErr != nil || partnerID <= 0 || version < 0 {
		h.ui.RemoveButtons(chatID, messageID)
		h.Reply(chatID, "Это подтверждение устарело. Открой меню партнёра заново.")
		return
	}
	user, err := h.api.GetMe(ctx, chatID)
	if err != nil || user == nil {
		h.HandleErr(chatID, "Error getting user", err)
		return
	}
	if user.PartnerID != partnerID || user.PairVersion != version {
		h.ui.RemoveButtons(chatID, messageID)
		h.Reply(chatID, "Это подтверждение устарело. Открой меню партнёра заново.")
		return
	}
	if err = h.api.Unpair(ctx, chatID, version); err != nil {
		if errors.Is(err, errs.ErrPairChanged) || errors.Is(err, errs.ErrPartnerNotFound) {
			h.ui.RemoveButtons(chatID, messageID)
			h.Reply(chatID, "Это подтверждение устарело. Открой меню партнёра заново.")
			return
		}
		h.HandleErr(chatID, "Error while trying to delete partners", err)
		return
	}
	h.ui.RemoveButtons(chatID, messageID)
	h.Reply(chatID, "🕊️ Партнёр удалён")
	h.ReplyToUser(ctx, partnerID, "💔 Твой партнёр больше не связан с тобой")
}
