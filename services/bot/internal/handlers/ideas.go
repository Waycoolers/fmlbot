package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"github.com/Waycoolers/fmlbot/services/bot/internal/domain"
	"github.com/Waycoolers/fmlbot/services/bot/internal/state"
)

func (h *Handler) ShowIdeasMenu(_ context.Context, msg *domain.Message) {
	chatID := msg.ChatID
	text := "🦄 <b>Генератор впечатлений</b>\n\n" +
		"Не знаете, как провести время вместе? Быт съедает романтику? " +
		"Я с радостью помогу придумать классную идею для свидания или уютного вечера. ✨"

	err := h.ui.IdeasMenu(chatID, text)
	if err != nil {
		h.HandleErr(chatID, "An error occurred while trying to display the ideas menu.", err)
		return
	}
}

func (h *Handler) GenerateLeisureIdea(ctx context.Context, msg *domain.Message) {
	chatID := msg.ChatID

	user, err := h.api.GetMe(ctx, chatID)
	if err != nil {
		h.HandleErr(chatID, "An error occurred while trying to display the user's information.", err)
		return
	}
	if user == nil {
		h.HandleErr(chatID, "An error occurred while trying to display the user's information.", err)
		return
	}
	if user.PartnerID == 0 {
		h.Reply(chatID, "🤍 Добавь партнёра, и ты сможешь генерировать идеи ✨")
		return
	}

	err = h.ideaDrafts.Save(ctx, chatID, &domain.IdeaDraft{})
	if err != nil {
		h.HandleErr(chatID, "Error creating idea draft", err)
		return
	}

	h.sm.SetStep(chatID, state.AwaitingLocationIdea)

	keyboard := domain.InlineKeyboard{
		Rows: []domain.InlineKeyboardRow{
			{
				Buttons: []domain.InlineKeyboardButton{
					{Text: "🏠 Дома", Data: "idea:loc:Дома"},
					{Text: "🏙️ В городе", Data: "idea:loc:В городе"},
				},
			},
			{
				Buttons: []domain.InlineKeyboardButton{
					{Text: "🌲 На природе", Data: "idea:loc:На природе"},
					{Text: "❌ Отмена", Data: "idea:cancel"},
				},
			},
		},
	}

	err = h.ui.Client.SendWithInlineKeyboard(chatID, "📍 Где вы хотите провести время?", keyboard)
	if err != nil {
		h.HandleErr(chatID, "Error sending location keyboard", err)
	}
}

func (h *Handler) HandleLocationIdeaCallback(ctx context.Context, cq *domain.CallbackQuery) {
	chatID := cq.ChatID
	messageID := cq.MessageID

	if cq.Data == "idea:cancel" {
		_ = h.ideaDrafts.Delete(ctx, chatID)
		h.sm.SetStep(chatID, state.Empty)
		_ = h.ui.Client.DeleteMessage(chatID, messageID)
		h.Reply(chatID, "😉 Генерация идеи отменена")
		return
	}

	location := strings.TrimPrefix(cq.Data, "idea:loc:")

	draft, err := h.ideaDrafts.Get(ctx, chatID)
	if err != nil || draft == nil {
		h.HandleErr(chatID, "Idea session expired", err)
		return
	}

	draft.Location = location
	err = h.ideaDrafts.Save(ctx, chatID, draft)
	if err != nil {
		h.HandleErr(chatID, "Error saving location", err)
		return
	}

	h.sm.SetStep(chatID, state.AwaitingActivityIdea)

	_ = h.ui.Client.DeleteMessage(chatID, messageID)

	keyboard := domain.InlineKeyboard{
		Rows: []domain.InlineKeyboardRow{
			{
				Buttons: []domain.InlineKeyboardButton{
					{Text: "🛋️ Спокойно (чилл)", Data: "idea:act:Спокойное"},
					{Text: "🏃 Активно", Data: "idea:act:Активное"},
				},
			},
			{
				Buttons: []domain.InlineKeyboardButton{
					{Text: "🧠 Познавательно", Data: "idea:act:Познавательное"},
					{Text: "💖 Романтично", Data: "idea:act:Романтичное"},
				},
			},
			{
				Buttons: []domain.InlineKeyboardButton{
					{Text: "❌ Отмена", Data: "idea:cancel"},
				},
			},
		},
	}

	err = h.ui.Client.SendWithInlineKeyboard(chatID, "🎭 Какое у вас сейчас настроение?", keyboard)
	if err != nil {
		h.HandleErr(chatID, "Error sending activity keyboard", err)
	}
}

