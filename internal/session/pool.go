package session

import (
	"context"
	"log"
	"sync/atomic"

	"x-smpp-client/internal/config"

	"github.com/linxGnu/gosmpp/pdu"
)

const DefaultPoolSize = 1

func PoolSize(n int) int {
	if n < 1 {
		return DefaultPoolSize
	}
	return n
}

type Pool struct {
	managers []*Manager
	next     atomic.Uint64
}

func NewPool(cfg *config.Config) *Pool {
	size := PoolSize(cfg.PoolSize)
	managers := make([]*Manager, size)
	for i := 0; i < size; i++ {
		managers[i] = NewManager(cfg)
	}
	return &Pool{managers: managers}
}

func (p *Pool) Start(ctx context.Context) error {
	for i, m := range p.managers {
		if err := m.Connect(ctx); err != nil {
			return err
		}
		log.Printf("session %d connected", i)
	}
	return nil
}

func (p *Pool) Send(sm *pdu.SubmitSM) error {
	idx := p.next.Add(1) - 1
	m := p.managers[idx%uint64(len(p.managers))]
	return m.Send(sm)
}

func (p *Pool) OnPDU(handler PDUHandler) {
	for _, m := range p.managers {
		m.OnPDU(handler)
	}
}

func (p *Pool) Close() error {
	for _, m := range p.managers {
		_ = m.Close()
	}
	return nil
}
