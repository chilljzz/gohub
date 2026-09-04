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

func (r *ConversationRepository) FindByDirectKey(
	directKey string,
) (*model.Conversation, error) {
	var conversation model.Conversation

	err := r.db.
		Where(
			"type = ? AND direct_key = ?",
			model.ConversationTypeDirect,
			directKey,
		).First(&conversation).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &conversation, nil
}

func (r *ConversationRepository) ListMember(
	conversationID uint,
) ([]model.ConversationMember, error) {
	var members []model.ConversationMember

	err := r.db.
		Where(
			"conversation_id = ?",
			conversationID,
		).
		Find(&members).Error

	if err != nil {
		return nil, err
	}

	return members, nil
}

func (r *ConversationRepository) CreateDirectWithMembers(
	conversation *model.Conversation,
	userID uint,
	otherUserID uint,
) error {
	return r.db.Transaction(
		func(tx *gorm.DB) error {
			if err := tx.Create(conversation).Error; err != nil {
				return err
			}

			members := []model.ConversationMember{
				{
					ConversationID: conversation.ID,
					UserID:         userID,
				},
				{
					ConversationID: conversation.ID,
					UserID:         otherUserID,
				},
			}
			if err := tx.Create(&members).Error; err != nil {
				return err
			}

			return nil
		},
	)
}