func (h *Handler) HandleActivityIdeaCallback(ctx context.Context, cq *domain.CallbackQuery) {
	chatID := cq.ChatID
	messageID := cq.MessageID

	if cq.Data == "idea:cancel" {
		_ = h.ideaDrafts.Delete(ctx, chatID)
		h.sm.SetStep(chatID, state.Empty)
		_ = h.ui.Client.DeleteMessage(chatID, messageID)
		h.Reply(chatID, "😉 Генерация идеи отменена")
		return
	}

	activity := strings.TrimPrefix(cq.Data, "idea:act:")

	draft, err := h.ideaDrafts.Get(ctx, chatID)
	if err != nil || draft == nil {
		h.HandleErr(chatID, "Idea session expired", err)
		return
	}

	draft.Activity = activity
	err = h.ideaDrafts.Save(ctx, chatID, draft)
	if err != nil {
		h.HandleErr(chatID, "Error saving activity", err)
		return
	}

	h.sm.SetStep(chatID, state.AwaitingBudgetIdea)

	_ = h.ui.Client.DeleteMessage(chatID, messageID)

	keyboard := domain.InlineKeyboard{
		Rows: []domain.InlineKeyboardRow{
			{
				Buttons: []domain.InlineKeyboardButton{
					{Text: "0️⃣ Бесплатно", Data: "idea:bud:Бесплатно"},
					{Text: "💸 Средний", Data: "idea:bud:Средний"},
				},
			},
			{
				Buttons: []domain.InlineKeyboardButton{
					{Text: "💎 Гуляем на все", Data: "idea:bud:Высокий"},
					{Text: "❌ Отмена", Data: "idea:cancel"},
				},
			},
		},
	}

	err = h.ui.Client.SendWithInlineKeyboard(chatID, "💰 Какой планируется бюджет?", keyboard)
	if err != nil {
		h.HandleErr(chatID, "Error sending budget keyboard", err)
	}
}

func (h *Handler) HandleBudgetIdeaCallback(ctx context.Context, cq *domain.CallbackQuery) {
	chatID := cq.ChatID
	messageID := cq.MessageID

	if cq.Data == "idea:cancel" {
		_ = h.ideaDrafts.Delete(ctx, chatID)
		h.sm.SetStep(chatID, state.Empty)
		_ = h.ui.Client.DeleteMessage(chatID, messageID)
		h.Reply(chatID, "😉 Генерация идеи отменена")
		return
	}

	budget := strings.TrimPrefix(cq.Data, "idea:bud:")

	draft, err := h.ideaDrafts.Get(ctx, chatID)
	if err != nil || draft == nil {
		h.HandleErr(chatID, "Idea session expired", err)
		return
	}

	_ = h.ui.Client.DeleteMessage(chatID, messageID)
	err = h.ui.Client.SendMessage(chatID, "⏳ Придумываю для вас идеальный план... Это займет пару секунд 🪄")
	if err != nil {
		slog.Error("Error sending budget keyboard", "error", err)
	}

	h.sm.SetStep(chatID, state.Empty)
	_ = h.ideaDrafts.Delete(ctx, chatID)

	ideaText, aiErr := h.api.GetLeisureIdea(ctx, chatID, draft.Location, draft.Activity, budget, "")
	if aiErr != nil || ideaText == "" {
		h.HandleErr(chatID, "Failed to get idea from AI service", aiErr)
		h.Reply(chatID, "Ой, мои нейроны запутались 😔 Попробуйте сгенерировать идею еще раз.")
		return
	}

	formattedIdea := formatMarkdownToHTML(ideaText)

	finalMessage := fmt.Sprintf(
		"🎲 <b>Идеальный план найден!</b>\n\n"+
			"📍 <b>Локация:</b> %s\n"+
			"🎭 <b>Настроение:</b> %s\n"+
			"💰 <b>Бюджет:</b> %s\n\n"+
			"%s",
		draft.Location, draft.Activity, budget, formattedIdea,
	)

	h.Reply(chatID, finalMessage)
}

func (h *Handler) HandleCancelIdeaCallback(ctx context.Context, cq *domain.CallbackQuery) {
	chatID := cq.ChatID
	messageID := cq.MessageID

	_ = h.ideaDrafts.Delete(ctx, chatID)
	h.sm.SetStep(chatID, state.Empty)

	err := h.ui.Client.DeleteMessage(chatID, messageID)
	if err != nil {
		h.HandleErr(chatID, "Error deleting message", err)
	}

	h.Reply(chatID, "😉 Генерация идеи отменена")
}

func formatMarkdownToHTML(input string) string {
	reBold := regexp.MustCompile(`\*\*(.*?)\*\*`)
	output := reBold.ReplaceAllString(input, "<b>$1</b>")

	reBullet := regexp.MustCompile(`(?m)^\s*\*\s+`)
	output = reBullet.ReplaceAllString(output, "• ")

	return output
}
