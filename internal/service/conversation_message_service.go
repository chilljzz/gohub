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

type ConversationMessageService struct {
	messageRepo         *repository.MessageRepository
	conversationService *ConversationService
}

func NewConversationMessageService(
	messageRepo *repository.MessageRepository,
	conversationService *ConversationService,
) *ConversationMessageService {
	return &ConversationMessageService{
		messageRepo: messageRepo,

		conversationService: conversationService,
	}
}

type CreateConversationMessageResult struct {
	Message *dto.ConversationMessageResult

	Created bool
}

func (s *ConversationMessageService) CreateMessage(
	userID uint,
	conversationID uint,
	content string,
	clientMessageID string,
) (*CreateConversationMessageResult, error) {
	content = strings.TrimSpace(content)

	clientMessageID = strings.TrimSpace(clientMessageID)

	if clientMessageID == "" || len(clientMessageID) > 64 {
		return nil, ErrInvalidClientMessageID
	}

	if content == "" {
		return nil, ErrMessageContentRequired
	}

	if utf8.RuneCountInString(content) > 1000 {
		return nil, ErrMessageTooLong
	}

	_, err := s.conversationService.GetAccessibleConversation(userID, conversationID)

	if err != nil {
		return nil, err
	}

	message := &model.Message{
		ConversationID: conversationID,

		SenderID: userID,

		ClientMessageID: clientMessageID,

		Content: content,
	}

	err = s.messageRepo.Create(
		message,
	)

	if err == nil {
		return &CreateConversationMessageResult{
			Message: toConversationMessageResult(message),
			Created: true,
		}, nil
	}

	if !errors.Is(
		err,
		gorm.ErrDuplicatedKey,
	) {
		return nil, err
	}

	existing, err := s.messageRepo.FindByClientMessageID(userID, clientMessageID)

	if err != nil {
		return nil, err
	}

	if existing == nil {
		return nil, ErrMessageCreateFailed
	}

	if existing.ConversationID !=
		conversationID ||
		existing.Content != content {

		return nil,
			ErrClientMessageConflict
	}

	return &CreateConversationMessageResult{
		Message: toConversationMessageResult(
			existing,
		),

		Created: false,
	}, nil

}

func (s *ConversationMessageService) ListBefore(
	userID uint,
	conversationID uint,
	beforeID uint,
	limit int,
) (*dto.ConversationMessagePageResult, error) {
	if limit <= 0 {
		limit = 20
	}

	if limit > 100 {
		limit = 100
	}

	_, err := s.conversationService.GetAccessibleConversation(userID, conversationID)

	if err != nil {
		return nil, err
	}

	messages, hasMore, err := s.messageRepo.ListBefore(conversationID, beforeID, limit)

	if err != nil {
		return nil, err
	}

	results := toConversationMessageResults(messages)

	nextBeforeID := uint(0)

	if hasMore && len(results) > 0 {
		nextBeforeID = results[0].ID
	}

	return &dto.ConversationMessagePageResult{
		Messages:     results,
		NextBeforeID: nextBeforeID,
		HasMore:      hasMore,
	}, nil

}

func (s *ConversationMessageService) SyncAfter(
	userID uint,
	conversationID uint,
	afterID uint,
	limit int,
) (*dto.ConversationMEssageSyncResult, error) {
	if limit <= 0 {
		limit = 100
	}

	if limit > 500 {
		limit = 500
	}

	_, err := s.conversationService.GetAccessibleConversation(
		userID,
		conversationID,
	)

	if err != nil {
		return nil, err
	}

	messages, hasMore, err := s.messageRepo.ListAfter(
		conversationID,
		afterID,
		limit,
	)
	if err != nil {
		return nil, err
	}

	results := toConversationMessageResults(messages)

	nextAfterID := afterID

	if len(results) > 0 {
		nextAfterID = results[len(results)-1].ID
	}

	return &dto.ConversationMEssageSyncResult{
		Messages:    results,
		NextAfterID: nextAfterID,
		HasMore:     hasMore,
	}, nil

}

func toConversationMessageResult(
	message *model.Message,
) *dto.ConversationMessageResult {

	return &dto.ConversationMessageResult{
		ID: message.ID,

		ConversationID: message.ConversationID,

		SenderID: message.SenderID,

		ClientMessageID: message.ClientMessageID,

		Content: message.Content,

		CreatedAt: message.CreatedAt,
	}
}

func toConversationMessageResults(
	messages []model.Message,
) []dto.ConversationMessageResult {

	results :=
		make(
			[]dto.ConversationMessageResult,
			0,
			len(messages),
		)

	for i := range messages {

		result :=
			toConversationMessageResult(
				&messages[i],
			)

		results =
			append(
				results,
				*result,
			)
	}

	return results
}
