package api

import (
	"context"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	accountsservice "x-smpp-client/internal/accounts/service"
	"x-smpp-client/internal/api/handlers"
	authservice "x-smpp-client/internal/auth/service"
	"x-smpp-client/internal/config"
	"x-smpp-client/internal/queue"
	"x-smpp-client/internal/routes"
)

type Server struct {
	app *fiber.App
	cfg config.ServerConfig
}

func New(q *queue.Queue, svc *accountsservice.Service, auth *authservice.AuthService, cfg config.ServerConfig) *Server {
	app := fiber.New(fiber.Config{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	})

	h := handlers.New(q, svc)
	routes.RegisterRoutes(app, h, auth)

	return &Server{app: app, cfg: cfg}
}

func (s *Server) Serve(ctx context.Context) error {
	go func() {
		<-ctx.Done()
		_ = s.app.ShutdownWithTimeout(5 * time.Second)
	}()

	log.Printf("API server listening on %s", s.cfg.ListenAddr)
	if err := s.app.Listen(s.cfg.ListenAddr); err != nil {
		return err
	}
	return nil
}
