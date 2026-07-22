package app

import (
	"context"
	"log/slog"

	"github.com/Waycoolers/fmlbot/services/bot/internal/domain"
)

type Bot struct {
	client domain.BotClient
	router *Router
}

func New(c domain.BotClient, r *Router) (*Bot, error) {
	return &Bot{client: c, router: r}, nil
}

func (b *Bot) Run(ctx context.Context) {
	slog.Info("Bot is running")

	updates := b.client.GetUpdatesChan()

	for update := range updates {
		if update.Message != nil || update.CallbackQuery != nil {
			b.router.HandleUpdate(ctx, update)
		}
	}
}

func (b *Bot) Stop() {
	if b.client != nil {
		slog.Info("Bot is stopping")
		b.client.StopReceivingUpdates()
	}
}
