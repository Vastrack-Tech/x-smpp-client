package handler

import (
	"github.com/gofiber/fiber/v2"
	"x-smpp-client/internal/api/middleware"
	"x-smpp-client/internal/auth/dto"
	"x-smpp-client/internal/auth/service"
)

type AuthHandler struct {
	svc *service.AuthService
}

func New(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) HandleLogin(c *fiber.Ctx) error {
	var req dto.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid body"})
	}
	if req.Email == "" || req.Password == "" {
		return c.Status(400).JSON(fiber.Map{"error": "email and password required"})
	}

	token, acc, err := h.svc.Login(c.Context(), req.Email, req.Password)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(dto.LoginResponse{
		Token:     token,
		AccountID: acc.ID,
		Name:      acc.Name,
		Email:     acc.Email,
	})
}

func (h *AuthHandler) HandleLogout(c *fiber.Ctx) error {
	accountID, _ := middleware.AccountIDFrom(c.Context())
	tokenStr := extractBearer(c)
	if tokenStr == "" {
		return c.SendStatus(204)
	}

	if err := h.svc.Logout(c.Context(), accountID, tokenStr); err != nil {
		return c.SendStatus(204)
	}
	return c.SendStatus(204)
}

func (h *AuthHandler) HandleMe(c *fiber.Ctx) error {
	accountID, _ := middleware.AccountIDFrom(c.Context())

	acc, err := h.svc.GetAccount(c.Context(), accountID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "account not found"})
	}

	return c.JSON(dto.MeResponse{
		AccountID: acc.ID,
		Name:      acc.Name,
		Email:     acc.Email,
	})
}

func extractBearer(c *fiber.Ctx) string {
	s := c.Get("Authorization")
	if len(s) > 7 && s[:7] == "Bearer " {
		return s[7:]
	}
	return ""
}
