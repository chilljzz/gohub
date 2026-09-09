package repository

import (
	"errors"

	"github.com/chilljzz/gohub/internal/database"
	"github.com/chilljzz/gohub/internal/model"
	"gorm.io/gorm"
)

type ChannelRepository struct{}

func NewChannelRepository() *ChannelRepository {
	return &ChannelRepository{}
}

func (r *ChannelRepository) Create(
	channel *model.Channel,
) error {
	return database.DB.Create(channel).Error
}

func (r *ChannelRepository) ListByTeamID(
	teamID uint,
) ([]model.Channel, error) {
	var channels []model.Channel
	err := database.DB.
		Where("team_id = ?", teamID).
		Order("created_at ASC").
		Find(&channels).
		Error

	return channels, err
}

func (r *ChannelRepository) FindByTeamIDAndName(
	teamID uint,
	name string,
) (*model.Channel, error) {
	var channel model.Channel

	err := database.DB.
		Where(
			"team_id = ? AND name = ?",
			teamID,
			name,
		).
		First(&channel).
		Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &channel, nil
}

func (r *ChannelRepository) FindByID(
	channelID uint,
) (*model.Channel, error) {
	var channel model.Channel

	err := database.DB.
		First(&channel, channelID).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &channel, nil
}

func (r *ChannelRepository) CreateWithConversation(
	channel *model.Channel,
) error {
	return database.DB.Transaction(
		func(tx *gorm.DB) error {
			if err := tx.Create(channel).Error; err != nil {
				return err
			}
			conversation := &model.Conversation{
				Type:      model.ConversationTypeChannel,
				ChannelID: &channel.ID,
			}

			if err := tx.Create(conversation).Error; err != nil {
				return err
			}
			return nil
		},
	)
}
