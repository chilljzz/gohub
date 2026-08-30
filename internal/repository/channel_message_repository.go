package repository

import (
	"errors"

	"github.com/chilljzz/gohub/internal/database"
	"github.com/chilljzz/gohub/internal/model"
	"gorm.io/gorm"
)

type ChannelMessageRepository struct{}

func NewChannelMessageRepository() *ChannelMessageRepository {
	return &ChannelMessageRepository{}
}

func (r *ChannelMessageRepository) Create(
	message *model.ChannelMessage,
) error {
	return database.DB.Create(message).Error
}

func (r *ChannelMessageRepository) ListRecentByChannelID(
	channelID uint,
	limit int,
) ([]model.ChannelMessage, error) {
	var messages []model.ChannelMessage
	err := database.DB.
		Where("channel_id = ?", channelID).
		Order("id DESC").
		Limit(limit).
		Find(&messages).Error

	if err != nil {
		return nil, err
	}

	reverseMessages(messages)
	return messages, nil
}

func (r *ChannelMessageRepository) ListBefore(
	channelID uint,
	beforeID uint,
	limit int,
) ([]model.ChannelMessage, bool, error) {
	var messages []model.ChannelMessage

	query := database.DB.
		Where("channel_id = ?", channelID)

	if beforeID > 0 {
		query = query.Where("id < ?", beforeID)
	}

	err := query.Order("id DESC").
		Limit(limit + 1).
		Find(&messages).Error

	if err != nil {
		return nil, false, err
	}

	hasMore := len(messages) > limit
	if hasMore {
		messages = messages[:limit]
	}
	reverseMessages(messages)
	return messages, hasMore, nil
}

func (r *ChannelMessageRepository) ListAfter(
	channelID uint,
	afterID uint,
	limit int,
) ([]model.ChannelMessage, bool, error) {
	var messages []model.ChannelMessage

	err := database.DB.
		Where(
			"channel_id = ? AND id > ?",
			channelID,
			afterID,
		).
		Order("id ASC").
		Limit(limit + 1).
		Find(&messages).Error

	if err != nil {
		return nil, false, err
	}

	hasMore := len(messages) > limit

	if hasMore {
		messages = messages[:limit]
	}
	return messages, hasMore, nil
}

func (r *ChannelMessageRepository) CountAfter(
	channelID uint,
	messageID uint,
) (int64, error) {
	var count int64

	err := database.DB.
		Model(&model.ChannelMessage{}).
		Where(
			"channel_id = ? AND id > ?",
			channelID,
			messageID,
		).
		Count(&count).Error

	return count, err
}

func (r *ChannelMessageRepository) FindByID(
	messageID uint,
) (*model.ChannelMessage, error) {
	var message model.ChannelMessage

	err := database.DB.
		First(&message, messageID).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &message, nil
}

func reverseMessages(messages []model.ChannelMessage) {
	for left, right := 0, len(messages)-1; left < right; left, right = left+1, right-1 {

		messages[left], messages[right] =
			messages[right], messages[left]
	}
}
