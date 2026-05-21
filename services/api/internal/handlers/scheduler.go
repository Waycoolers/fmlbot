package handlers

import (
	"context"
	"log/slog"

	"github.com/Waycoolers/fmlbot/services/api/internal/domain"
)

func (h *Handler) DoMidnightTasks(ctx context.Context) {
	err := h.uc.DoMidnightTasks(ctx)
	if err != nil {
		slog.Error("Failed to do midnight tasks", "error", err)
	}

	slog.Info("Tasks completed")
}

func (h *Handler) NotifyImportantDatesCron(ctx context.Context, httpSender, fcmSender domain.Sender) {
	messages, err := h.uc.GetAllImportantDatesMessages(ctx)
	if err != nil {
		slog.Error("Error getting important dates messages", "error", err)
		return
	}

	for _, msg := range messages {
		ok := true
		err = httpSender.SendImportantDatesNotification(ctx, msg)
		if err != nil {
			slog.Error("Error sending important dates message to bot client", "error", err)
			ok = false
		}
		err = fcmSender.SendImportantDatesNotification(ctx, msg)
		if err != nil {
			slog.Error("Error sending important dates message to fcm", "error", err)
			ok = false
		}
		if !ok {
			continue
		}
		err = h.uc.UpdateLastNotificationAt(ctx, msg)
		if err != nil {
			slog.Error("Error updating last notification at", "error", err)
		}
	}

	slog.Info("Notifications about important dates have been processed")
}
