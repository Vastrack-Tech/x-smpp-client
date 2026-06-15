package main

import (
	"context"
	"log"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/linxGnu/gosmpp/pdu"
	"x-smpp-client/internal/api"
	"x-smpp-client/internal/config"
	"x-smpp-client/internal/queue"
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

	msgQueue := queue.New(cfg.Server.QueueSize)

	worker := queue.NewWorker(pool, msgQueue, queue.Defaults{
		SourceAddr: cfg.SourceAddr.Address,
		SourceTon:  cfg.SourceAddr.Ton,
		SourceNpi:  cfg.SourceAddr.Npi,
		Encoding:   cfg.Encoding,
	})

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		worker.Start(ctx)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		srv := api.New(msgQueue, cfg.Server)
		if err := srv.Serve(ctx); err != nil {
			log.Printf("api server error: %v", err)
		}
	}()

	wg.Wait()
	pool.Close()
	log.Println("shutdown complete")
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
					receipt.SubmitDate.Format("2006-01-02 15:04:05"),
					receipt.DoneDate.Format("2006-01-02 15:04:05"))
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
