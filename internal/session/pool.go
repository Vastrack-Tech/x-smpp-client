package session

import (
	"context"
	"log"
	"sync/atomic"

	"github.com/linxGnu/gosmpp/pdu"
	"x-smpp-client/internal/config"
)

const DefaultPoolSize = 1

func PoolSize(cfg config.AppConfig) int {
	if cfg.PoolSize < 1 {
		return DefaultPoolSize
	}
	return cfg.PoolSize
}

type Pool struct {
	managers []*Manager
	next     atomic.Uint64
}

func NewPool(cfg *config.Config) *Pool {
	size := PoolSize(cfg.App)
	managers := make([]*Manager, size)
	for i := 0; i < size; i++ {
		managers[i] = NewManager(cfg.SMSC, cfg.TLS, cfg.App)
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
