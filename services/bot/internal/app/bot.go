package app

import (
	"context"
	"log/slog"
	"runtime/debug"
	"sync"

	"github.com/Waycoolers/fmlbot/services/bot/internal/domain"
)

const (
	workersCount = 16
	queueSize    = 64
)

type Bot struct {
	client domain.BotClient
	router *Router
}

func New(c domain.BotClient, r *Router) (*Bot, error) {
	return &Bot{client: c, router: r}, nil
}

func chatIDOf(update domain.Update) int64 {
	if update.Message != nil {
		return update.Message.ChatID
	}
	return update.CallbackQuery.ChatID
}

// Run раздаёт обновления воркерам. Обновления одного чата всегда попадают к одному воркеру
// и идут по порядку, а долгий запрос одного пользователя не задерживает остальных.
func (b *Bot) Run(ctx context.Context) {
	slog.Info("Bot is running")

	updates := b.client.GetUpdatesChan()

	var wg sync.WaitGroup
	queues := make([]chan domain.Update, workersCount)
	for i := range queues {
		queues[i] = make(chan domain.Update, queueSize)
		wg.Add(1)
		go func(queue <-chan domain.Update) {
			defer wg.Done()
			for update := range queue {
				b.handle(ctx, update)
			}
		}(queues[i])
	}
	defer func() {
		for _, queue := range queues {
			close(queue)
		}
		wg.Wait()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case update, ok := <-updates:
			if !ok {
				return
			}
			if update.Message == nil && update.CallbackQuery == nil {
				continue
			}
			queue := queues[uint64(chatIDOf(update))%workersCount]
			select {
			case queue <- update:
			case <-ctx.Done():
				return
			}
		}
	}
}

func (b *Bot) handle(ctx context.Context, update domain.Update) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("Panic while handling update", "panic", r, "stack", string(debug.Stack()))
		}
	}()
	b.router.HandleUpdate(ctx, update)
}

func (b *Bot) Stop() {
	if b.client != nil {
		slog.Info("Bot is stopping")
		b.client.StopReceivingUpdates()
	}
}
