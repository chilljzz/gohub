package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/chilljzz/gohub/internal/event"
	"github.com/chilljzz/gohub/internal/outbox"
	"github.com/twmb/franz-go/pkg/kgo"
)

func (p *KafkaMessagePublisher) PublishOutbox(ctx context.Context, e outbox.Record) error {
	if e.EventType != event.TypeMessageCreated || e.AggregateType != "conversation" {
		return fmt.Errorf("unsupported outbox event: %s / %s", e.AggregateType, e.EventType)
	}

	var payload event.MessageCreated

	if err := json.Unmarshal([]byte(e.Payload), &payload); err != nil {
		return fmt.Errorf("decode outbox payload: %w", err)
	}

	if payload.Version != event.MessageCreatedV1 || payload.EventID != e.EventID ||
		payload.Type != e.EventType || uint64(payload.ConversationID) != e.AggregateID {
		return fmt.Errorf("outbox payload metadate mismatch: %s", e.EventID)
	}

	record := &kgo.Record{
		Topic: p.topic,
		Key:   []byte(strconv.FormatUint(e.AggregateID, 10)),
		Value: []byte(e.Payload),
		Headers: []kgo.RecordHeader{
			{Key: "event_id", Value: []byte(e.EventID)},
			{Key: "event_type", Value: []byte(e.EventType)},
			{Key: "event_version", Value: []byte(strconv.Itoa(payload.Version))},
		},
	}
	if err := p.client.ProduceSync(ctx, record).FirstErr(); err != nil {
		return fmt.Errorf("produce stored outbox event %s: %w", e.EventID, err)
	}
	return nil
}
