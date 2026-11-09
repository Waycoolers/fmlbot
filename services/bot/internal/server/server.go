package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/Waycoolers/fmlbot/pkg/middlewares"
	"github.com/Waycoolers/fmlbot/services/bot/internal/config"
	"github.com/Waycoolers/fmlbot/services/bot/internal/handlers"
)

type HTTPServer struct {
	s   *http.Server
	cfg *config.ServerConfig
	h   *handlers.Handler
}

func NewHTTPServer(cfg *config.ServerConfig, h *handlers.Handler, internalSecret []byte) *HTTPServer {
	addr := fmt.Sprintf(":%d", cfg.Port)

	mux := http.NewServeMux()
	mux.HandleFunc("/updates/message", h.SendMessage)
	mux.HandleFunc("/updates/important_dates", h.NotifyAllImportantDates)
	mux.HandleFunc("POST /updates/plans", h.NotifyPlan)

	return &HTTPServer{
		s: &http.Server{
			Addr:              addr,
			Handler:           middlewares.LimitBody(1 << 20)(middlewares.InternalAuthMiddlewareWithPublicPaths(string(internalSecret))(mux)),
			ReadHeaderTimeout: 5 * time.Second,
			ReadTimeout:       15 * time.Second,
			WriteTimeout:      2 * time.Minute,
			IdleTimeout:       2 * time.Minute,
		},
		cfg: cfg,
		h:   h,
	}
}

func (s *HTTPServer) Run() {
	go func() {
		slog.Info("Starting HTTP server", "port", s.cfg.Port)
		err := s.s.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Server stopped with error", "error", err)
			os.Exit(1)
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
