package bootstrap

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"x-smpp-client/infra/cache"
	"x-smpp-client/internal/accounts/repository"
	"x-smpp-client/internal/accounts/service"
	"x-smpp-client/internal/api"
	authservice "x-smpp-client/internal/auth/service"
	"x-smpp-client/internal/config"
	"x-smpp-client/internal/database"
	"x-smpp-client/internal/queue"
	"x-smpp-client/internal/session"

	"github.com/linxGnu/gosmpp/pdu"
)

func Run(cfgPath string) error {
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		return err
	}
	log.Printf("cfg.DatabaseDSN=%q", cfg.DatabaseDSN)
    log.Printf("os.Getenv(DATABASE_DSN)=%q", os.Getenv("DATABASE_DSN"))

	log.Printf("cfg.CacheAddr=%q", cfg.CacheAddr)
    log.Printf("os.Getenv(CACHE_ADDR)=%q", os.Getenv("CACHE_ADDR"))
    log.Printf("os.Getenv(CACHE_PASSWORD)=%q", os.Getenv("CACHE_PASSWORD"))
    log.Printf("os.Getenv(CACHE_DB)=%q", os.Getenv("CACHE_DB"))

    log.Printf("cfg.EnquireLink=%d", cfg.EnquireLink)
    log.Printf("cfg.ReadTimeout=%d", cfg.ReadTimeout)
    log.Printf("os.Getenv(ENQUIRE_LINK)=%q", os.Getenv("ENQUIRE_LINK"))
    log.Printf("os.Getenv(READ_TIMEOUT)=%q", os.Getenv("READ_TIMEOUT"))
	
	// startup database and redis (migrations)
	db, err := database.New(context.Background(), cfg.DatabaseDSN)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := db.Migrate(context.Background()); err != nil {
		return err
	}
	log.Printf("CACHE_ADDR=%q", cfg.CacheAddr)
	rdb, err := database.NewRedis(context.Background(), cfg.CacheAddr, cfg.CachePassword, cfg.CacheDB)
	if err != nil {
		return err
	}
	defer rdb.Close()
	//  // /// //

	// initialize repository and service
	repo := repository.NewPostgresRepo(db)
	svc := service.New(repo)

	sessionStore := cache.NewSessionStore(rdb.Client)
	authSvc := authservice.New(repo, sessionStore, cfg.JWTSecret)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// setup session pool (number of managers of the underlying gosmpp connections)
	pool := session.NewPool(cfg)
	pool.OnPDU(makePDUHandler(svc))

	if err := pool.Start(ctx); err != nil {
		return err
	}

	msgQueue := queue.New(cfg.ServerQueueSize)

	worker := queue.NewWorker(pool, msgQueue, svc, queue.Defaults{
		SourceAddr: cfg.SourceAddr,
		SourceTon:  cfg.SourceTon,
		SourceNpi:  cfg.SourceNpi,
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
		srv := api.New(msgQueue, svc, authSvc, cfg.ServerListenAddr)
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
