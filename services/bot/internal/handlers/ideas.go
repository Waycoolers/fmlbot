package handlers

import (
	"context"
	"fmt"
	"html"
	"log/slog"
	"regexp"
	"strconv"
	"strings"

	"github.com/Waycoolers/fmlbot/services/bot/internal/domain"
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

type ideaDraftStore interface {
	Get(context.Context, int64) (*domain.IdeaDraft, error)
	Save(context.Context, int64, *domain.IdeaDraft) error
	Delete(context.Context, int64) error
}

type ideaChoice struct {
	label string
	value string
}

var ideaChoices = map[string][]ideaChoice{
	"loc": {{"🏠 Дома", "Дома"}, {"🏙️ В городе", "В городе"}, {"🌲 На природе", "На природе"}},
	"act": {{"🛋️ Спокойно", "Спокойное"}, {"🏃 Активно", "Активное"}, {"🧠 Познавательно", "Познавательное"}, {"💖 Романтично", "Романтичное"}},
	"bud": {{"0️⃣ Бесплатно", "Бесплатно"}, {"💸 Средний", "Средний"}, {"💎 Гуляем на все", "Высокий"}},
}

func (h *Handler) showIdeaStep(chatID int64, draft *domain.IdeaDraft) {
	prompt := map[string]string{
		"loc": "📍 Где вы хотите провести время?",
		"act": "🎭 Какое у вас сейчас настроение?",
		"bud": "💰 Какой планируется бюджет?",
	}[draft.Step]
	var keyboard domain.InlineKeyboard
	for i, choice := range ideaChoices[draft.Step] {
		if i%2 == 0 {
			keyboard.Rows = append(keyboard.Rows, domain.InlineKeyboardRow{})
		}
		row := &keyboard.Rows[len(keyboard.Rows)-1]
		row.Buttons = append(row.Buttons, domain.InlineKeyboardButton{
			Text: choice.label,
			Data: fmt.Sprintf("idea:%s:%s:%d", draft.Step, draft.Session, i),
		})
	}
	keyboard.Rows = append(keyboard.Rows, domain.InlineKeyboardRow{Buttons: []domain.InlineKeyboardButton{
		{Text: "❌ Отмена", Data: fmt.Sprintf("idea:cancel:%s:%s", draft.Session, draft.Step)},
	}})
	if err := h.ui.Client.SendWithInlineKeyboard(chatID, prompt, keyboard); err != nil {
		h.HandleErr(chatID, "Error sending idea keyboard", err)
	}
}

func (h *Handler) GenerateLeisureIdea(ctx context.Context, msg *domain.Message) {
	chatID := msg.ChatID
	if _, busy := h.ideaInFlight.Load(chatID); busy {
		h.Reply(chatID, "⏳ Идея уже готовится.")
		return
	}
	user, err := h.api.GetMe(ctx, chatID)
	if err != nil || user == nil {
		h.HandleErr(chatID, "Error loading user for idea", err)
		return
	}
	if user.PartnerID == 0 {
		h.Reply(chatID, "🤍 Добавь партнёра, и ты сможешь генерировать идеи ✨")
		return
	}

	draft := &domain.IdeaDraft{Session: wizardSession(), Step: "loc"}
	if err := h.ideaDrafts.Save(ctx, chatID, draft); err != nil {
		h.HandleErr(chatID, "Error creating idea draft", err)
		return
	}
	h.showIdeaStep(chatID, draft)
}

func (h *Handler) handleIdea(ctx context.Context, cq *domain.CallbackQuery, action, payload string) {
	chatID := cq.ChatID
	if _, busy := h.ideaInFlight.Load(chatID); busy {
		h.Reply(chatID, "⏳ Идея уже готовится.")
		return
	}
	session, value, valid := strings.Cut(payload, ":")
	draft, err := h.ideaDrafts.Get(ctx, chatID)
	if err != nil {
		h.HandleErr(chatID, "Error loading idea draft", err)
		return
	}
	step := action
	if action == "cancel" {
		step = value
	}
	if !valid || draft == nil || session == "" || draft.Session != session || draft.Step != step {
		h.ui.RemoveButtons(chatID, cq.MessageID)
		h.Reply(chatID, "Эта кнопка устарела. Открой генератор идей заново.")
		return
	}
	if action == "cancel" {
		if err := h.ideaDrafts.Delete(ctx, chatID); err != nil {
			h.HandleErr(chatID, "Error deleting idea draft", err)
			return
		}
		h.ui.RemoveButtons(chatID, cq.MessageID)
		h.Reply(chatID, "😉 Генерация идеи отменена")
		return
	}

	choices := ideaChoices[action]
	i, err := strconv.Atoi(value)
	if err != nil || i < 0 || i >= len(choices) {
		h.ReplyUnknownCallback(ctx, cq)
		return
	}
	choice := choices[i].value
	switch action {
	case "loc":
		draft.Location, draft.Step = choice, "act"
	case "act":
		draft.Activity, draft.Step = choice, "bud"
	case "bud":
		h.generateIdea(ctx, cq, draft, choice)
		return
	}
	if err := h.ideaDrafts.Save(ctx, chatID, draft); err != nil {
		h.HandleErr(chatID, "Error saving idea draft", err)
		return
	}
	h.ui.RemoveButtons(chatID, cq.MessageID)
	h.showIdeaStep(chatID, draft)
}

func (h *Handler) generateIdea(ctx context.Context, cq *domain.CallbackQuery, draft *domain.IdeaDraft, budget string) {
	chatID := cq.ChatID
	select {
	case h.ideaJobs <- struct{}{}:
	default:
		h.Reply(chatID, "Сейчас генератор занят. Попробуй ещё раз чуть позже.")
		return
	}
	h.ideaInFlight.Store(chatID, struct{}{})
	go func() {
		defer func() {
			h.ideaInFlight.Delete(chatID)
			<-h.ideaJobs
		}()
		h.Reply(chatID, "⏳ Придумываю для вас план…")
		idea, err := h.api.GetLeisureIdea(ctx, chatID, draft.Location, draft.Activity, budget, "")
		if ctx.Err() != nil {
			return
		}
		if err != nil || strings.TrimSpace(idea) == "" {
			slog.Warn("Idea generation failed", "error", err)
			h.Reply(chatID, "Не удалось сгенерировать идею. Нажми кнопку бюджета ещё раз, чтобы повторить.")
			return
		}
		result := fmt.Sprintf("🎲 <b>План найден!</b>\n\n📍 %s · 🎭 %s · 💰 %s\n\n%s",
			html.EscapeString(draft.Location), html.EscapeString(draft.Activity), html.EscapeString(budget), formatMarkdownToHTML(idea))
		if err := h.ui.Client.SendMessage(chatID, result); err != nil {
			slog.Error("Error sending generated idea", "error", err)
			return
		}
		if err := h.ideaDrafts.Delete(ctx, chatID); err != nil {
			slog.Error("Error deleting idea draft", "error", err)
		}
		h.ui.RemoveButtons(chatID, cq.MessageID)
	}()
}

var (
	reBold   = regexp.MustCompile(`\*\*(.*?)\*\*`)
	reBullet = regexp.MustCompile(`(?m)^\s*\*\s+`)
)

func formatMarkdownToHTML(input string) string {
	output := reBold.ReplaceAllString(html.EscapeString(input), "<b>$1</b>")
	output = reBullet.ReplaceAllString(output, "• ")

	return output
}
