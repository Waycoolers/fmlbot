package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/Waycoolers/fmlbot/services/auth/internal/config"
	"github.com/Waycoolers/fmlbot/services/auth/internal/handlers"
	"github.com/Waycoolers/fmlbot/services/auth/internal/middleware"
)

type Server struct {
	config *config.ServerConfig
	server *http.Server
	h      *handlers.Handler
}

func New(cfg *config.ServerConfig, h *handlers.Handler) *Server {
	return &Server{
		config: cfg,
		h:      h,
	}
}

func (s *Server) newServer() *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /auth/token", s.h.Token)
	mux.HandleFunc("POST /auth/refresh", s.h.Refresh)
	mux.HandleFunc("POST /auth/revoke", s.h.Revoke)
	mux.HandleFunc("GET /health", s.h.Health)

	handler := middleware.Logging(mux)
	addr := fmt.Sprintf("%s:%s", s.config.Host, s.config.Port)
	return &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
}

func (s *Server) Start() {
	s.server = s.newServer()

	go func() {
		slog.Info("Starting server", "host", s.config.Host, "port", s.config.Port)
		err := s.server.ListenAndServe()
		if err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				slog.Error("Update server failed", "error", err)
			}
			slog.Error("Server error", "error", err)
		}
	}()
}

func (s *Server) Stop(ctx context.Context) error {
	slog.Info("Stopping server", "host", s.config.Host, "port", s.config.Port)
	if s.server != nil {
		err := s.server.Shutdown(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}
