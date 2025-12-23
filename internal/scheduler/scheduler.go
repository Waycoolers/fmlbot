package scheduler

import (
	"context"
	"log"
	"time"

	"github.com/Waycoolers/fmlbot/internal/handlers"
	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	h *handlers.Handler
	c *cron.Cron
}

func New(h *handlers.Handler) *Scheduler {
	loc, _ := time.LoadLocation("Europe/Moscow")
	c := cron.New(cron.WithSeconds(), cron.WithLocation(loc))
	return &Scheduler{h: h, c: c}
}

func (s *Scheduler) Run(ctx context.Context) {
	_, err := s.c.AddFunc("0 0 0 * * *", func() {
		log.Println("Выполняем ежедневную задачу в 00:00")
		s.h.DoMidnightTasks(ctx)
	})
	if err != nil {
		log.Printf("Ошибка добавления cron-задачи: %v", err)
		return
	}

	_, err = s.c.AddFunc("0 0 12 * * *", func() {
		log.Println("🔔 Проверяем важные даты (12:00 МСК)")
		s.h.NotifyImportantDatesCron(ctx)
	})
	if err != nil {
		log.Printf("Ошибка добавления cron-задачи уведомлений о важных датах: %v", err)
		return
	}

	s.c.Start()

	go func() {
		<-ctx.Done()
		log.Println("Контекст завершён, останавливаем cron")
		s.c.Stop()
	}()
}
