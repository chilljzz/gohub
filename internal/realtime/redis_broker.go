package realtime

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/redis/go-redis/v9"
)

const conversationTopicPrefix = "gohub:conversation:"

type RedisBroker struct {
	client *redis.Client
}

func NewRedisBroker(
	client *redis.Client,
) *RedisBroker {
	return &RedisBroker{
		client: client,
	}
}

func ConversationTopic(conversationID uint) string {
	return fmt.Sprintf(
		"%s%d",
		conversationTopicPrefix,
		conversationID,
	)
}

func ParseConversationTopic(
	topic string,
) (uint, error) {
	if !strings.HasPrefix(
		topic,
		conversationTopicPrefix,
	) {
		return 0, fmt.Errorf(
			"invalid conversation topic: %s",
			topic,
		)
	}

	idStr := strings.TrimPrefix(
		topic,
		conversationTopicPrefix,
	)
	id, err := strconv.ParseUint(idStr, 10, 64)

	if err != nil {
		return 0, fmt.Errorf(
			"parse conversation id failed: %w",
			err,
		)
	}

	if id == 0 {
		return 0, fmt.Errorf(
			"invalid conversation id: %d",
			id,
		)
	}

	return uint(id), nil

}

func (b *RedisBroker) PublishConversation(
	ctx context.Context,
	conversationID uint,
	message []byte,
) error {

	if conversationID == 0 {
		return fmt.Errorf(
			"conversation id is required",
		)
	}

	topic := ConversationTopic(conversationID)

	return b.client.Publish(
		ctx,
		topic,
		message,
	).Err()
}

func (b *RedisBroker) SubscribeConversations(
	ctx context.Context,
	handler func(
		topic string,
		payload []byte,
	),
) error {
	pubsub := b.client.PSubscribe(
		ctx,
		conversationTopicPrefix+"*",
	)
	defer pubsub.Close()

	if _, err := pubsub.Receive(ctx); err != nil {
		return err
	}

	messages := pubsub.Channel()

	for msg := range messages {
		handler(msg.Channel, []byte(msg.Payload))
	}
	return nil
}
