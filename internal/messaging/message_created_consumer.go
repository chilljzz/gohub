package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/chilljzz/gohub/internal/event"
	"github.com/chilljzz/gohub/pkg/config"
	"github.com/twmb/franz-go/pkg/kgo"
)

type MessageCreatedConsumer struct {
	client *kgo.Client
	ctx    context.Context
}

func NewMessageCreatedConsumer(
	cfg config.KafkaConfig,
	ctx context.Context,
) (*MessageCreatedConsumer, error) {

	brokers := cfg.Brokers

	if len(brokers) == 0 {
		return nil, errors.New("kafka brokers required")
	}

	if cfg.ConsumerGroup == "" {
		return nil, errors.New("kafka consumer group required")
	}

	client, err :=
		kgo.NewClient(
			kgo.SeedBrokers(brokers...),
			kgo.ClientID("gohub-message-audit"),
			kgo.ConsumerGroup(cfg.ConsumerGroup),
			kgo.ConsumeTopics(
				cfg.MessageCreatedTopic,
			),
		)

	if err != nil {
		return nil, err
	}

	processCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := client.Ping(processCtx); err != nil {
		client.Close()
		return nil, fmt.Errorf("ping kafka consumer: %w", err)
	}

	return &MessageCreatedConsumer{
		client: client,
	}, nil

}

func (c *MessageCreatedConsumer) Run(
	ctx context.Context,
) error {

	for {
		fetches := c.client.PollFetches(ctx)

		if ctx.Err() != nil {
			return nil
		}

		if fetches.IsClientClosed() {
			return nil
		}

		for _, fetchErr := range fetches.Errors() {
			slog.Error(
				"kafka consume faild",
				slog.String("topic", fetchErr.Topic),
				slog.Int("partition", int(fetchErr.Partition)),
				slog.Any("error", fetchErr.Err),
			)
		}

		iter := fetches.RecordIter()

		for !iter.Done() {
			record := iter.Next()

			var e event.MessageCreated

			if err := json.Unmarshal(record.Value, &e); err != nil {
				slog.Error(
					"decode message.created failed",

					slog.String("topic", record.Topic),

					slog.Int("partition", int(record.Partition)),

					slog.Int64("offset", record.Offset),

					slog.Any("error", err),
				)
				continue
			}

			slog.Debug(
				"message.created consumed",
				slog.String("event_id", e.EventID),

				slog.Uint64("message_id", uint64(e.MessageID)),

				slog.Uint64("conversation_id", uint64(e.ConversationID)),

				slog.Int("partition", int(record.Partition)),

				slog.Int64("offset", record.Offset),
			)
		}

	}
}

func (
	c *MessageCreatedConsumer,
) Close() {
	c.client.Close()
}
