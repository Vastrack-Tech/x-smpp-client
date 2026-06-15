package main

import (
	"context"
	"log"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/linxGnu/gosmpp"
	"github.com/linxGnu/gosmpp/data"
	"github.com/linxGnu/gosmpp/pdu"
	"x-smpp-client/internal/config"

	"crypto/tls"
	"net"
)

var (
	TLSDialer = func(addr string) (net.Conn, error) {
		conf := &tls.Config{
			InsecureSkipVerify: true,
		}
		return tls.Dial("tcp", addr, conf)
	}
)

func main() {
	cfg, err := config.LoadConfig("")
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	var wg sync.WaitGroup
	wg.Add(1)
	go runSession(ctx, &wg, cfg)

	wg.Wait()
	log.Println("shutdown complete")
}

func runSession(ctx context.Context, wg *sync.WaitGroup, cfg *config.Config) {
	defer wg.Done()

	auth := gosmpp.Auth{
		SMSC:       cfg.SMSC.Addr,
		SystemID:   cfg.SMSC.SystemID,
		Password:   cfg.SMSC.Password,
		SystemType: cfg.SMSC.SystemType,
	}

	dialer := gosmpp.NonTLSDialer
	if cfg.TLS.Enabled {
		dialer = TLSDialer
	}

	session, err := gosmpp.NewSession(
		gosmpp.TRXConnector(dialer, auth),
		gosmpp.Settings{
			EnquireLink: cfg.App.EnquireLink,
			ReadTimeout: cfg.App.ReadTimeout,

			OnSubmitError: func(_ pdu.PDU, err error) {
				log.Printf("SubmitPDU error: %v", err)
			},

			OnReceivingError: func(err error) {
				log.Printf("Receiving PDU/Network error: %v", err)
			},

			OnRebindingError: func(err error) {
				log.Printf("Rebinding error: %v", err)
			},

			OnPDU: handlePDU(),

			OnClosed: func(state gosmpp.State) {
				log.Printf("Session closed: %v", state)
			},
		}, cfg.App.WriteTimeout)
	if err != nil {
		log.Printf("Failed to create session: %v", err)
		return
	}
	defer func() {
		_ = session.Close()
	}()

	log.Printf("connected to SMSC %s as %s", cfg.SMSC.Addr, cfg.SMSC.SystemID)

	for {
		select {
		case <-ctx.Done():
			log.Println("shutting down sender")
			return
		default:
		}

		sm := newSubmitSM(cfg)
		if err := session.Transceiver().Submit(sm); err != nil {
			log.Printf("submit error: %v", err)
		}

		select {
		case <-ctx.Done():
			log.Println("shutting down sender")
			return
		case <-time.After(5 * time.Second):
		}
	}
}

func handlePDU() func(pdu.PDU, bool) {
	concatenated := map[uint8][]string{}
	return func(p pdu.PDU, _ bool) {
		switch pd := p.(type) {
		case *pdu.SubmitSMResp:
			log.Printf("SubmitSMResp: message_id=%s", pd.MessageID)

		case *pdu.GenericNack:
			log.Println("GenericNack Received")

		case *pdu.EnquireLinkResp:
			log.Println("EnquireLinkResp Received")

		case *pdu.DataSM:
			log.Printf("DataSM: %+v", pd)

		case *pdu.DeliverSM:
			message, err := pd.Message.GetMessage()
			if err != nil {
				log.Printf("failed to get DeliverSM message: %v", err)
				return
			}
			totalParts, sequence, reference, found := pd.Message.UDH().GetConcatInfo()
			if found {
				if _, ok := concatenated[reference]; !ok {
					concatenated[reference] = make([]string, totalParts)
				}
				concatenated[reference][sequence-1] = message
				if parts, ok := concatenated[reference]; ok && isConcatenatedDone(parts, totalParts) {
					log.Printf("DeliverSM (concatenated): %s", strings.Join(parts, ""))
					delete(concatenated, reference)
				}
			} else {
				log.Printf("DeliverSM: %s", message)
			}
		}
	}
}

func newSubmitSM(cfg *config.Config) *pdu.SubmitSM {
	srcAddr := pdu.NewAddress()
	srcAddr.SetTon(cfg.SourceAddr.Ton)
	srcAddr.SetNpi(cfg.SourceAddr.Npi)
	_ = srcAddr.SetAddress(cfg.SourceAddr.Address)

	destAddr := pdu.NewAddress()
	destAddr.SetTon(cfg.DefaultDest.Ton)
	destAddr.SetNpi(cfg.DefaultDest.Npi)
	_ = destAddr.SetAddress(cfg.DefaultDest.Address)

	submitSM := pdu.NewSubmitSM().(*pdu.SubmitSM)
	submitSM.SourceAddr = srcAddr
	submitSM.DestAddr = destAddr
	_ = submitSM.Message.SetMessageWithEncoding("Hello World ", data.UCS2)
	submitSM.ProtocolID = 0
	submitSM.RegisteredDelivery = 1
	submitSM.ReplaceIfPresentFlag = 0
	submitSM.EsmClass = 0

	return submitSM
}

func isConcatenatedDone(parts []string, total byte) bool {
	for _, part := range parts {
		if part != "" {
			total--
		}
	}
	return total == 0
}
