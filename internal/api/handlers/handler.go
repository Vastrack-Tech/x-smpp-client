package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"x-smpp-client/internal/queue"
)

type Handler struct {
	Queue     *queue.Queue
	StartTime time.Time
}

func New(q *queue.Queue) *Handler {
	return &Handler{
		Queue:     q,
		StartTime: time.Now(),
	}
}

func newID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
