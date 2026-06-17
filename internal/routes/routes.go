package routes

import (
	"github.com/gofiber/fiber/v2"
	"x-smpp-client/internal/api/handlers"
	"x-smpp-client/internal/api/middleware"
	"x-smpp-client/internal/auth/handler"
	"x-smpp-client/internal/auth/service"
)

func RegisterRoutes(app *fiber.App, h *handlers.Handler, authSvc *service.AuthService) {
	app.Get("/health", h.HandleHealth)

	authH := handler.New(authSvc)

	api := app.Group("/api")

	api.Post("/accounts", h.HandleCreate)

	authed := api.Group("", middleware.APIKeyAuth(h.Accounts))
	authed.Post("/send", h.HandleSend)
	authed.Get("/balance", h.HandleGetBalance)

	app.Post("/login", authH.HandleLogin)

	dashboard := app.Group("/dashboard", middleware.JWTAuth(authSvc))
	dashboard.Post("/logout", authH.HandleLogout)
	dashboard.Get("/me", authH.HandleMe)
	dashboard.Post("/accounts/topup", h.HandleTopUp)
	dashboard.Get("/accounts/transactions", h.HandleGetTransactions)
	dashboard.Post("/accounts/apikeys", h.HandleCreateAPIKey)
}
