package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/chilljzz/gohub/internal/event"
	"github.com/chilljzz/gohub/pkg/config"
	"github.com/twmb/franz-go/pkg/kgo"
)

type MessageCreatedConsumer struct {
	client *kgo.Client
}

func NewMessageCreatedConsumer(
	cfg config.KafkaConfig,
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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx); err != nil {
		client.Close()
		return nil, fmt.Errorf("ping kafka consumer: %w", err)
	}

	return &MessageCreatedConsumer{
		client: client,
	}, nil

}

func (c *MessageCreatedConsumer) Run(ctx context.Context) {
	for {
		fetches := c.client.PollFetches(ctx)

		if ctx.Err() != nil {
			return
		}

		for _, fetchErr := range fetches.Errors() {
			log.Printf(
				"kafka consume error: topic=%s partition=%d err=%s",
				fetchErr.Topic,
				fetchErr.Partition,
				fetchErr.Err,
			)
		}

		iter := fetches.RecordIter()

		for !iter.Done() {
			record := iter.Next()

			var e event.MessageCreated

			if err := json.Unmarshal(record.Value, &e); err != nil {
				log.Printf(
					"decode message.created failed: topic=%s partition=%d offset=%d err=%v",
					record.Topic,
					record.Partition,
					record.Offset,
					err,
				)
				continue
			}

			log.Printf(
				"kafka message.created consumed: event_id=%s message_id=%d conversation_id=%d partition=%d offset=%d",
				e.EventID,
				e.MessageID,
				e.ConversationID,
				record.Partition,
				record.Offset,
			)
		}

	}
}

func (
	c *MessageCreatedConsumer,
) Close() {
	c.client.Close()
}
