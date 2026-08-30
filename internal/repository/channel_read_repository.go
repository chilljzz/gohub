package repository

import (
	"errors"

	"github.com/chilljzz/gohub/internal/database"
	"github.com/chilljzz/gohub/internal/model"
	"gorm.io/gorm"
)

type ChannelReadRepository struct{}

func NewChannelReadRepository() *ChannelReadRepository {
	return &ChannelReadRepository{}
}

func (r *ChannelReadRepository) Find(
	channelID uint,
	userID uint,
) (*model.ChannelRead, error) {

	var read model.ChannelRead

	err := database.DB.
		Where(
			"channel_id = ? AND user_id = ?",
			channelID,
			userID,
		).
		First(&read).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &read, nil
}

func (r *ChannelReadRepository) Create(
	read *model.ChannelRead,
) error {
	return database.DB.Create(read).Error
}

func (r *ChannelReadRepository) UpdateLastRead(
	readID uint,
	messageID uint,
) error {
	result := database.DB.
		Model(&model.ChannelRead{}).
		Where("id = ?", readID).
		Update(
			"last_read_message_id",
			messageID,
		)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("channel read not found")
	}
	return nil
}
