package service

import (
	"errors"
	"fmt"

	"github.com/chilljzz/gohub/internal/dto"
	"github.com/chilljzz/gohub/internal/model"
	"gorm.io/gorm"
)

type ConversationRepository interface {
	FindByID(id uint) (*model.Conversation, error)

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

	IsMember(
		conversationID uint,
		userID uint,
	) (bool, error)

	// ListMembers(
	// 	conversationID uint,
	// ) ([]model.ConversationMember, error)
}

type FriendshipChecker interface {
	IsFriend(
		userID uint,
		friendID uint,
	) (bool, error)
}

type ChannelAccessChecker interface {
	CheckChannelAccess(
		userID uint,
		channelID uint,
	) error
}

type ConversationService struct {
	conversationRepo     ConversationRepository
	friendChecker        FriendshipChecker
	channelAccessChecker ChannelAccessChecker
}

func NewConversationService(
	conversationRepo ConversationRepository,
	friendChecker FriendshipChecker,
	channelAccessChecker ChannelAccessChecker,
) *ConversationService {
	return &ConversationService{
		conversationRepo:     conversationRepo,
		friendChecker:        friendChecker,
		channelAccessChecker: channelAccessChecker,
	}
}

var (
	ErrConversationNotFound = errors.New("conversation not found")

	ErrCannotChatWithSelf = errors.New("cannot chat with self")

	ErrConversationAccessDenied = errors.New("conversation access denied")

	ErrDirectChatRequiresFriend = errors.New("direct chat requires friendship")

	ErrConversationCreateFailed = errors.New("conversation create failed")

	ErrInvalidConversationType = errors.New("invalid conversation type")
)

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

func (s *ConversationService) GetOrCreateDirectConversation(
	userID uint,
	otherUserID uint,
) (*dto.DirectConversationResult, error) {
	if userID == otherUserID {
		return nil, ErrCannotChatWithSelf
	}
	isFriend, err := s.friendChecker.IsFriend(userID, otherUserID)

	if err != nil {
		return nil, err
	}

	if !isFriend {
		return nil, ErrDirectChatRequiresFriend
	}

	directKey := buildDirectKey(userID, otherUserID)

	existing, err := s.conversationRepo.FindByDirectKey(directKey)

	if err != nil {
		return nil, err
	}
	if existing != nil {
		return toDirectConversationResult(
			existing,
			otherUserID,
		), nil
	}
	conversation :=
		&model.Conversation{
			Type: model.ConversationTypeDirect,

			DirectKey: &directKey,
		}

	err =
		s.conversationRepo.CreateDirectWithMembers(
			conversation,
			userID,
			otherUserID,
		)

	if err == nil {
		return toDirectConversationResult(
			conversation,
			otherUserID,
		), nil
	}

	if !errors.Is(err, gorm.ErrDuplicatedKey) {
		return nil, err
	}

	existing, err = s.conversationRepo.FindByDirectKey(directKey)

	if err != nil {
		return nil, err
	}

	if existing == nil {
		return nil, ErrConversationCreateFailed
	}

	return toDirectConversationResult(
		existing,
		otherUserID,
	), nil

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

func toDirectConversationResult(
	conversation *model.Conversation,
	otherUserID uint,
) *dto.DirectConversationResult {
	return &dto.DirectConversationResult{
		ConversationID: conversation.ID,
		Type:           string(conversation.Type),
		OtherUserID:    otherUserID,
		CreatedAt:      conversation.CreatedAt,
	}
}

func (s *ConversationService) GetAccessibleConversation(
	userID uint,
	conversationID uint,
) (*model.Conversation, error) {
	conversation, err := s.conversationRepo.FindByID(conversationID)
	if err != nil {
		return nil, err
	}

	if conversation == nil {
		return nil, ErrConversationNotFound
	}

	switch conversation.Type {
	case model.ConversationTypeDirect:
		isMember, err := s.conversationRepo.IsMember(
			conversationID,
			userID,
		)

		if err != nil {
			return nil, err
		}
		if !isMember {
			return nil, ErrConversationAccessDenied
		}

	case model.ConversationTypeChannel:
		if conversation.ChannelID == nil {
			return nil, ErrInvalidConversationType
		}

		if err := s.channelAccessChecker.CheckChannelAccess(
			userID,
			*conversation.ChannelID,
		); err != nil {
			return nil, err
		}

	default:
		return nil, ErrInvalidConversationType
	}

	return conversation, nil
}
