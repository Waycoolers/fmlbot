package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/Waycoolers/fmlbot/pkg/middlewares"
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
	mux.HandleFunc("POST /users", s.h.AddUser)
	mux.HandleFunc("GET /users/me", s.h.GetMe)
	mux.HandleFunc("PATCH /users/me/password", s.h.ChangePassword)
	mux.HandleFunc("GET /users/partner", s.h.GetPartner)
	mux.HandleFunc("DELETE /users/me", s.h.DeleteUser)
	mux.HandleFunc("PUT /users/me", s.h.UpdateUser)
	mux.HandleFunc("PUT /users/partner", s.h.UpdatePartner)
	mux.HandleFunc("POST /users/pair", s.h.AddPartners)
	mux.HandleFunc("PATCH /users/unpair", s.h.DeletePartners)
	mux.HandleFunc("GET /users/by-username/{username}", s.h.GetUserByUsername)
	mux.HandleFunc("POST /users/fcm-token", s.h.SaveFCMToken)

	mux.HandleFunc("GET /user_config/me", s.h.GetMyUserConfig)
	mux.HandleFunc("GET /user_config/partner", s.h.GetPartnerUserConfig)
	mux.HandleFunc("PATCH /user_config/me", s.h.UpdateUserConfig)
	mux.HandleFunc("POST /user_config/reset/me", s.h.ResetMyUserConfig)
	mux.HandleFunc("POST /user_config/reset/partner", s.h.ResetPartnerUserConfig)

	mux.HandleFunc("POST /compliments", s.h.AddCompliment)
	mux.HandleFunc("GET /compliments", s.h.GetAllCompliments)
	mux.HandleFunc("GET /compliments/received", s.h.GetAllReceivedCompliments)
	mux.HandleFunc("PUT /compliments/{id}", s.h.UpdateCompliment)
	mux.HandleFunc("DELETE /compliments/{id}", s.h.RemoveCompliment)
	mux.HandleFunc("POST /compliments/next", s.h.ReceiveCompliment)

	mux.HandleFunc("POST /important_dates", s.h.AddImportantDate)
	mux.HandleFunc("GET /important_dates/{id}", s.h.GetImportantDate)
	mux.HandleFunc("GET /important_dates", s.h.GetAllImportantDates)
	mux.HandleFunc("PATCH /important_dates/{id}", s.h.UpdateImportantDate)
	mux.HandleFunc("PATCH /important_dates/{id}/sharing", s.h.UpdateImportantDateSharing)
	mux.HandleFunc("DELETE /important_dates/{id}", s.h.RemoveImportantDate)

	mux.HandleFunc("POST /auth/verify", s.h.VerifyUser) // Приватный
	mux.HandleFunc("POST /auth/register", s.h.Register) // Приватный

	handler := middlewares.Middleware(s.config.JwtSecret, "GET /users/by-username", "POST /auth/verify", "POST /auth/register")(mux)
	handler = middlewares.InternalAuthMiddlewareWithPrivatePaths(string(s.config.InternalSecret), "POST /auth/verify", "POST /auth/register")(handler)
	addr := fmt.Sprintf("%s:%s", s.config.Host, s.config.Port)
	return &http.Server{
		Addr:    addr,
		Handler: handler,
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
	if s.server != nil {
		slog.Info("Stopping server", "host", s.config.Host, "port", s.config.Port)
		err := s.server.Shutdown(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}
