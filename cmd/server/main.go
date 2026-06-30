package main

import (
	"fmt"
	"log"

	"github.com/chilljzz/gohub/configs/config"
	"github.com/chilljzz/gohub/internal/router"
)

func main() {

	err := config.InitConfig()
	if err != nil {
		log.Fatalf("init config failed: %v", err)
	}

	r := router.InitRouter()
	addr := fmt.Sprintf(":%d", config.Conf.Server.Port)

	if err := r.Run(addr); err != nil {
		log.Fatalf("server run failed: %v", err)
	}
}
