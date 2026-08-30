package service

import (
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/chilljzz/gohub/internal/dto"
	"github.com/chilljzz/gohub/internal/model"
	"github.com/chilljzz/gohub/internal/repository"
)

var (
	ErrMessageContentRequired = errors.New(
		"message content is required",
	)

	ErrMessageTooLong = errors.New(
		"message is too long",
	)
)

type ChannelMessageService struct {
	messageRepo     *repository.ChannelMessageRepository
	realtimeService *RealtimeService
}

func NewChannelMessageService(
	messageRepo *repository.ChannelMessageRepository,
	realtimeService *RealtimeService,
) *ChannelMessageService {
	return &ChannelMessageService{
		messageRepo:     messageRepo,
		realtimeService: realtimeService,
	}
}

func (s *ChannelMessageService) CreateChannelMessage(
	userID uint,
	channelID uint,
	content string,
) (*dto.ChannelMessageResult, error) {
	content = strings.TrimSpace(content)

	if content == "" {
		return nil, ErrMessageContentRequired
	}
	if utf8.RuneCountInString(content) > 1000 {
		return nil, ErrMessageTooLong
	}

	err := s.realtimeService.CheckChannelAccess(
		userID,
		channelID,
	)
	if err != nil {
		return nil, err
	}

	message := &model.ChannelMessage{
		ChannelID: channelID,
		SenderID:  userID,
		Content:   content,
	}

	if err := s.messageRepo.Create(message); err != nil {
		return nil, err
	}

	result := &dto.ChannelMessageResult{
		ID:        message.ID,
		ChannelID: message.ChannelID,
		SenderID:  message.SenderID,
		Content:   message.Content,
		CreatedAt: message.CreatedAt,
	}

	return result, nil
}

func (s *ChannelMessageService) ListMessageBefore(
	userID uint,
	channelID uint,
	beforeID uint,
	limit int,
) (*dto.ChannelMessagePageResult, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	err := s.realtimeService.CheckChannelAccess(userID, channelID)
	if err != nil {
		return nil, err
	}

	messages, hasMore, err := s.messageRepo.ListBefore(
		channelID,
		beforeID,
		limit,
	)
	if err != nil {
		return nil, err
	}
	results := toChannelMessageResults(messages)

	nextBeforeID := uint(0)
	if hasMore && len(results) > 0 {
		nextBeforeID = results[0].ID
	}

	return &dto.ChannelMessagePageResult{
		Messages:     results,
		NextBeforeID: nextBeforeID,
		HasMore:      hasMore,
	}, nil

}

func (s *ChannelMessageService) SyncMessagesAfter(
	userID uint,
	channelID uint,
	afterID uint,
	limit int,
) (*dto.ChannelMessageSyncResult, error) {
	if limit <= 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}
	err := s.realtimeService.CheckChannelAccess(
		userID,
		channelID,
	)
	if err != nil {
		return nil, err
	}
	messages, hasMore, err := s.messageRepo.ListAfter(channelID, afterID, limit)
	if err != nil {
		return nil, err
	}

	results := toChannelMessageResults(messages)

	nextAfterID := afterID

	if len(results) > 0 {
		nextAfterID = results[len(results)-1].ID
	}

	return &dto.ChannelMessageSyncResult{
		Messages:    results,
		NextAfterID: nextAfterID,
		HasMore:     hasMore,
	}, nil

}

func toChannelMessageResults(
	messages []model.ChannelMessage,
) []dto.ChannelMessageResult {
	results := make(
		[]dto.ChannelMessageResult,
		0,
		len(messages),
	)
	for _, message := range messages {
		results = append(results,
			dto.ChannelMessageResult{
				ID:        message.ID,
				ChannelID: message.ChannelID,
				SenderID:  message.SenderID,
				Content:   message.Content,
				CreatedAt: message.CreatedAt,
			})
	}
	return results
}
