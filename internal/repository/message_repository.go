package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/chilljzz/gohub/internal/event"
	"github.com/chilljzz/gohub/internal/model"
	"gorm.io/gorm"
)

type MessageRepository struct {
	db *gorm.DB
}

func NewMessageRepository(
	db *gorm.DB,
) *MessageRepository {
	return &MessageRepository{
		db: db,
	}
}

func (r *MessageRepository) Create(
	message *model.Message,
) error {
	return r.db.Create(message).Error
}

func (r *MessageRepository) CreateWithOutbox(
	message *model.Message,
) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(message).Error; err != nil {
			return err
		}

		e := event.MessageCreated{
			EventID:    "message.created:" + strconv.FormatUint(uint64(message.ID), 10),
			Type:       event.TypeMessageCreated,
			Version:    event.MessageCreatedV1,
			OccurredAt: time.Now().UTC(),

			MessageID:       message.ID,
			ConversationID:  message.ConversationID,
			SenderID:        message.SenderID,
			ClientMessageID: message.ClientMessageID,
			Content:         message.Content,
			CreatedAt:       message.CreatedAt,
		}
		payload, err := json.Marshal(e)

		if err != nil {
			return fmt.Errorf(
				"marshal message event: %w", err,
			)
		}

		outbox := model.OutBoxEvent{
			EventID:       e.EventID,
			AggregateType: "conversation",
			AggregateID:   uint64(message.ConversationID),
			EventType:     e.Type,
			Payload:       string(payload),
			Status:        "pending",
		}

		if err := tx.Create(&outbox).Error; err != nil {
			return fmt.Errorf(
				"insert outbox event: %w", err,
			)
		}

		return nil

	})
}

func (r *MessageRepository) FindByID(
	messageID uint,
) (*model.Message, error) {
	var message model.Message

	err := r.db.
		First(&message, messageID).
		Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &message, nil
}

func (r *MessageRepository) FindByClientMessageID(
	senderID uint,
	clientMessageID string,
) (*model.Message, error) {
	var message model.Message

	err := r.db.
		Where(
			"sender_id = ? AND client_message_id = ?",
			senderID,
			clientMessageID,
		).
		First(&message).
		Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &message, nil

}

func (r *MessageRepository) ListBefore(
	conversationID uint,
	beforeID uint,
	limit int,
) ([]model.Message, bool, error) {
	var messages []model.Message

	query := r.db.
		Where(
			"conversation_id = ?",
			conversationID,
		)

	if beforeID > 0 {
		query = query.Where(
			"id < ?",
			beforeID,
		)
	}

	err := query.
		Order("id DESC").
		Limit(limit + 1).
		Find(&messages).
		Error

	if err != nil {
		return nil, false, err
	}

	hasMore := len(messages) > limit

	if hasMore {
		messages = messages[:limit]
	}

	reverseMessageSlice(messages)

	return messages, hasMore, nil

}

func (r *MessageRepository) ListAfter(
	conversationID uint,
	afterID uint,
	limit int,
) ([]model.Message, bool, error) {
	var messages []model.Message

	err := r.db.
		Where(
			"conversation_id = ? AND id > ?",
			conversationID,
			afterID,
		).
		Order("id ASC").
		Limit(limit + 1).
		Find(&messages).
		Error

	if err != nil {
		return nil, false, err
	}

	hasMore := len(messages) > limit

	if hasMore {
		messages = messages[:limit]
	}

	return messages, hasMore, nil

}

func (r *MessageRepository) CountUnreadAfter(
	conversationID uint,
	messageID uint,
	userID uint,
) (int64, error) {
	var count int64

	err := r.db.
		Model(&model.Message{}).
		Where(
			"conversation_id = ? AND id > ? AND sender_id <> ?",
			conversationID,
			messageID,
			userID,
		).
		Count(&count).
		Error

	return count, err
}

func reverseMessageSlice(
	messages []model.Message,
) {
	for left, right :=
		0, len(messages)-1; left < right; left, right =
		left+1, right-1 {

		messages[left], messages[right] =
			messages[right], messages[left]
	}
}
