package app

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/Waycoolers/fmlbot/services/bot/internal/domain"
)

type clientStub struct {
	domain.BotClient
	updates chan domain.Update
}

func (c *clientStub) GetUpdatesChan() <-chan domain.Update { return c.updates }

func TestRunStopsOnContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	b := &Bot{client: &clientStub{updates: make(chan domain.Update)}}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		b.Run(ctx)
	}()
	cancel()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("Run did not stop after context cancel")
	}
}
