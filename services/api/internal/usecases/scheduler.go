package usecases

import (
	"context"
	"fmt"
	"html"
	"log/slog"
	"time"

	"github.com/Waycoolers/fmlbot/services/api/internal/domain"
)

const midnightTask = "midnight"

func (uc *UseCase) DoMidnightTasks(ctx context.Context) error {
	if err := uc.scheduler.DoMidnightTasksWithCompliments(ctx); err != nil {
		return err
	}
	return uc.scheduler.SetLastRun(ctx, midnightTask, time.Now())
}

// RunMissedMidnightTasks выполняет ночные задачи, если сервис не работал в полночь.
func (uc *UseCase) RunMissedMidnightTasks(ctx context.Context, loc *time.Location) error {
	lastRun, found, err := uc.scheduler.GetLastRun(ctx, midnightTask)
	if err != nil {
		return err
	}
	now := time.Now().In(loc)
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	if found && !lastRun.Before(midnight) {
		return nil
	}
	return uc.DoMidnightTasks(ctx)
}

func dayOf(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

// occurrenceIn возвращает день годовщины в указанном году. 29 февраля в невисокосный год
// отмечается 28 февраля.
func occurrenceIn(date time.Time, year int) time.Time {
	day := time.Date(year, date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	if day.Month() != date.Month() {
		return time.Date(year, date.Month(), 28, 0, 0, 0, 0, time.UTC)
	}
	return day
}

// Важные даты повторяются каждый год, а год самой даты только хранит, когда событие случилось.
func isEventAndNotifyDay(date time.Time, notifyBeforeDays int, today time.Time) (isEventDay, isNotifyDay bool) {
	next := occurrenceIn(date, today.Year())
	isEventDay = next.Equal(today)
	if next.Before(today) {
		next = occurrenceIn(date, today.Year()+1)
	}
	isNotifyDay = next.AddDate(0, 0, -notifyBeforeDays).Equal(today)
	return isEventDay, isNotifyDay
}

func (uc *UseCase) GetAllImportantDatesMessages(ctx context.Context, now time.Time) ([]domain.ImportantDateMessage, error) {
	today := dayOf(now)

	importantDates, err := uc.importantDates.GetAllActiveImportantDates(ctx)
	if err != nil {
		return nil, err
	}

	var messages []domain.ImportantDateMessage
	for _, importantDate := range importantDates {
		if !importantDate.IsActive {
			continue
		}

		// Дата создана только сегодня: уведомлять о ней не нужно
		if dayOf(importantDate.CreatedAt.In(now.Location())).Equal(today) {
			continue
		}

		isEventDay, isNotifyDay := isEventAndNotifyDay(importantDate.Date, importantDate.NotifyBeforeDays, today)
		if !isNotifyDay && !isEventDay {
			continue
		}

		if importantDate.LastNotificationAt.Valid && dayOf(importantDate.LastNotificationAt.Time.In(now.Location())).Equal(today) {
			continue
		}

		title := importantDate.Title
		htmlTitle := html.EscapeString(title)
		dateText := importantDate.Date.Format("02.01.2006")

		var text, plain string
		if isEventDay {
			text = fmt.Sprintf("🎉 Ура! Сегодня важная дата!\n\n<b>%s</b>\n%s", htmlTitle, dateText)
			plain = fmt.Sprintf("🎉 Ура! Сегодня важная дата!\n\n%s\n%s", title, dateText)
		} else {
			text = fmt.Sprintf(
				"⏰ Напоминание: через %d дн.\n\n<b>%s</b>\n%s",
				importantDate.NotifyBeforeDays,
				htmlTitle,
				dateText,
			)
			plain = fmt.Sprintf(
				"⏰ Напоминание: через %d дн.\n\n%s\n%s",
				importantDate.NotifyBeforeDays,
				title,
				dateText,
			)
		}

		hasUser := importantDate.UserID.Valid && importantDate.UserID.Int64 != 0
		hasPartner := importantDate.PartnerID.Valid && importantDate.PartnerID.Int64 != 0

		if hasUser && hasPartner && isEventDay {
			hint := "\n\n💡 За идеей для праздника загляни в меню «Идеи»."
			text += hint
			plain += hint
		}

		var tgIDs []int64
		message := domain.ImportantDateMessage{
			ImportantDateID: importantDate.ID,
			Message:         text,
			PlainMessage:    plain,
		}
		if hasUser {
			tgIDs = append(tgIDs, importantDate.UserID.Int64)
		}
		if hasPartner {
			tgIDs = append(tgIDs, importantDate.PartnerID.Int64)
		}
		message.UserIDs = tgIDs
		messages = append(messages, message)
	}
	return messages, nil
}

func (uc *UseCase) QueueImportantDateDeliveries(ctx context.Context, now time.Time, pushEnabled bool) error {
	messages, err := uc.GetAllImportantDatesMessages(ctx, now)
	if err != nil {
		return err
	}
	for _, message := range messages {
		for _, userID := range message.UserIDs {
			if err := uc.scheduler.QueueImportantDateDelivery(ctx, message.ImportantDateID, userID, now, "telegram", message.Message); err != nil {
				return err
			}
			if pushEnabled {
				if err := uc.scheduler.QueueImportantDateDelivery(ctx, message.ImportantDateID, userID, now, "push", message.PlainMessage); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (uc *UseCase) DeliverImportantDates(ctx context.Context, telegram, push domain.Sender) error {
	pending, err := uc.scheduler.PendingImportantDateDeliveries(ctx, push != nil)
	if err != nil {
		return err
	}
	for _, delivery := range pending {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		channel := telegram
		if delivery.Channel == "push" {
			channel = push
		}
		err = uc.scheduler.DeliverImportantDate(ctx, delivery, func(ctx context.Context) error {
			ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
			defer cancel()
			return channel.SendImportantDatesNotification(ctx, domain.ImportantDateMessage{
				ImportantDateID: delivery.DateID,
				UserIDs:         []int64{delivery.UserID},
				Message:         delivery.Message,
				PlainMessage:    delivery.Message,
			})
		})
		if err != nil {
			slog.Error("Important date delivery failed", "delivery_id", delivery.ID, "channel", delivery.Channel, "error", err)
		}
	}
	return nil
}
