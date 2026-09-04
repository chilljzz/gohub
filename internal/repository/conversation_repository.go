package repository

import (
	"errors"

	"github.com/chilljzz/gohub/internal/model"
	"gorm.io/gorm"
)

type ConversationRepository struct {
	db *gorm.DB
}

func NewConversationRepository(
	db *gorm.DB,
) *ConversationRepository {
	return &ConversationRepository{
		db: db,
	}
}

func (r *ConversationRepository) Create(
	conversation *model.Conversation,
) error {
	return r.db.Create(conversation).Error
}

func (r *ConversationRepository) FindByID(
	id uint,
) (*model.Conversation, error) {
	var conversation model.Conversation

	err := r.db.First(&conversation, id).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &conversation, nil
}

func (r *ConversationRepository) FindByChannelID(
	channelID uint,
) (*model.Conversation, error) {
	var conversation model.Conversation

	err := r.db.
		Where(
			"type = ? AND channel_id = ?",
			model.ConversationTypeChannel,
			channelID,
		).
		First(
			&conversation,
		).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &conversation, nil
}

func (r *ConversationRepository) AddMember(
	member *model.ConversationMember,
) error {
	return r.db.Create(member).Error
}

func (r *ConversationRepository) IsMember(
	conversationID uint,
	userID uint,
) (bool, error) {
	var count int64

	err := r.db.
		Model(
			&model.ConversationMember{},
		).
		Where(
			"conversation_id = ? AND user_id = ?",
			conversationID,
			userID,
		).
		Count(
			&count,
		).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}
