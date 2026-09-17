package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"

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

	ListForUser(
		userID uint,
		cursorLastMessageID uint,
		cursorConversationID uint,
		hasCursor bool,
		limit int,
	) (
		[]model.ConversationListProjection,
		bool,
		error,
	)
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

type PresenceReader interface {
	GetMany(
		ctx context.Context,
		userIDs []uint,
	) (map[uint]PresenceStatus, error)
}

type conversationListCursor struct {
	LastMessageID uint `json:"m"`

	ConversationID uint `json:"c"`
}

type ConversationService struct {
	conversationRepo     ConversationRepository
	friendChecker        FriendshipChecker
	channelAccessChecker ChannelAccessChecker

	presenceReader PresenceReader
}

func NewConversationService(
	conversationRepo ConversationRepository,
	friendChecker FriendshipChecker,
	channelAccessChecker ChannelAccessChecker,
	presenceReader PresenceReader,
) *ConversationService {
	return &ConversationService{
		conversationRepo:     conversationRepo,
		friendChecker:        friendChecker,
		channelAccessChecker: channelAccessChecker,
		presenceReader:       presenceReader,
	}
}

var (
	ErrConversationNotFound = errors.New("conversation not found")

	ErrCannotChatWithSelf = errors.New("cannot chat with self")

	ErrConversationAccessDenied = errors.New("conversation access denied")

	ErrDirectChatRequiresFriend = errors.New("direct chat requires friendship")

	ErrConversationCreateFailed = errors.New("conversation create failed")

	ErrInvalidConversationType = errors.New("invalid conversation type")

	ErrInvalidConversationCursor = errors.New("invalid conversation cursor")
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

func (s *ConversationService) ListConversations(
	ctx context.Context,
	userID uint,
	limit int,
	cursorText string,
) (*dto.ConversationListResult, error) {
	if limit <= 0 {
		limit = 20
	}

	if limit > 100 {
		limit = 100
	}

	var (
		cursorLastMessageID  uint
		cursorConversationID uint
		hasCursor            bool
	)

	if cursorText != "" {
		cursor, err := decodeConversationCursor(
			cursorText,
		)
		if err != nil {
			return nil, err
		}
		cursorLastMessageID = cursor.LastMessageID

		cursorConversationID = cursor.ConversationID

		hasCursor = true
	}

	rows, hasMore, err := s.conversationRepo.
		ListForUser(userID, cursorLastMessageID, cursorConversationID, hasCursor, limit)

	if err != nil {
		return nil, err
	}
	peerIDs := make([]uint, 0, len(rows))
	peerIDSet := make(map[uint]struct{})
	for i := range rows {
		row := &rows[i]

		if row.Type != model.ConversationTypeDirect {
			continue
		}
		if row.PeerUserID == nil {
			continue
		}

		peerID := *row.PeerUserID
		if _, exists := peerIDSet[peerID]; exists {
			continue
		}
		peerIDSet[peerID] = struct{}{}

		peerIDs = append(peerIDs, peerID)
	}
	presenceMap := make(map[uint]PresenceStatus)
	if s.presenceReader != nil && len(peerIDs) > 0 {
		statuses, err := s.presenceReader.GetMany(ctx, peerIDs)
		if err != nil {
			log.Printf("get presence failed: %v", err)
		} else {
			presenceMap = statuses
		}
	}

	items := make([]dto.ConversationListItem, 0, len(rows))

	for i := range rows {
		item, err := toConversationListItem(&rows[i])
		if err != nil {
			return nil, err
		}
		if item.Peer != nil {
			peerID := item.Peer.ID

			if presence, ok := presenceMap[peerID]; ok {
				item.Peer.Online = presence.Online
				item.Peer.LastSeenAt = presence.LastSeenAt
			}
		}

		items = append(items, item)
	}

	result := &dto.ConversationListResult{
		Items:   items,
		Hasmore: hasMore,
	}

	if hasMore && len(rows) > 0 {
		last := rows[len(rows)-1]
		lastMessageID := uint(0)

		if last.LastMessageID != nil {
			lastMessageID = *last.LastMessageID
		}
		nextCursor, err := encodeConversationCursor(
			lastMessageID,
			last.ConversationID,
		)
		if err != nil {
			return nil, err
		}

		result.NextCursor = nextCursor
	}

	return result, nil

}

func toConversationListItem(
	row *model.ConversationListProjection,
) (
	dto.ConversationListItem,
	error,
) {

	item :=
		dto.ConversationListItem{
			ID: row.ConversationID,

			Type: string(row.Type),

			UnreadCount: row.UnreadCount,
		}

	switch row.Type {

	case model.ConversationTypeDirect:

		if row.PeerUserID == nil {
			return item,
				ErrInvalidConversationType
		}

		item.Peer =
			&dto.ConversationPeerResult{
				ID: *row.PeerUserID,
			}

		if row.PeerUsername != nil {
			item.Peer.Username =
				*row.PeerUsername
		}

		if row.PeerNickname != nil {
			item.Peer.Nickname =
				*row.PeerNickname
		}

		if row.PeerAvatar != nil {
			item.Peer.Avatar =
				*row.PeerAvatar
		}

	case model.ConversationTypeChannel:

		if row.ChannelID == nil ||
			row.TeamID == nil {

			return item,
				ErrInvalidConversationType
		}

		item.Channel =
			&dto.ConversationChannelResult{
				ID: *row.ChannelID,

				TeamID: *row.TeamID,
			}

		if row.ChannelName != nil {
			item.Channel.Name =
				*row.ChannelName
		}

		if row.TeamName != nil {
			item.Channel.TeamName =
				*row.TeamName
		}

	default:

		return item,
			ErrInvalidConversationType
	}

	if row.LastMessageID != nil {

		lastMessage :=
			&dto.ConversationLastMessageResult{
				ID: *row.LastMessageID,
			}

		if row.LastSenderID != nil {
			lastMessage.SenderID =
				*row.LastSenderID
		}

		if row.LastMessageContent != nil {
			lastMessage.Content =
				*row.LastMessageContent
		}

		if row.LastMessageAt != nil {
			lastMessage.CreatedAt =
				*row.LastMessageAt
		}

		item.LastMessage =
			lastMessage
	}

	return item, nil
}

func encodeConversationCursor(
	lastMessageID uint,
	conversationID uint,
) (string, error) {
	cursor := conversationListCursor{
		LastMessageID:  lastMessageID,
		ConversationID: conversationID,
	}

	data, err := json.Marshal(cursor)

	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(data), nil
}

func decodeConversationCursor(
	value string,
) (conversationListCursor, error) {
	var cursor conversationListCursor

	data, err := base64.RawURLEncoding.DecodeString(value)

	if err != nil {
		return cursor, ErrInvalidConversationCursor
	}

	if err :=
		json.Unmarshal(data, &cursor); err != nil {
		return cursor, ErrInvalidConversationCursor
	}

	if cursor.ConversationID == 0 {
		return cursor, ErrInvalidConversationCursor
	}

	return cursor, nil
}
