package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/chilljzz/gohub/internal/event"
	"github.com/chilljzz/gohub/internal/model"
	"github.com/chilljzz/gohub/pkg/config"
	"github.com/twmb/franz-go/pkg/kgo"
)

type KafkaMessagePublisher struct {
	client *kgo.Client

	topic string
}

func NewKafkaMessagePublisher(
	cfg config.KafkaConfig,
) (*KafkaMessagePublisher, error) {

	broker := cfg.Brokers

	if len(broker) == 0 {
		return nil, errors.New("kafka brokers required")
	}

	if cfg.MessageCreatedTopic == "" {
		return nil, errors.New("kafka message created topic required")
	}

	client, err := kgo.NewClient(
		kgo.SeedBrokers(broker...),
		kgo.ClientID("gohub-api"),
		kgo.RecordPartitioner(kgo.StickyKeyPartitioner(nil)),
	)

	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx); err != nil {
		client.Close()

		return nil, fmt.Errorf("ping kafka: %w", err)
	}

	return &KafkaMessagePublisher{
		client: client,
		topic:  cfg.MessageCreatedTopic,
	}, nil

}

func (p *KafkaMessagePublisher) PublishMessageCreated(
	ctx context.Context,
	message *model.Message,
) error {

	e := event.MessageCreated{
		EventID: fmt.Sprintf("message.created:%d", message.ID),

		Type: event.TypeMessageCreated,

		Version: event.MessageCreatedV1,

		OccurredAt: time.Now().UTC(),

		MessageID: message.ID,

		ConversationID: message.ConversationID,

		SenderID: message.SenderID,

		ClientMessageID: message.ClientMessageID,

		Content: message.Content,

		CreatedAt: message.CreatedAt,
	}

	payload, err := json.Marshal(e)

	if err != nil {
		return fmt.Errorf(
			"marshal message created event: %w",
			err,
		)
	}

	key := strconv.FormatUint(uint64(message.ConversationID), 10)

	record := &kgo.Record{
		Topic: p.topic,

		Key: []byte(key),

		Value: payload,

		Headers: []kgo.RecordHeader{
			{
				Key:   "event_type",
				Value: []byte(event.TypeMessageCreated),
			},
			{
				Key:   "event_version",
				Value: []byte("1"),
			},
		},
	}
	if err := p.client.ProduceSync(ctx, record).FirstErr(); err != nil {
		return fmt.Errorf(
			"produce message.created: %w",
			err,
		)
	}

	return nil

}

func (p *KafkaMessagePublisher) Close() {
	p.client.Close()
}
