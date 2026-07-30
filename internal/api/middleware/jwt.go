package middleware

import (
        "errors"

	"github.com/gofiber/fiber/v2"
	"x-smpp-client/internal/auth/service"
)

type ctxKey string

const accountIDKey ctxKey = "account_id"

func AccountIDFrom(c *fiber.Ctx) (string, bool) {
    id, ok := c.Locals("account_id").(string)
    return id, ok
}

func JWTAuth(auth *service.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		tokenStr := extractBearerToken(c)
		if tokenStr == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing Authorization header"})
		}

		accountID, err := auth.ValidateToken(c.Context(), tokenStr)
		if err != nil {
			status := fiber.StatusUnauthorized
			if errors.Is(err, service.ErrTokenExpired) {
				status = fiber.StatusUnauthorized
			}
			return c.Status(status).JSON(fiber.Map{"error": err.Error()})
		}

		c.Locals("account_id", accountID)

        return c.Next()
	}
}

func extractBearerToken(c *fiber.Ctx) string {
	s := c.Get("Authorization")
	if len(s) > 7 && s[:7] == "Bearer " {
		return s[7:]
	}
	return ""
}
