package middleware

import (
	"context"
	"errors"

	"github.com/gofiber/fiber/v2"
	"x-smpp-client/internal/auth/service"
)

type ctxKey string

const accountIDKey ctxKey = "account_id"

func AccountIDFrom(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(accountIDKey).(string)
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

		ctx := context.WithValue(c.Context(), accountIDKey, accountID)
		c.SetUserContext(ctx)
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
