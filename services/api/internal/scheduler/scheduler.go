package scheduler

import (
	"context"
	"log/slog"
	"time"

	"github.com/Waycoolers/fmlbot/services/api/internal/domain"
	"github.com/Waycoolers/fmlbot/services/api/internal/handlers"
	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	h *handlers.Handler
	c *cron.Cron
	s domain.Sender
}

func New(h *handlers.Handler, s domain.Sender) *Scheduler {
	loc, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		slog.Error("Failed to load timezone, using UTC", "error", err)
		loc = time.UTC
	}

	c := cron.New(cron.WithSeconds(), cron.WithLocation(loc))
	return &Scheduler{h: h, c: c, s: s}
}

func (s *Scheduler) Run(ctx context.Context) {
	_, err := s.c.AddFunc("0 0 0 * * *", func() {
		slog.Info("Complete your daily task at 00:00")
		s.h.DoMidnightTasks(ctx)
	})
	if err != nil {
		slog.Error("Error adding cron task", "error", err)
		return
	}

	_, err = s.c.AddFunc("0 0 12 * * *", func() {
		slog.Info("Checking important dates (12:00 MSK)")
		s.h.NotifyImportantDatesCron(ctx, s.s)
	})
	if err != nil {
		slog.Error("Error adding cron task for important date notifications", "error", err)
		return
	}

	s.c.Start()
}

func (s *Scheduler) Stop() {
	if s.c != nil {
		slog.Info("The context is complete, stop cron")
		s.c.Stop()
	}
}
