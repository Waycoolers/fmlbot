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
	h          *handlers.Handler
	c          *cron.Cron
	loc        *time.Location
	httpSender domain.Sender
	fcmSender  domain.Sender
}

func New(h *handlers.Handler, httpSender, fcmSender domain.Sender) *Scheduler {
	loc, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		slog.Error("Failed to load timezone, using UTC", "error", err)
		loc = time.UTC
	}

	c := cron.New(cron.WithSeconds(), cron.WithLocation(loc))
	return &Scheduler{h: h, c: c, loc: loc, httpSender: httpSender, fcmSender: fcmSender}
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

	// Основной запуск в 12:00, остальные повторяют только то, что не удалось доставить
	_, err = s.c.AddFunc("0 0 12-23 * * *", func() {
		slog.Info("Checking important dates")
		s.h.NotifyImportantDatesCron(ctx, s.httpSender, s.fcmSender)
	})
	if err != nil {
		slog.Error("Error adding cron task for important date notifications", "error", err)
		return
	}

	s.h.RunMissedMidnightTasks(ctx, s.loc)

	s.c.Start()
}

func (s *Scheduler) Stop() {
	if s.c != nil {
		slog.Info("The context is complete, stop cron")
		s.c.Stop()
	}
}
