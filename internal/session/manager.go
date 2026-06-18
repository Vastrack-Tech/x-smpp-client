package session

import (
	"context"
	"crypto/tls"
	"errors"
	"log"
	"net"
	"sync"
	"time"

	"github.com/linxGnu/gosmpp"
	"github.com/linxGnu/gosmpp/pdu"
	"x-smpp-client/internal/config"
)

type Manager struct {
	cfg *config.Config

	session *gosmpp.Session
	mu      sync.RWMutex
	handler PDUHandler

	ctx    context.Context
	cancel context.CancelFunc
}

func NewManager(cfg *config.Config) *Manager {
	return &Manager{cfg: cfg}
}

func (m *Manager) Connect(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	auth := gosmpp.Auth{
		SMSC:       m.cfg.SMSCAddr,
		SystemID:   m.cfg.SMSCSystemID,
		Password:   m.cfg.SMSCPassword,
		SystemType: m.cfg.SMSCSystemType,
	}

	dialer := gosmpp.NonTLSDialer
	if m.cfg.TLSEnabled {
		dialer = m.tlsDialer()
	}

	m.ctx, m.cancel = context.WithCancel(ctx)

	session, err := gosmpp.NewSession(
		gosmpp.TRXConnector(dialer, auth),
		gosmpp.Settings{
			EnquireLink: time.Duration(m.cfg.EnquireLink) * time.Second,
			ReadTimeout: time.Duration(m.cfg.ReadTimeout) * time.Second,

			OnSubmitError: func(_ pdu.PDU, err error) {
				log.Printf("SubmitPDU error: %v", err)
			},

			OnReceivingError: func(err error) {
				log.Printf("Receiving PDU/Network error: %v", err)
			},

			OnRebindingError: func(err error) {
				log.Printf("Rebinding error: %v", err)
			},

			OnPDU: func(p pdu.PDU, _ bool) {
				if m.handler != nil {
					m.handler(p)
				}
			},

			OnClosed: func(state gosmpp.State) {
				log.Printf("Session closed: %v", state)
			},
		}, time.Duration(m.cfg.WriteTimeout)*time.Second)
	if err != nil {
		return err
	}

	m.session = session
	log.Printf("connected to SMSC %s as %s", m.cfg.SMSCAddr, m.cfg.SMSCSystemID)
	return nil
}

func (m *Manager) Send(sm *pdu.SubmitSM) error {
	m.mu.RLock()
	session := m.session
	m.mu.RUnlock()

	if session == nil {
		return nil
	}

	for {
		err := session.Transceiver().Submit(sm)
		if err == nil {
			return nil
		}
		if errors.Is(err, gosmpp.ErrWindowsFull) {
			time.Sleep(50 * time.Millisecond)
			continue
		}
		return err
	}
}

func (m *Manager) OnPDU(handler PDUHandler) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.handler = handler
}

func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.cancel != nil {
		m.cancel()
	}
	if m.session != nil {
		return m.session.Close()
	}
	return nil
}

func (m *Manager) tlsDialer() gosmpp.Dialer {
	return func(addr string) (net.Conn, error) {
		conf := &tls.Config{
			InsecureSkipVerify: m.cfg.TLSSkipVerify,
		}
		return tls.Dial("tcp", addr, conf)
	}
}

func (m *Manager) WaitForConnection() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.session != nil
}

func (m *Manager) Reconnect(ctx context.Context) error {
	m.mu.Lock()
	if m.session != nil {
		_ = m.session.Close()
		m.session = nil
	}
	m.mu.Unlock()

	backoff := time.Second
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		err := m.Connect(ctx)
		if err == nil {
			return nil
		}
		log.Printf("reconnect failed: %v, retrying in %v", err, backoff)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff):
			backoff *= 2
			if backoff > 30*time.Second {
				backoff = 30 * time.Second
			}
		}
	}
}
