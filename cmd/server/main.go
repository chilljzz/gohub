package main

import (
	"fmt"
	"log"

	"github.com/chilljzz/gohub/internal/database"
	"github.com/chilljzz/gohub/internal/router"
	"github.com/chilljzz/gohub/pkg/config"
)

func main() {

	if err := config.InitConfig(); err != nil {
		log.Fatalf("init config failed: %v", err)
	}

	if err := database.InitMySQL(); err != nil {
		log.Fatalf("init mysql failed: %v", err)
	}

	r := router.InitRouter()
	addr := fmt.Sprintf(":%d", config.Conf.Server.Port)

	if err := r.Run(addr); err != nil {
		log.Fatalf("server run failed: %v", err)
	}
}
