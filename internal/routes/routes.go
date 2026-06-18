package routes

import (
	"github.com/gofiber/fiber/v2"
	accountsservice "x-smpp-client/internal/accounts/service"
	"x-smpp-client/internal/api/handlers"
	"x-smpp-client/internal/api/middleware"
	"x-smpp-client/internal/auth/handler"
	"x-smpp-client/internal/auth/service"
)

func RegisterRoutes(app *fiber.App, h *handlers.Handler, authSvc *service.AuthService) {
	app.Get("/health", h.HandleHealth)

	authH := handler.New(authSvc, h.Accounts)

	api := app.Group("/api")

	api.Post("/accounts", h.HandleCreate)

	authed := api.Group("", middleware.APIKeyAuth(h.Accounts))
	authed.Post("/send", h.HandleSend)
	authed.Get("/balance", h.HandleGetBalance)

	app.Post("/auth/register", authH.HandleRegister)
	app.Post("/auth/login", authH.HandleLogin)

	dashboard := app.Group("/dashboard", middleware.JWTAuth(authSvc), resolveAccount(h.Accounts))
	dashboard.Post("/logout", authH.HandleLogout)
	dashboard.Get("/me", authH.HandleMe)
	dashboard.Post("/accounts/topup", h.HandleTopUp)
	dashboard.Get("/accounts/transactions", h.HandleGetTransactions)
	dashboard.Post("/accounts/apikeys", h.HandleCreateAPIKey)
	dashboard.Get("/accounts/apikeys", h.HandleListAPIKeys)
}

func resolveAccount(svc *accountsservice.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, ok := middleware.AccountIDFrom(c.Context())
		if !ok {
			return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
		}
		account, err := svc.GetAccount(c.Context(), id)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{"error": "account not found"})
		}
		c.Locals("account", account)
		return c.Next()
	}
}
