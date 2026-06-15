package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"x-smpp-client/internal/accounts/dto"
	"x-smpp-client/internal/accounts/service"
)

type Handler struct {
	svc *service.Service
}

func New(svc *service.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) HandleCreate(c *fiber.Ctx) error {
	var req dto.CreateAccountRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid JSON"})
	}
	if req.Name == "" || req.Email == "" {
		return c.Status(400).JSON(fiber.Map{"error": "name and email are required"})
	}

	a, err := h.svc.CreateAccount(c.Context(), req.Name, req.Email)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(201).JSON(a)
}

func (h *Handler) HandleGetBalance(c *fiber.Ctx) error {
	id := c.Params("id")
	bal, err := h.svc.GetBalance(c.Context(), id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(dto.BalanceResponse{
		AccountID: id,
		Balance:   bal,
	})
}

func (h *Handler) HandleTopUp(c *fiber.Ctx) error {
	id := c.Params("id")
	var req dto.TopUpRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid JSON"})
	}
	if req.Amount <= 0 {
		return c.Status(400).JSON(fiber.Map{"error": "amount must be positive"})
	}

	e, err := h.svc.TopUpBalance(c.Context(), id, req.Amount, req.Description)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(e)
}

func (h *Handler) HandleGetTransactions(c *fiber.Ctx) error {
	id := c.Params("id")
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	entries, err := h.svc.GetLedgerEntries(c.Context(), id, limit, offset)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(entries)
}

func (h *Handler) HandleCreateAPIKey(c *fiber.Ctx) error {
	id := c.Params("id")
	var req dto.CreateAPIKeyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid JSON"})
	}

	k, err := h.svc.CreateAPIKey(c.Context(), id, req.Name)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(201).JSON(k)
}
