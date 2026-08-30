package realtime

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/redis/go-redis/v9"
	"golang.org/x/net/context"
)

const channelTopicPrefix = "gohub:channel:"

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

func ChannelTopic(channelID uint) string {
	return fmt.Sprintf(
		"%s%d",
		channelTopicPrefix,
		channelID,
	)
}

func ParseChannelTopic(
	topic string,
) (uint, error) {
	if !strings.HasPrefix(
		topic,
		channelTopicPrefix,
	) {
		return 0, fmt.Errorf(
			"incalid channel topic: %s",
			topic,
		)
	}

	idStr := strings.TrimPrefix(
		topic,
		channelTopicPrefix,
	)
	id, err := strconv.ParseUint(idStr, 10, 64)

	if err != nil {
		return 0, fmt.Errorf(
			"parse channel id failed: %w",
			err,
		)
	}

	if id == 0 {
		return 0, fmt.Errorf(
			"invalid channel id: %d",
			id,
		)
	}

	return uint(id), nil

}

func (b *RedisBroker) PublishChannel(
	ctx context.Context,
	channelID uint,
	message []byte,
) error {
	topic := ChannelTopic(channelID)

	return b.client.Publish(
		ctx,
		topic,
		message,
	).Err()
}

func (b *RedisBroker) SubscribeChannels(
	ctx context.Context,
	handler func(
		channel string,
		payload []byte,
	),
) error {
	pubsub := b.client.PSubscribe(
		ctx,
		channelTopicPrefix+"*",
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
