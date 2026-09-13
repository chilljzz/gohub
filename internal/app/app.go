package app

import (
	"context"
	"fmt"
	"log"
	"runtime/debug"

	"github.com/chilljzz/gohub/internal/database"
	"github.com/chilljzz/gohub/internal/realtime"
	"github.com/chilljzz/gohub/internal/router"
	"github.com/chilljzz/gohub/internal/ws"
	"github.com/chilljzz/gohub/pkg/config"
	"github.com/gin-gonic/gin"
)

type App struct {
	router  *gin.Engine
	manager *ws.Manager
	broker  *realtime.RedisBroker
}

func New() *App {
	manager := ws.NewManager()

	broker := realtime.NewRedisBroker(
		database.RedisClient,
	)

	r := router.InitRouter(
		broker,
		manager,
	)
	return &App{
		router:  r,
		manager: manager,
		broker:  broker,
	}
}

func (a *App) startRedisSubscriber(
	ctx context.Context,
) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Printf(
					"redis subscriber panic: %v\n%s",
					err,
					debug.Stack(),
				)
			}

		}()
		err := a.broker.SubscribeConversations(
			ctx,
			func(topic string, payload []byte) {
				conversationID, err := realtime.ParseConversationTopic(topic)
				if err != nil {
					log.Printf(
						"invalid redis conversation=%s err=%v",
						topic,
						err,
					)
					return
				}

				count := a.manager.BroadcastToConversation(
					conversationID,
					payload,
				)

				log.Printf(
					"redis broadcast: conversation_id=%d clients=%d",
					conversationID,
					count,
				)
			},
		)
		if err != nil {
			log.Printf(
				"redis subscriber stopped: %v",
				err,
			)
		}
	}()
}

func (a *App) Run(
	ctx context.Context,
) error {
	a.startRedisSubscriber(ctx)

	addr := fmt.Sprintf(
		":%d",
		config.Conf.Server.Port,
	)

	log.Printf(
		"server starting: addr=%s",
		addr,
	)
	return a.router.Run(addr)
}
