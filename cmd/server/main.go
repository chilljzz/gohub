package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/chilljzz/gohub/internal/app"
	"github.com/chilljzz/gohub/internal/database"
	"github.com/chilljzz/gohub/internal/logging"
	"github.com/chilljzz/gohub/pkg/config"
)

func main() {
	if err := config.InitConfig(); err != nil {
		log.Fatalf("init config failed: %v", err)
	}

	c := config.Conf

	logger := logging.New(c.Server.Mode)

	slog.SetDefault(logger)

	if err := database.InitMySQL(c.Mysql); err != nil {
		log.Fatalf("init mysql failed: %v", err)
	}

	if err := database.InitRedis(c.Redis); err != nil {
		log.Fatalf("init redis failed: %v", err)
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	application, err := app.New(ctx)
	if err != nil {
		log.Fatalf("init app failed: %v", err)
	}

	if err := application.Run(); err != nil {
		log.Fatalf("server run failed: %v", err)
	}

}
