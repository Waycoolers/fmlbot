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

	c := cron.New(cron.WithSeconds(), cron.WithLocation(loc), cron.WithChain(cron.SkipIfStillRunning(cron.DefaultLogger)))
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

	checkDates := func() {
		s.h.NotifyImportantDatesCron(ctx, time.Now().In(s.loc), s.httpSender, s.fcmSender)
	}
	// Новые уведомления появляются с 12:00. Очередь повторяется и после полуночи.
	dateJob := cron.SkipIfStillRunning(cron.DefaultLogger)(cron.FuncJob(checkDates))
	_, err = s.c.AddJob("0 0 * * * *", dateJob)
	if err != nil {
		slog.Error("Error adding cron task for important date notifications", "error", err)
		return
	}

	s.h.RunMissedMidnightTasks(ctx, s.loc)

	s.c.Start()
	go dateJob.Run()
}

func (s *Scheduler) Stop() {
	if s.c != nil {
		slog.Info("The context is complete, stop cron")
		s.c.Stop()
	}
}
