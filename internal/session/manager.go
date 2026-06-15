package session

import (
	"context"
	"crypto/tls"
	"log"
	"net"
	"sync"
	"time"

	"github.com/linxGnu/gosmpp"
	"github.com/linxGnu/gosmpp/pdu"
	"x-smpp-client/internal/config"
)

type Manager struct {
	smsc config.SMSCConfig
	tls  config.TLSConfig
	app  config.AppConfig

	session *gosmpp.Session
	mu      sync.RWMutex
	handler PDUHandler

	ctx    context.Context
	cancel context.CancelFunc
}

func NewManager(smsc config.SMSCConfig, tlsCfg config.TLSConfig, app config.AppConfig) *Manager {
	return &Manager{
		smsc: smsc,
		tls:  tlsCfg,
		app:  app,
	}
}

func (m *Manager) Connect(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	auth := gosmpp.Auth{
		SMSC:       m.smsc.Addr,
		SystemID:   m.smsc.SystemID,
		Password:   m.smsc.Password,
		SystemType: m.smsc.SystemType,
	}

	dialer := gosmpp.NonTLSDialer
	if m.tls.Enabled {
		dialer = m.tlsDialer()
	}

	m.ctx, m.cancel = context.WithCancel(ctx)

	session, err := gosmpp.NewSession(
		gosmpp.TRXConnector(dialer, auth),
		gosmpp.Settings{
			EnquireLink: m.app.EnquireLink,
			ReadTimeout: m.app.ReadTimeout,

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
		}, m.app.WriteTimeout)
	if err != nil {
		return err
	}

	m.session = session
	log.Printf("connected to SMSC %s as %s", m.smsc.Addr, m.smsc.SystemID)
	return nil
}

func (m *Manager) Send(sm *pdu.SubmitSM) error {
	m.mu.RLock()
	session := m.session
	m.mu.RUnlock()

	if session == nil {
		return nil
	}
	return session.Transceiver().Submit(sm)
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
	skip := m.tls.SkipVerify
	return func(addr string) (net.Conn, error) {
		conf := &tls.Config{
			InsecureSkipVerify: skip,
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
