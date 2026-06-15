package queue

import (
	"context"
	"log"

	"github.com/linxGnu/gosmpp/pdu"
	"x-smpp-client/internal/message"
	"x-smpp-client/internal/session"
)

type Queue struct {
	ch chan message.Message
}

func New(size int) *Queue {
	return &Queue{
		ch: make(chan message.Message, size),
	}
}

func (q *Queue) Push(msg message.Message) {
	q.ch <- msg
}

func (q *Queue) Consume() <-chan message.Message {
	return q.ch
}

func (q *Queue) Len() int {
	return len(q.ch)
}

type Worker struct {
	pool *session.Pool
	q    *Queue
	def  Defaults
}

type Defaults struct {
	SourceAddr string
	SourceTon  byte
	SourceNpi  byte
	Encoding   string
}

func NewWorker(pool *session.Pool, q *Queue, def Defaults) *Worker {
	return &Worker{pool: pool, q: q, def: def}
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

func (w *Worker) process(ctx context.Context, msg message.Message) {
	enc := w.def.Encoding
	if msg.Encoding != "" {
		enc = msg.Encoding
	}
	encoding, err := message.ParseEncoding(enc)
	if err != nil {
		log.Printf("message %s: %v", msg.ID, err)
		return
	}

	srcAddr := msg.SourceAddr
	srcTon := msg.SourceTon
	srcNpi := msg.SourceNpi
	if srcAddr == "" {
		srcAddr = w.def.SourceAddr
		srcTon = w.def.SourceTon
		srcNpi = w.def.SourceNpi
	}

	destTon := msg.ToTon
	destNpi := msg.ToNpi
	if destTon == 0 && destNpi == 0 {
		destTon = 1
		destNpi = 1
	}

	src := buildAddr(srcTon, srcNpi, srcAddr)
	dst := buildAddr(destTon, destNpi, msg.To)

	result, err := session.SplitMessage(msg.Text, encoding, src, dst, 1)
	if err != nil {
		log.Printf("message %s: split error: %v", msg.ID, err)
		return
	}

	for _, sm := range result.Parts {
		if err := w.pool.Send(sm); err != nil {
			log.Printf("message %s: submit error: %v", msg.ID, err)
		}
	}
}

func buildAddr(ton, npi byte, addr string) pdu.Address {
	a := pdu.NewAddress()
	a.SetTon(ton)
	a.SetNpi(npi)
	_ = a.SetAddress(addr)
	return a
}
