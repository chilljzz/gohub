package repository

import (
	"errors"

	"github.com/chilljzz/gohub/internal/model"
	"gorm.io/gorm"
)

type ConversationReadRepository struct {
	db *gorm.DB
}

func NewConversationReadRepository(
	db *gorm.DB,
) *ConversationReadRepository {
	return &ConversationReadRepository{
		db: db,
	}
}

func (r *ConversationReadRepository) Find(
	conversationID uint,
	userID uint,
) (*model.ConversationRead, error) {
	var read model.ConversationRead

	err := r.db.
		Where(
			"conversation_id = ? AND user_id = ?",
			conversationID,
			userID,
		).
		First(&read).
		Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if err != nil {
		return nil, nil
	}

	return &read, nil
}

func (r *ConversationReadRepository) UpsertLastRead(
	conversationID uint,
	userID uint,
	messageID uint,
) error {
	const query = `
		INSERT INTO conversation_reads (
			conversation_id,
			user_id,
			last_read_message_id,
			created_at,
			updated_at
		)
		VALUES(?,?,?,NOW(),NOW())
		ON DUPLICATE KEY UPDATE
			last_read_message_id = 
				GREATEST(last_read_message_id,?),
			updated_at = NOW()
	`

	return r.db.Exec(query, conversationID, userID, messageID, messageID).Error
}
