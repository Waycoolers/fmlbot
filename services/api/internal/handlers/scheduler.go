package handlers

import (
	"context"
	"log/slog"
	"time"

	"github.com/Waycoolers/fmlbot/services/api/internal/domain"
)

func (h *Handler) DoMidnightTasks(ctx context.Context) {
	err := h.uc.DoMidnightTasks(ctx)
	if err != nil {
		slog.Error("Failed to do midnight tasks", "error", err)
	}

	slog.Info("Tasks completed")
}

func (h *Handler) RunMissedMidnightTasks(ctx context.Context, loc *time.Location) {
	if err := h.uc.RunMissedMidnightTasks(ctx, loc); err != nil {
		slog.Error("Failed to run missed midnight tasks", "error", err)
	}
}

func (h *Handler) NotifyImportantDatesCron(ctx context.Context, now time.Time, httpSender, fcmSender domain.Sender) {
	if now.Hour() >= 12 {
		if err := h.uc.QueueImportantDateDeliveries(ctx, now, fcmSender != nil); err != nil {
			slog.Error("Error queueing important date notifications", "error", err)
		}
	}
	if err := h.uc.DeliverImportantDates(ctx, httpSender, fcmSender); err != nil {
		slog.Error("Error processing important date notifications", "error", err)
	}
}

func (h *Handler) NotifyPlansCron(ctx context.Context, sender domain.PlanNoticeSender) {
	if err := h.uc.DeliverPlanNotices(ctx, sender); err != nil {
		slog.Error("Error processing plan notifications", "error", err)
	}
}
