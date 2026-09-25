package service

import (
	"errors"
	"time"

	"github.com/chilljzz/gohub/internal/dto"
	"github.com/chilljzz/gohub/internal/model"
	"github.com/chilljzz/gohub/internal/repository"
)

var (
	ErrMessageNotInConversation = errors.New(
		"message does not belong to conversation",
	)
)

type DirectReceiptService struct {
	readRepo *repository.ConversationReadRepository

	messageRepo *repository.MessageRepository

	conversationService *ConversationService
}

func NewDirectReceiptService(
	readRepo *repository.ConversationReadRepository,

	messageRepo *repository.MessageRepository,

	conversationService *ConversationService,
) *DirectReceiptService {

	return &DirectReceiptService{
		readRepo: readRepo,

		messageRepo: messageRepo,

		conversationService: conversationService,
	}
}

type DirectReceiptResult struct {
	ConversationID uint

	UserID uint

	MessageID uint

	At time.Time
}

func (s *DirectReceiptService) validateTarget(
	userID uint,
	conversationID uint,
	messageID uint,
) (*model.Conversation, *model.Message, error) {
	if conversationID == 0 || messageID == 0 {
		return nil, nil, ErrInvalidParam
	}

	conversation, err := s.conversationService.GetAccessibleConversation(userID, conversationID)

	if err != nil {
		return nil, nil, err
	}

	if conversation.Type != model.ConversationTypeDirect {
		return nil, nil, ErrInvalidConversationType
	}

	message, err := s.messageRepo.FindByID(messageID)

	if err != nil {
		return nil, nil, err
	}

	if message == nil {
		return nil, nil, ErrMessageNotFound
	}

	if message.ConversationID != conversationID {
		return nil, nil, ErrMessageNotInChannel
	}

	return conversation, message, nil

}

func (s *DirectReceiptService) MarkDelivered(
	userID uint,
	conversationID uint,
	messageID uint,
) (*DirectReceiptResult, error) {
	_, _, err := s.validateTarget(userID, conversationID, messageID)

	if err != nil {
		return nil, err
	}

	if err := s.readRepo.UpsertDelivered(conversationID, userID, messageID); err != nil {
		return nil, err
	}

	return &DirectReceiptResult{
		ConversationID: conversationID,
		UserID:         userID,
		MessageID:      messageID,
		At:             time.Now().UTC(),
	}, nil
}

func (s *DirectReceiptService) MarkRead(
	userID uint,
	conversationID uint,
	messageID uint,
) (*DirectReceiptResult, error) {
	_, _, err := s.validateTarget(userID, conversationID, messageID)

	if err != nil {
		return nil, err
	}

	if err := s.readRepo.UpsertRead(conversationID, userID, messageID); err != nil {
		return nil, err
	}

	return &DirectReceiptResult{
		ConversationID: conversationID,
		UserID:         userID,
		MessageID:      messageID,
		At:             time.Now().UTC(),
	}, nil

}

func (s *DirectReceiptService) GetPeerState(
	userID uint,
	conversationID uint,
) (*dto.DirectReceiptStatrResult, error) {
	peerID, err := s.conversationService.GetDirectPeeID(userID, conversationID)
	if err != nil {
		return nil, err
	}

	read, err := s.readRepo.Find(conversationID, peerID)
	if err != nil {
		return nil, err
	}

	result :=
		&dto.DirectReceiptStatrResult{
			ConversationID: conversationID,

			PeerUserID: peerID,
		}

	if read != nil {
		result.LastDeliveredMessageID = read.LastDeliveredMessageID
		result.LastReadMessageID = read.LastReadMessageID
	}

	return result, nil

}
