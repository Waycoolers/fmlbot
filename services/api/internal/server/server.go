package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/Waycoolers/fmlbot/services/api/internal/config"
	"github.com/Waycoolers/fmlbot/services/api/internal/handlers"
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
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {})

	addr := fmt.Sprintf("http://%s:%s", s.config.Host, s.config.Port)
	return &http.Server{
		Addr:    addr,
		Handler: mux,
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
