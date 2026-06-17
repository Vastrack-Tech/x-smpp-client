package handlers

import (
	"github.com/gofiber/fiber/v2"
	"x-smpp-client/internal/models"
	"x-smpp-client/internal/utils"
)

func (h *Handler) HandleSend(c *fiber.Ctx) error {
	var req utils.SendRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid JSON body",
		})
	}

	if err := utils.ValidateSendRequest(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	account := c.Locals("account").(*models.Account)

	if err := h.Accounts.CheckBalance(c.Context(), account.ID, 1); err != nil {
		return c.Status(fiber.StatusPaymentRequired).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	msg := models.Message{
		AccountID:  account.ID,
		To:         req.To,
		Text:       req.Text,
		SourceAddr: req.SourceAddr,
		Encoding:   req.Encoding,
	}

	if err := h.Accounts.CreateMessage(c.Context(), &msg); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create message",
		})
	}

	h.Queue.Push(msg)

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"message_id": msg.ID,
		"status":     "queued",
	})
}
