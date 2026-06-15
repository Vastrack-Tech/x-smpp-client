package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

func (h *Handler) HandleHealth(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":     "ok",
		"uptime":     time.Since(h.StartTime).Round(time.Second).String(),
		"queue_size": h.Queue.Len(),
	})
}
