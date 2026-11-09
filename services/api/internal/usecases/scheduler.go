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

func (uc *UseCase) GetAllImportantDatesMessages(ctx context.Context) ([]domain.ImportantDateMessage, error) {
	today := dayOf(time.Now())

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
		if dayOf(importantDate.CreatedAt).Equal(today) {
			continue
		}

		isEventDay, isNotifyDay := isEventAndNotifyDay(importantDate.Date, importantDate.NotifyBeforeDays, today)
		if !isNotifyDay && !isEventDay {
			continue
		}

		if importantDate.LastNotificationAt.Valid && dayOf(importantDate.LastNotificationAt.Time).Equal(today) {
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
			ideas, aiErr := uc.ai.GetDateIdea(ctx, importantDate.Title, 0, "Простое оригинальное милое свидание чтобы отметить дату")
			if aiErr == nil && ideas != "" {
				text = fmt.Sprintf("%s\n\n💡 <b>Случайная идея по этому поводу:</b>\n%s", text, html.EscapeString(ideas))
				plain = fmt.Sprintf("%s\n\n💡 Случайная идея по этому поводу:\n%s", plain, ideas)
			} else {
				slog.Error("Failed to get AI date idea", "error", aiErr)
			}
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

func (uc *UseCase) UpdateLastNotificationAt(ctx context.Context, message domain.ImportantDateMessage) error {
	now := time.Now().UTC()
	err := uc.importantDates.UpdateLastNotificationAt(ctx, message.ImportantDateID, now)
	if err != nil {
		return err
	}
	return nil
}
