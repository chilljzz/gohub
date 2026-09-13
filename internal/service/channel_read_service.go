package service

import (
	"errors"

	"github.com/chilljzz/gohub/internal/dto"
	"github.com/chilljzz/gohub/internal/repository"
)

type ChannelReadService struct {
	readRepo            *repository.ConversationReadRepository
	messageRepo         *repository.MessageRepository
	realtimeService     *RealtimeService
	conversationService *ConversationService
}

func NewChannelReadService(
	readRepo *repository.ConversationReadRepository,
	messageRepo *repository.MessageRepository,
	realtimeService *RealtimeService,
	conversationService *ConversationService,
) *ChannelReadService {
	return &ChannelReadService{
		readRepo:            readRepo,
		messageRepo:         messageRepo,
		realtimeService:     realtimeService,
		conversationService: conversationService,
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
		return err
	}

	conversation, err := s.conversationService.GetChannelConversation(channelID)
	if err != nil {
		return err
	}

	message, err := s.messageRepo.FindByID(messageID)
	if err != nil {
		return err
	}
	if message == nil {
		return ErrMessageNotFound
	}

	if message.ConversationID != conversation.ID {
		return ErrMessageNotInChannel
	}

	return s.readRepo.UpsertLastRead(
		conversation.ID,
		userID,
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

	conversation, err := s.conversationService.GetChannelConversation(channelID)
	if err != nil {
		return nil, err
	}

	read, err := s.readRepo.Find(conversation.ID, userID)
	if err != nil {
		return nil, err
	}

	lastReadMessageID := uint(0)

	if read != nil {
		lastReadMessageID = read.LastReadMessageID
	}

	count, err := s.messageRepo.CountAfter(
		conversation.ID,
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
