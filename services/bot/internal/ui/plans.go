package ui

import (
	"fmt"
	"html"
	"strings"
	"time"

	"github.com/Waycoolers/fmlbot/pkg/plans"
	"github.com/Waycoolers/fmlbot/services/bot/internal/domain"
)

const planTimeLayout = "02.01.2006 15:04"

var planTimeZone = time.FixedZone("МСК", 3*60*60)

func planCardText(card plans.Card) string {
	var text strings.Builder
	fmt.Fprintf(&text, "<b>%s</b>\n⏱ %d мин. · 💰 На двоих примерно %s\n\n<b>Подготовка</b>\n%s\n\n<b>Что делать</b>\n",
		html.EscapeString(card.Title), card.DurationMinutes, html.EscapeString(card.EstimatedCost),
		html.EscapeString(card.Preparation))
	for i, step := range card.Steps {
		fmt.Fprintf(&text, "%d. %s\n", i+1, html.EscapeString(step))
	}
	fmt.Fprintf(&text, "\n<b>Запасной вариант</b>\n%s", html.EscapeString(card.Backup))
	return text.String()
}

func (ui *MenuUI) SendPlanOptions(chatID int64, session string, options []plans.Card) error {
	for i, card := range options {
		keyboard := domain.InlineKeyboard{Rows: []domain.InlineKeyboardRow{
			{Buttons: []domain.InlineKeyboardButton{
				{Text: "💌 Предложить партнёру", Data: fmt.Sprintf("plan:offer:%s:%d", session, i)},
			}},
			{Buttons: []domain.InlineKeyboardButton{
				{Text: "Проще", Data: fmt.Sprintf("plan:simpler:%s:%d", session, i)},
				{Text: "Поинтереснее", Data: fmt.Sprintf("plan:more_involved:%s:%d", session, i)},
			}},
			{Buttons: []domain.InlineKeyboardButton{
				{Text: "Дешевле", Data: fmt.Sprintf("plan:cheaper:%s:%d", session, i)},
				{Text: "Подороже", Data: fmt.Sprintf("plan:more_expensive:%s:%d", session, i)},
			}},
		}}
		text := fmt.Sprintf("Вариант %d из %d\n\n%s", i+1, len(options), planCardText(card))
		if err := ui.Client.SendWithInlineKeyboard(chatID, text, keyboard); err != nil {
			return err
		}
	}
	return nil
}

func (ui *MenuUI) SendPlansList(chatID int64, items []plans.Plan) error {
	if len(items) == 0 {
		return ui.Client.SendMessage(chatID, "🗓 Пока нет предстоящих планов. Начни с кнопки «План для двоих» в меню «Идеи».")
	}
	var keyboard domain.InlineKeyboard
	for _, plan := range items {
		icon := "⏳"
		if plan.Status == plans.Confirmed {
			icon = "✅"
		}
		label := fmt.Sprintf("%s %s · %s", icon, plan.StartsAt.In(planTimeZone).Format("02.01 15:04"), plan.Card.Title)
		keyboard.Rows = append(keyboard.Rows, domain.InlineKeyboardRow{Buttons: []domain.InlineKeyboardButton{
			{Text: label, Data: fmt.Sprintf("plan:view:%d", plan.ID)},
		}})
	}
	return ui.Client.SendWithInlineKeyboard(chatID, "🗓 <b>Ближайшие планы</b>\n⏳ На согласовании · ✅ Подтверждён", keyboard)
}

func (ui *MenuUI) SendPlan(chatID int64, view plans.View, reminder bool) error {
	heading := "⏳ Ждём ответа партнёра"
	switch view.Status {
	case plans.Pending:
		if view.ViewerID != view.ProposedBy {
			heading = "💌 Партнёр предложил план"
			if view.Revision > 1 {
				heading = "🕓 Партнёр предложил другое время"
			}
		}
	case plans.Confirmed:
		heading = "✅ План подтверждён"
	case plans.Declined:
		heading = "🤍 Предложение отклонено"
	case plans.Cancelled:
		heading = "🤍 План отменён"
	}
	if reminder {
		heading = "⏰ Скоро ваш план"
	}
	upcoming := view.StartsAt.After(time.Now())
	if view.Status == plans.Pending && !upcoming {
		heading = "⌛ Время предложения истекло"
	}
	text := fmt.Sprintf("<b>%s</b>\n🗓 %s МСК\n", heading, view.StartsAt.In(planTimeZone).Format(planTimeLayout))
	if view.Status == plans.Confirmed && upcoming && !reminder {
		text += "Напомню вам обоим перед началом.\n"
	}
	text += "\n" + planCardText(view.Card)

	var keyboard domain.InlineKeyboard
	if upcoming {
		if view.Can(view.ViewerID, plans.Accept) {
			keyboard.Rows = append(keyboard.Rows, domain.InlineKeyboardRow{Buttons: []domain.InlineKeyboardButton{
				planButton("✅ Принять", plans.Accept, view),
				planButton("🕓 Другое время", plans.Reschedule, view),
			}}, domain.InlineKeyboardRow{Buttons: []domain.InlineKeyboardButton{
				planButton("Отклонить", plans.Decline, view),
			}})
		} else if view.Can(view.ViewerID, plans.Cancel) {
			keyboard.Rows = append(keyboard.Rows, domain.InlineKeyboardRow{Buttons: []domain.InlineKeyboardButton{
				planButton("Отменить план", plans.Cancel, view),
			}})
		}
	}
	if len(keyboard.Rows) == 0 {
		return ui.Client.SendMessage(chatID, text)
	}
	return ui.Client.SendWithInlineKeyboard(chatID, text, keyboard)
}

func planButton(text string, action plans.Action, view plans.View) domain.InlineKeyboardButton {
	return domain.InlineKeyboardButton{
		Text: text,
		Data: fmt.Sprintf("plan:%s:%d:%d", action, view.ID, view.Revision),
	}
}
