package queue

import (
	"context"
	"log"

	"github.com/linxGnu/gosmpp/pdu"
	"x-smpp-client/internal/accounts/service"
	"x-smpp-client/internal/message"
	"x-smpp-client/internal/models"
	"x-smpp-client/internal/session"
)

type Queue struct {
	ch chan models.Message
}

func New(size int) *Queue {
	return &Queue{
		ch: make(chan models.Message, size),
	}
}

func (q *Queue) Push(msg models.Message) {
	q.ch <- msg
}

func (q *Queue) Consume() <-chan models.Message {
	return q.ch
}

func (q *Queue) Len() int {
	return len(q.ch)
}

type Worker struct {
	pool  *session.Pool
	q     *Queue
	svc   *service.Service
	def   Defaults
}

type Defaults struct {
	SourceAddr string
	SourceTon  byte
	SourceNpi  byte
	Encoding   string
}

func NewWorker(pool *session.Pool, q *Queue, svc *service.Service, def Defaults) *Worker {
	return &Worker{pool: pool, q: q, svc: svc, def: def}
}

func (w *Worker) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Println("worker shutting down")
			return
		case msg := <-w.q.Consume():
			w.process(ctx, msg)
		}
	}
}

func (w *Worker) process(ctx context.Context, msg models.Message) {
	enc := w.def.Encoding
	if msg.Encoding != "" {
		enc = msg.Encoding
	}
	encoding, err := message.ParseEncoding(enc)
	if err != nil {
		log.Printf("message %s: %v", msg.ID, err)
		_ = w.svc.UpdateMessageStatus(ctx, msg.ID, "failed", 0, 0)
		return
	}

	srcAddr := msg.SourceAddr
	srcTon := w.def.SourceTon
	srcNpi := w.def.SourceNpi
	if srcAddr == "" {
		srcAddr = w.def.SourceAddr
	}

	src := buildAddr(srcTon, srcNpi, srcAddr)
	dst := buildAddr(1, 1, msg.To)

	result, err := session.SplitMessage(msg.Text, encoding, src, dst, 1)
	if err != nil {
		log.Printf("message %s: split error: %v", msg.ID, err)
		_ = w.svc.UpdateMessageStatus(ctx, msg.ID, "failed", 0, 0)
		return
	}

	cost := w.svc.EstimateCost(len(result.Parts))
	_ = w.svc.UpdateMessageStatus(ctx, msg.ID, "sending", len(result.Parts), cost)

	for _, sm := range result.Parts {
		if err := w.pool.Send(sm); err != nil {
			log.Printf("message %s: submit error: %v", msg.ID, err)
			_ = w.svc.UpdateMessageStatus(ctx, msg.ID, "failed", len(result.Parts), cost)
			return
		}
	}

	if _, err := w.svc.DeductBalance(ctx, msg.AccountID, cost, "sms:"+msg.ID, "sms"); err != nil {
		log.Printf("message %s: deduct balance error: %v", msg.ID, err)
	}

	_ = w.svc.UpdateMessageStatus(ctx, msg.ID, "sent", len(result.Parts), cost)
}

func buildAddr(ton, npi byte, addr string) pdu.Address {
	a := pdu.NewAddress()
	a.SetTon(ton)
	a.SetNpi(npi)
	_ = a.SetAddress(addr)
	return a
}
