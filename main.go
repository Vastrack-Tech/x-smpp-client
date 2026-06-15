package main

import (
	"context"
	"log"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/linxGnu/gosmpp/data"
	"github.com/linxGnu/gosmpp/pdu"
	"x-smpp-client/internal/config"
	"x-smpp-client/internal/session"
)

func main() {
	cfg, err := config.LoadConfig("")
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool := session.NewPool(cfg)
	pool.OnPDU(handlePDU())

	if err := pool.Start(ctx); err != nil {
		log.Fatalf("start session pool: %v", err)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go sendLoop(ctx, &wg, cfg, pool)

	wg.Wait()
	pool.Close()
	log.Println("shutdown complete")
}

func sendLoop(ctx context.Context, wg *sync.WaitGroup, cfg *config.Config, pool *session.Pool) {
	defer wg.Done()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("shutting down sender")
			return
		case <-ticker.C:
		}

		srcAddr := buildAddr(cfg.SourceAddr.Ton, cfg.SourceAddr.Npi, cfg.SourceAddr.Address)
		destAddr := buildAddr(cfg.DefaultDest.Ton, cfg.DefaultDest.Npi, cfg.DefaultDest.Address)

		result, err := session.SplitMessage("Hello World ", data.GSM7BIT, srcAddr, destAddr, 1)
		if err != nil {
			log.Printf("split message error: %v", err)
			continue
		}

		for _, sm := range result.Parts {
			if err := pool.Send(sm); err != nil {
				log.Printf("submit error: %v", err)
			}
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

func handlePDU() session.PDUHandler {
	concatenated := map[byte][]string{}
	return func(p pdu.PDU) {
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

			if session.IsDeliveryReceipt(pd) {
				receipt, err := session.ParseDeliveryReceipt(message)
				if err != nil {
					log.Printf("parse delivery receipt error: %v", err)
					return
				}
				log.Printf("DeliveryReceipt: id=%s stat=%s err=%s submit=%s done=%s",
					receipt.MessageID, receipt.Status, receipt.Err,
					receipt.SubmitDate.Format(time.RFC3339),
					receipt.DoneDate.Format(time.RFC3339))
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

func isConcatenatedDone(parts []string, total byte) bool {
	for _, part := range parts {
		if part != "" {
			total--
		}
	}
	return total == 0
}
