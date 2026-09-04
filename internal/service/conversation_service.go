package service

import (
	"errors"
	"fmt"

	"github.com/chilljzz/gohub/internal/model"
	"github.com/chilljzz/gohub/internal/repository"
	"gorm.io/gorm"
)

type ConversationService struct {
	conversationRepo *repository.ConversationRepository
}

var (
	ErrConversationNotFound = errors.New("conversation not found")
)

func NewConversationService(
	conversationRepo *repository.ConversationRepository,
) *ConversationService {
	return &ConversationService{
		conversationRepo: conversationRepo,
	}
}

func (s *ConversationService) GetChannelConversation(
	channelID uint,
) (*model.Conversation, error) {
	conversation, err := s.conversationRepo.FindByChannelID(channelID)

	if err != nil {
		return nil, err
	}
	if conversation == nil {
		return nil, ErrConversationNotFound
	}

	return conversation, nil

}

func (s *ConversationService) EnsureChannelConversation(
	channelID uint,
) (*model.Conversation, error) {
	existing, err := s.conversationRepo.FindByChannelID(channelID)
	if err != nil {
		return nil, err
	}

	if existing != nil {
		return existing, nil
	}

	conversation := &model.Conversation{
		Type:      model.ConversationTypeChannel,
		ChannelID: &channelID,
	}

	err = s.conversationRepo.Create(conversation)

	if err == nil {
		return conversation, nil
	}

	if !errors.Is(err, gorm.ErrDuplicatedKey) {
		return nil, err
	}

	return s.conversationRepo.FindByChannelID(channelID)

}

func buildDirectKey(userID uint, otherUserID uint) string {
	if userID < otherUserID {
		return fmt.Sprintf(
			"%d:%d",
			userID,
			otherUserID,
		)
	}

	return fmt.Sprintf(
		"%d:%d",
		otherUserID,
		userID,
	)
}
