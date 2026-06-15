package routes

import (
	"github.com/gofiber/fiber/v2"
	"x-smpp-client/internal/api/handlers"
)

func RegisterRoutes(app *fiber.App, h *handlers.Handler) {
	v1 := app.Group("/v1")

	v1.Post("/send", h.HandleSend)
	v1.Get("/health", h.HandleHealth)

	accounts := v1.Group("/accounts")
	accounts.Post("/", h.HandleCreate)
	accounts.Get("/:id/balance", h.HandleGetBalance)
	accounts.Post("/:id/topup", h.HandleTopUp)
	accounts.Get("/:id/transactions", h.HandleGetTransactions)
	accounts.Post("/:id/apikeys", h.HandleCreateAPIKey)
}
