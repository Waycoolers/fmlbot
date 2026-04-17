package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/Waycoolers/fmlbot/services/bot/internal/config"
	"github.com/Waycoolers/fmlbot/services/bot/internal/handlers"
)

type HTTPServer struct {
	s   *http.Server
	cfg *config.ServerConfig
	h   *handlers.Handler
}

func NewHTTPServer(cfg *config.ServerConfig, h *handlers.Handler) *HTTPServer {
	addr := fmt.Sprintf(":%d", cfg.Port)

	mux := http.NewServeMux()
	mux.HandleFunc("/updates/important_dates", h.NotifyAllImportantDates)

	return &HTTPServer{
		s:   &http.Server{Addr: addr, Handler: mux},
		cfg: cfg,
		h:   h,
	}
}

func (s *HTTPServer) Run() {
	go func() {
		slog.Info("Starting HTTP server")
		err := s.s.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Server stopped with error", "error", err)
		}
	}()
}

func (s *HTTPServer) Stop(ctx context.Context) error {
	if s.s != nil {
		slog.Info("Stopping HTTP server...")
		return s.s.Shutdown(ctx)
	}
	return nil
}
