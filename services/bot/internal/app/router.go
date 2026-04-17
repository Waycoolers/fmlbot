package app

import (
	"context"

	"github.com/Waycoolers/fmlbot/services/bot/internal/domain"
	"github.com/Waycoolers/fmlbot/services/bot/internal/handlers"
)

type Router struct {
	h *handlers.Handler
}

func NewRouter(h *handlers.Handler) *Router {
	return &Router{h: h}
}

func (r *Router) HandleUpdate(ctx context.Context, update domain.Update) {
	if update.CallbackQuery != nil {
		r.h.HandleCallback(ctx, update.CallbackQuery)
		return
	}

	if update.Message != nil {
		r.h.HandleMessage(ctx, update.Message)
		return
	}
}
