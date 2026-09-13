package service

import (
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/chilljzz/gohub/internal/dto"
	"github.com/chilljzz/gohub/internal/model"
	"github.com/chilljzz/gohub/internal/repository"
	"gorm.io/gorm"
)

var (
	ErrMessageContentRequired = errors.New("message content is required")

	ErrMessageTooLong = errors.New("message is too long")

	ErrMessageCreateFailed = errors.New("message create failed")

	ErrInvalidClientMessageID = errors.New("invalid client message id")

	ErrClientMessageConflict = errors.New("client message id conflict")
)

type ChannelMessageService struct {
	messageRepo         *repository.MessageRepository
	realtimeService     *RealtimeService
	conversationService *ConversationService
}

type CreateMessageResult struct {
	Message *dto.ChannelMessageResult
	Created bool
}

func NewChannelMessageService(
	messageRepo *repository.MessageRepository,
	realtimeService *RealtimeService,
	conversationService *ConversationService,
) *ChannelMessageService {
	return &ChannelMessageService{
		messageRepo:         messageRepo,
		realtimeService:     realtimeService,
		conversationService: conversationService,
	}
}

func (s *ChannelMessageService) CreateChannelMessage(
	userID uint,
	channelID uint,
	content string,
	clientMessageID string,
) (*CreateMessageResult, error) {
	content = strings.TrimSpace(content)
	clientMessageID = strings.TrimSpace(clientMessageID)

	if clientMessageID == "" ||
		len(clientMessageID) > 64 {

		return nil, ErrInvalidClientMessageID
	}

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

	conversation, err := s.conversationService.GetChannelConversation(channelID)

	if err != nil {
		return nil, err
	}

	message := &model.Message{
		ConversationID:  conversation.ID,
		SenderID:        userID,
		ClientMessageID: clientMessageID,
		Content:         content,
	}

	err = s.messageRepo.Create(message)

	if err == nil {
		result := toChannelMessageResult(
			message,
			channelID,
		)

		return &CreateMessageResult{
			Message: result,
			Created: true,
		}, nil
	}

	if !errors.Is(err, gorm.ErrDuplicatedKey) {
		return nil, err
	}

	existing, err := s.messageRepo.FindByClientMessageID(userID, clientMessageID)

	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrMessageCreateFailed
	}
	if existing.ConversationID != conversation.ID ||
		existing.Content != content {

		return nil, ErrClientMessageConflict
	}
	exist := toChannelMessageResult(
		existing,
		channelID,
	)
	return &CreateMessageResult{
		Message: exist,
		Created: false,
	}, nil

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

	conversation, err := s.conversationService.GetChannelConversation(channelID)

	if err != nil {
		return nil, err
	}

	messages, hasMore, err := s.messageRepo.ListBefore(
		conversation.ID,
		beforeID,
		limit,
	)
	if err != nil {
		return nil, err
	}
	results := toChannelMessageResults(messages, channelID)

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

	conversation, err := s.conversationService.GetChannelConversation(channelID)
	if err != nil {
		return nil, err
	}

	messages, hasMore, err := s.messageRepo.ListAfter(conversation.ID, afterID, limit)
	if err != nil {
		return nil, err
	}

	results := toChannelMessageResults(messages, channelID)

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
	messages []model.Message,
	channelID uint,
) []dto.ChannelMessageResult {
	results := make(
		[]dto.ChannelMessageResult,
		0,
		len(messages),
	)
	for _, message := range messages {
		results = append(
			results,
			dto.ChannelMessageResult{
				ID:              message.ID,
				ChannelID:       channelID,
				SenderID:        message.SenderID,
				ClientMessageID: message.ClientMessageID,
				Content:         message.Content,
				CreatedAt:       message.CreatedAt,
			},
		)
	}
	return results
}

func toChannelMessageResult(
	message *model.Message,
	channelID uint,
) *dto.ChannelMessageResult {
	return &dto.ChannelMessageResult{
		ID:              message.ID,
		ChannelID:       channelID,
		SenderID:        message.SenderID,
		ClientMessageID: message.ClientMessageID,
		Content:         message.Content,
		CreatedAt:       message.CreatedAt,
	}
}
