package handlers

import (
	"github.com/gofiber/fiber/v2"
	"x-smpp-client/internal/message"
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

	id := newID()

	h.Queue.Push(message.Message{
		ID:         id,
		To:         req.To,
		Text:       req.Text,
		SourceAddr: req.SourceAddr,
		Encoding:   req.Encoding,
	})

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"message_id": id,
		"status":     "queued",
	})
}
