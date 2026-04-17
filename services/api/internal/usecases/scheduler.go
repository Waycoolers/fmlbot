package usecases

import (
	"context"
	"fmt"
	"time"

	"github.com/Waycoolers/fmlbot/services/api/internal/domain"
)

func (uc *UseCase) DoMidnightTasks(ctx context.Context) error {
	return uc.scheduler.DoMidnightTasksWithCompliments(ctx)
}

func (uc *UseCase) GetAllImportantDatesMessages(ctx context.Context) ([]domain.ImportantDateMessage, error) {
	now := time.Now()
	today := time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		0, 0, 0, 0,
		time.Local,
	)

	importantDates, err := uc.importantDates.GetAllActiveImportantDates(ctx)
	if err != nil {
		return nil, err
	}

	var messages []domain.ImportantDateMessage
	for _, importantDate := range importantDates {
		if !importantDate.IsActive {
			continue
		}

		eventDate := importantDate.Date.In(time.Local)
		eventDay := time.Date(
			eventDate.Year(),
			eventDate.Month(),
			eventDate.Day(),
			0, 0, 0, 0,
			time.Local,
		)

		notifyDay := eventDay.AddDate(0, 0, -importantDate.NotifyBeforeDays)

		isNotifyDay := notifyDay.Equal(today)
		isEventDay := eventDay.Equal(today)

		if !isNotifyDay && !isEventDay {
			continue
		}

		if importantDate.LastNotificationAt.Valid {
			last := importantDate.LastNotificationAt.Time.In(time.Local)
			lastDay := time.Date(
				last.Year(),
				last.Month(),
				last.Day(),
				0, 0, 0, 0,
				time.Local,
			)
			if lastDay.Equal(today) {
				continue
			}
		}

		var text string
		if isEventDay {
			text = fmt.Sprintf("🎉 Ура! Сегодня важная дата!\n\n<b>%s</b>\n%s",
				importantDate.Title,
				eventDate.Format("02.01.2006"),
			)
		} else {
			text = fmt.Sprintf(
				"⏰ Напоминание: через %d дн.\n\n<b>%s</b>\n%s",
				importantDate.NotifyBeforeDays,
				importantDate.Title,
				eventDate.Format("02.01.2006"),
			)
		}

		var tgIDs []int64
		message := domain.ImportantDateMessage{
			ImportantDateID: importantDate.ID,
			Message:         text,
		}
		if importantDate.UserID.Valid && importantDate.UserID.Int64 != 0 {
			tgIDs = append(tgIDs, importantDate.UserID.Int64)
		}
		if importantDate.PartnerID.Valid && importantDate.PartnerID.Int64 != 0 {
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
