package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"runtime/debug"
	"sync"
	"time"

	"github.com/chilljzz/gohub/internal/database"
	"github.com/chilljzz/gohub/internal/messaging"
	"github.com/chilljzz/gohub/internal/realtime"
	"github.com/chilljzz/gohub/internal/router"
	"github.com/chilljzz/gohub/internal/ws"
	"github.com/chilljzz/gohub/pkg/config"
	"github.com/gin-gonic/gin"
)

type App struct {
	router         *gin.Engine
	manager        *ws.Manager
	broker         *realtime.RedisBroker
	kafkaPublisher *messaging.KafkaMessagePublisher
	kafkaConsumer  *messaging.MessageCreatedConsumer
	ctx            context.Context
}

func New(ctx context.Context) (*App, error) {

	gin.SetMode(
		config.Conf.Server.Mode,
	)

	manager := ws.NewManager()

	broker := realtime.NewRedisBroker(
		database.RedisClient,
	)

	kafkaPublisher, err := messaging.NewKafkaMessagePublisher(config.Conf.Kafka)
	if err != nil {
		return nil, fmt.Errorf("init kafka publisher:%w", err)
	}
	kafkaConsumer, err := messaging.NewMessageCreatedConsumer(config.Conf.Kafka, ctx)

	if err != nil {
		kafkaPublisher.Close()
		return nil, fmt.Errorf("init kafka consumer: %w", err)
	}

	r := router.InitRouter(
		broker,
		manager,
		kafkaPublisher,
	)
	return &App{
		router:         r,
		manager:        manager,
		broker:         broker,
		kafkaPublisher: kafkaPublisher,
		kafkaConsumer:  kafkaConsumer,
		ctx:            ctx,
	}, nil
}

func (a *App) startRedisSubscriber(
	ctx context.Context,
	wg *sync.WaitGroup,
) {

	wg.Add(1)

	go func() {
		slog.Info("redis subscriber started")
		defer func() {
			slog.Info("redis subscriber stopped")
			wg.Done()
		}()
		defer func() {
			if err := recover(); err != nil {
				slog.Error(
					"redis subscriber panic",
					slog.Any("error", err),
					slog.String("stack", string(debug.Stack())),
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
		if err != nil && ctx.Err() == nil {
			log.Printf(
				"redis subscriber stopped: %v",
				err,
			)
		}
	}()
}

func (
	a *App,
) startKafkaConsumer(
	ctx context.Context,
	wg *sync.WaitGroup,
) {
	wg.Add(1)
	go func() {
		slog.Info("kafka consumer started")

		defer func() {
			slog.Info("kafka consumer exited")
			wg.Done()
		}()

		err := a.kafkaConsumer.Run(ctx)

		if err != nil && ctx.Err() == nil {
			slog.Error(
				"kafka consumer stopped unexpectedly",

				slog.Any(
					"error",
					err,
				),
			)
		}
	}()
}

func (a *App) Run() error {

	var wg sync.WaitGroup

	runCtx, cancel := context.WithCancel(a.ctx)
	defer cancel()

	a.startRedisSubscriber(runCtx, &wg)

	a.startKafkaConsumer(runCtx, &wg)

	defer a.kafkaConsumer.Close()

	defer a.kafkaPublisher.Close()

	addr := fmt.Sprintf(
		":%d",
		config.Conf.Server.Port,
	)

	server := &http.Server{
		Addr:              addr,
		Handler:           a.router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverErrCh := make(chan error, 1)

	go func() {
		slog.Info(
			"http server starting",

			slog.String(
				"addr",
				addr,
			),
		)

		err := server.ListenAndServe()

		if err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			serverErrCh <- err
			return
		}

		serverErrCh <- nil
	}()

	select {
	case <-a.ctx.Done():
		log.Printf("shutdown signal received")

	case err := <-serverErrCh:
		if err != nil {
			return fmt.Errorf(
				"http server failed: %w",
				err,
			)
		}

		return nil
	}

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	slog.Info("shutting down http server")

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf(
			"http server shutdown failed: %w",
			err,
		)
	}

	slog.Info("http server stopped")

	a.manager.Shutdown()

	workerDone := make(chan struct{})

	go func() {
		wg.Wait()
		close(workerDone)
	}()

	workerWaitCtx, workWaitCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer workWaitCancel()

	select {

	case <-workerDone:

		log.Printf(
			"background workers stopped",
		)

	case <-workerWaitCtx.Done():

		return fmt.Errorf(
			"background workers shutdown timeout: %w",
			workerWaitCtx.Err(),
		)
	}

	return nil
}
