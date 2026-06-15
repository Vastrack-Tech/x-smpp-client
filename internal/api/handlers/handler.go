package handlers

import (
	"time"

	"x-smpp-client/internal/accounts/handler"
	"x-smpp-client/internal/accounts/service"
	"x-smpp-client/internal/queue"
)

type Handler struct {
	*handler.Handler
	Accounts  *service.Service
	Queue     *queue.Queue
	StartTime time.Time
}

func New(q *queue.Queue, svc *service.Service) *Handler {
	return &Handler{
		Handler:   handler.New(svc),
		Accounts:  svc,
		Queue:     q,
		StartTime: time.Now(),
	}
}
