package service

import (
	"context"
	"log"
	"time"

	"github.com/chilljzz/gohub/internal/model"
)

type MessageEventPublisher interface {
	PublishMessageCreated(
		ctx context.Context,
		message *model.Message,
	) error
}

func publishMessageCreatedBestEffort(
	publisher MessageEventPublisher,
	message *model.Message,
) {
	if publisher == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)

	defer cancel()

	if err := publisher.PublishMessageCreated(ctx, message); err != nil {
		log.Printf(
			"publish message created event failed: message_id=%d conversation_id=%d err=%v",
			message.ID,
			message.ConversationID,
			err,
		)
	}
}
