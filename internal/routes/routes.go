package routes

import (
	"github.com/gofiber/fiber/v2"
	"x-smpp-client/internal/api/handlers"
)

func RegisterRoutes(app *fiber.App, h *handlers.Handler) {
	v1 := app.Group("/v1")

	v1.Post("/send", h.HandleSend)
	v1.Get("/health", h.HandleHealth)
}
