package bootstrap

import (
	"context"
	"log"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/linxGnu/gosmpp/pdu"
	"x-smpp-client/infra/cache"
	"x-smpp-client/internal/accounts/repository"
	"x-smpp-client/internal/accounts/service"
	"x-smpp-client/internal/api"
	authservice "x-smpp-client/internal/auth/service"
	"x-smpp-client/internal/config"
	"x-smpp-client/internal/database"
	"x-smpp-client/internal/queue"
	"x-smpp-client/internal/session"
)

func Run(cfgPath string) error {
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		return err
	}

	db, err := database.New(context.Background(), cfg.Database.DSN)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := db.Migrate(context.Background()); err != nil {
		return err
	}

	rdb, err := database.NewRedis(context.Background(), cfg.Cache.Addr, cfg.Cache.Password, cfg.Cache.DB)
	if err != nil {
		return err
	}
	defer rdb.Close()

	repo := repository.NewPostgresRepo(db)
	svc := service.New(repo)

	sessionStore := cache.NewSessionStore(rdb.Client)
	authSvc := authservice.New(repo, sessionStore, cfg.JWTSecret)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool := session.NewPool(cfg)
	pool.OnPDU(makePDUHandler(svc))

	if err := pool.Start(ctx); err != nil {
		return err
	}

	msgQueue := queue.New(cfg.Server.QueueSize)

	worker := queue.NewWorker(pool, msgQueue, svc, queue.Defaults{
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
		srv := api.New(msgQueue, svc, authSvc, cfg.Server)
		if err := srv.Serve(ctx); err != nil {
			log.Printf("api server error: %v", err)
		}
	}()

	wg.Wait()
	pool.Close()
	log.Println("shutdown complete")
	return nil
}

func makePDUHandler(svc *service.Service) session.PDUHandler {
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
				log.Printf("DeliveryReceipt: id=%s stat=%s err=%s",
					receipt.MessageID, receipt.Status, receipt.Err)

				dr := receipt.ToModel()
				_ = svc.SaveDeliveryReceipt(context.Background(), &dr)
				_ = svc.UpdateMessageStatus(context.Background(), receipt.MessageID, "delivered", 0, 0)
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
