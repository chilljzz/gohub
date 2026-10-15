package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/chilljzz/gohub/internal/database"
	"github.com/chilljzz/gohub/internal/logging"
	"github.com/chilljzz/gohub/internal/messaging"
	"github.com/chilljzz/gohub/internal/outbox"
	"github.com/chilljzz/gohub/pkg/config"
)

func main() {
	if err := config.InitConfig(); err != nil {
		log.Fatalf("init config: %v", err)
	}
	slog.SetDefault(logging.New(config.Conf.Server.Mode))

	db, err := database.OpenMySQL(config.Conf.Mysql)
	if err != nil {
		log.Fatalf("open worker mysql: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("get sql db: %v", err)
	}
	defer sqlDB.Close()

	publisher, err := messaging.NewKafkaMessagePublisher(config.Conf.Kafka)
	if err != nil {
		log.Fatalf("new kafka publisher: %v", err)
	}
	defer publisher.Close()

	worker, err := outbox.NewWorker(outbox.NewRepository(db), publisher, outbox.Config{
		PollInterval:   500 * time.Millisecond,
		LeaseTTL:       30 * time.Second,
		PublishTimeout: 5 * time.Second,
		MaxAttempts:    8,
	})
	if err != nil {
		log.Fatalf("new worker: %v", err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	slog.Info("outbox worker started")
	if err := worker.Run(ctx); err != nil {
		log.Fatalf("worker stopped: %v", err)
	}
	slog.Info("outbox worker stopped")
}
