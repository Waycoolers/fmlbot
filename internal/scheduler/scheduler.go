package scheduler

import (
	"context"
	"log"

	"github.com/Waycoolers/fmlbot/internal/handlers"
	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	h *handlers.Handler
	c *cron.Cron
}

func New(h *handlers.Handler) *Scheduler {
	c := cron.New(cron.WithSeconds())
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

	s.c.Start()

	go func() {
		<-ctx.Done()
		log.Println("Контекст завершён, останавливаем cron")
		s.c.Stop()
	}()
}
