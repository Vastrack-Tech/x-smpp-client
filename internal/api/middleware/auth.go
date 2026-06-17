package middleware

import (
	"github.com/gofiber/fiber/v2"
	"x-smpp-client/internal/accounts/service"
)

func APIKeyAuth(svc *service.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		key := c.Get("X-API-Key")
		if key == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "missing X-API-Key header",
			})
		}

		account, err := svc.ValidateAPIKey(c.Context(), key)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid API key",
			})
		}

		c.Locals("account", account)
		return c.Next()
	}
}
