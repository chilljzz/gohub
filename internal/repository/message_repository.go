package repository

import (
	"errors"

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

func (r *MessageRepository) CountAfter(
	conversationID uint,
	messageID uint,
) (int64, error) {
	var count int64

	err := r.db.
		Model(&model.Message{}).
		Where(
			"conversation_id = ? AND id > ?",
			conversationID,
			messageID,
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

func (r *MessageRepository) CreateChannelCompat(
	message *model.Message,
	channelID uint,
) error {
	return r.db.Transaction(
		func(tx *gorm.DB) error {
			legacyMessage := &model.ChannelMessage{
				ChannelID:       channelID,
				SenderID:        message.SenderID,
				ClientMessageID: message.ClientMessageID,
				Content:         message.Content,
			}
			if err := tx.Create(legacyMessage).Error; err != nil {
				return err
			}
			message.ID = legacyMessage.ID
			message.CreatedAt = legacyMessage.CreatedAt

			if err := tx.Create(message).Error; err != nil {
				return nil
			}

			return nil
		},
	)
}
