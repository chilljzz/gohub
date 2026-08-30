package service

import (
	"errors"

	"github.com/chilljzz/gohub/internal/dto"
	"github.com/chilljzz/gohub/internal/model"
	"github.com/chilljzz/gohub/internal/repository"
)

type ChannelReadService struct {
	readRepo        *repository.ChannelReadRepository
	messageRepo     *repository.ChannelMessageRepository
	realtimeService *RealtimeService
}

func NewChannelReadService(
	readRepo *repository.ChannelReadRepository,
	messageRepo *repository.ChannelMessageRepository,
	realtimeService *RealtimeService,
) *ChannelReadService {
	return &ChannelReadService{
		readRepo:        readRepo,
		messageRepo:     messageRepo,
		realtimeService: realtimeService,
	}
}

var (
	ErrMessageNotFound = errors.New("message not found")

	ErrMessageNotInChannel = errors.New(
		"message does not belong to this channel",
	)
)

func (s *ChannelReadService) MarkRead(
	userID uint,
	channelID uint,
	messageID uint,
) error {
	err := s.realtimeService.CheckChannelAccess(
		userID,
		channelID,
	)
	if err != nil {
		return nil
	}
	message, err := s.messageRepo.FindByID(messageID)
	if err != nil {
		return err
	}
	if message == nil {
		return ErrMessageNotFound
	}
	if message.ChannelID != channelID {
		return ErrMessageNotInChannel
	}
	read, err := s.readRepo.Find(channelID, userID)
	if err != nil {
		return err
	}
	if read == nil {
		return s.readRepo.Create(
			&model.ChannelRead{
				ChannelID:         channelID,
				UserID:            userID,
				LastReadMessageID: messageID,
			},
		)
	}

	if messageID <= read.LastReadMessageID {
		return nil
	}

	return s.readRepo.UpdateLastRead(
		read.ID,
		messageID,
	)
}

func (s *ChannelReadService) GetUnreadCount(
	userID uint,
	channelID uint,
) (*dto.ChannelUnreadResult, error) {
	err := s.realtimeService.CheckChannelAccess(
		userID,
		channelID,
	)
	if err != nil {
		return nil, err
	}
	read, err := s.readRepo.Find(channelID, userID)
	if err != nil {
		return nil, err
	}
	lastReadMessageID := uint(0)

	if read != nil {
		lastReadMessageID = read.LastReadMessageID
	}

	count, err := s.messageRepo.CountAfter(
		channelID,
		lastReadMessageID,
	)
	if err != nil {
		return nil, err
	}
	return &dto.ChannelUnreadResult{
		ChannelID:         channelID,
		LastReadMessageID: lastReadMessageID,
		UnreadCount:       count,
	}, nil
}
