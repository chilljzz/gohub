package service

import (
	"errors"
	"fmt"

	"github.com/chilljzz/gohub/internal/model"
	"gorm.io/gorm"
)

type ConversationRepository interface {
	FindByChannelID(
		channelID uint,
	) (*model.Conversation, error)

	Create(
		conversation *model.Conversation,
	) error

	FindByDirectKey(
		directKey string,
	) (*model.Conversation, error)

	CreateDirectWithMembers(
		conversation *model.Conversation,
		userID uint,
		otherUserID uint,
	) error

	ListMembers(
		conversationID uint,
	) ([]model.ConversationMember, error)
}

type FriendshipChecker interface {
	IsFriend(
		userID uint,
		friendID uint,
	) (bool, error)
}

type ConversationService struct {
	conversationRepo ConversationRepository
	friendChecker    FriendshipChecker
}

var (
	ErrConversationNotFound = errors.New("conversation not found")
)

func NewConversationService(
	conversationRepo ConversationRepository,
	friendChecker FriendshipChecker,
) *ConversationService {
	return &ConversationService{
		conversationRepo: conversationRepo,
		friendChecker:    friendChecker,
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
