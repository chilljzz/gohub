package repository

import (
	"errors"

	"github.com/chilljzz/gohub/internal/model"
	"gorm.io/gorm"
)

type ConversationRepository struct {
	db *gorm.DB
}

func NewConversationRepository(
	db *gorm.DB,
) *ConversationRepository {
	return &ConversationRepository{
		db: db,
	}
}

func (r *ConversationRepository) Create(
	conversation *model.Conversation,
) error {
	return r.db.Create(conversation).Error
}

func (r *ConversationRepository) FindByID(
	id uint,
) (*model.Conversation, error) {
	var conversation model.Conversation

	err := r.db.First(&conversation, id).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &conversation, nil
}

func (r *ConversationRepository) FindByChannelID(
	channelID uint,
) (*model.Conversation, error) {
	var conversation model.Conversation

	err := r.db.
		Where(
			"type = ? AND channel_id = ?",
			model.ConversationTypeChannel,
			channelID,
		).
		First(
			&conversation,
		).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &conversation, nil
}

func (r *ConversationRepository) AddMember(
	member *model.ConversationMember,
) error {
	return r.db.Create(member).Error
}

func (r *ConversationRepository) IsMember(
	conversationID uint,
	userID uint,
) (bool, error) {
	var count int64

	err := r.db.
		Model(
			&model.ConversationMember{},
		).
		Where(
			"conversation_id = ? AND user_id = ?",
			conversationID,
			userID,
		).
		Count(
			&count,
		).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *ConversationRepository) FindByDirectKey(
	directKey string,
) (*model.Conversation, error) {
	var conversation model.Conversation

	err := r.db.
		Where(
			"type = ? AND direct_key = ?",
			model.ConversationTypeDirect,
			directKey,
		).First(&conversation).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &conversation, nil
}

func (r *ConversationRepository) ListMember(
	conversationID uint,
) ([]model.ConversationMember, error) {
	var members []model.ConversationMember

	err := r.db.
		Where(
			"conversation_id = ?",
			conversationID,
		).
		Find(&members).Error

	if err != nil {
		return nil, err
	}

	return members, nil
}

func (r *ConversationRepository) CreateDirectWithMembers(
	conversation *model.Conversation,
	userID uint,
	otherUserID uint,
) error {
	return r.db.Transaction(
		func(tx *gorm.DB) error {
			if err := tx.Create(conversation).Error; err != nil {
				return err
			}

			members := []model.ConversationMember{
				{
					ConversationID: conversation.ID,
					UserID:         userID,
				},
				{
					ConversationID: conversation.ID,
					UserID:         otherUserID,
				},
			}
			if err := tx.Create(&members).Error; err != nil {
				return err
			}

			return nil
		},
	)
}

func (r *ConversationRepository) ListForUser(
	userID uint,
	cursorLastMessageID uint,
	cursorConversationID uint,
	hasCursor bool,
	limit int,
) ([]model.ConversationListProjection,
	bool,
	error,
) {
	var rows []model.ConversationListProjection

	cursorEnabled := 0
	if hasCursor {
		cursorEnabled = 1
	}

	const query = `
	WITH visible_conversations AS(
		-- DirectConversation:
		SELECT
			cm.conversation_id
		FROM conversation_members cm
		INNER JOIN conversations c
			ON c.id = cm.conversation_id
			AND c.type = 'direct'
		WHERE cm.user_id = ?

		UNION
		-- Channel Conversation:
		SELECT
			c.id AS conversation_id
		FROM team_members tm
		INNER JOIN channels ch
			ON ch.team_id = tm.team_id
		INNER JOIN conversations c
			ON c.type = 'channel'
			AND c.channel_id = ch.id
		WHERE tm.user_id = ?
	),

	last_messages AS (
		-- conversation last messageID
		SELECT 
			m.conversation_id,
			MAX(m.id) AS last_message_id
		FROM messages m
		GROUP BY m.conversation_id
	),

	last_message_details AS (
		-- last message model
		SELECT
			m.id,
			m.conversation_id,
			m.sender_id,
			m.content,
			m.created_at
		FROM messages m
		INNER JOIN last_messages lm
			ON lm.last_message_id = m.id
	),

	user_reads AS (
		-- user conversation last read
		SELECT
			cr.conversation_id,
			cr.last_read_message_id
		FROM conversation_reads cr
		WHERE cr.user_id = ?
	),

	unread_counts AS (
		SELECT
			m.conversation_id,
			COUNT(*) AS unread_count
		FROM messages m
		LEFT JOIN user_reads ur
			ON ur.conversation_id = m.conversation_id
		WHERE 
			m.id > COALESCE(ur.last_read_message_id,0)
			AND
			m.sender_id <> ?
		GROUP BY m.conversation_id
	)
	
	SELECT
	c.id AS conversation_id,
	c.type AS conversation_type,

	c.channel_id AS channel_id,
	ch.name AS channel_name,

	ch.team_id AS team_id,
	t.name AS team_name,

	peer.id AS peer_user_id,
	peer.username AS peer_username,
	peer.nickname AS peer_nickname,
	peer.avatar AS peer_avatar,

	lmd.id AS last_message_id,
	lmd.sender_id AS last_sender_id,
	lmd.content AS last_message_content,
	lmd.created_at AS last_message_at,

	COALESCE(
		ur.last_read_message_id,
		0
	) AS last_read_message_id,

	COALESCE(
		uc.unread_count,
		0
	) AS unread_count

	FROM visible_conversations vc

	INNER JOIN conversations c
		ON c.id = vc.conversation_id

	LEFT JOIN last_message_details lmd
		ON lmd.conversation_id = c.id

	LEFT JOIN user_reads ur
		ON ur.conversation_id = c.id

	LEFT JOIN unread_counts uc
		ON uc.conversation_id = c.id

	LEFT JOIN channels ch
		ON ch.id = c.channel_id

	LEFT JOIN teams t
		ON t.id = ch.team_id


	LEFT JOIN conversation_members peer_cm
		ON c.type = 'direct'
		AND peer_cm.conversation_id = c.id
		AND peer_cm.user_id <> ?

	LEFT JOIN users peer
		ON peer.id = peer_cm.user_id


	WHERE(
		? = 0 
		OR COALESCE(lmd.id,0) < 0
		OR COALESCE(lmd.id,0) < ?
		OR (COALESCE(lmd.id,0) = ? AND c.id < ?)
	)
	ORDER BY
		COALESCE(lmd.id,0) DESC,
		c.id DESC

	LIMIT ?


	`

	err := r.db.
		Raw(
			query,
			//visible_conversations:
			userID,
			//visible_conversations:
			userID,
			//user_reads:
			userID,
			//unread_counts:
			userID,
			//peer_cm:
			userID,
			//cursor:
			cursorEnabled,

			// last_message_id < cursor
			cursorLastMessageID,

			// last_message_id = cursor
			cursorLastMessageID,

			// conversation_id < cursor
			cursorConversationID,

			limit+1,
		).
		Scan(&rows).Error

	if err != nil {
		return nil, false, err
	}

	hasMore := len(rows) > limit

	if hasMore {
		rows = rows[:limit]
	}

	return rows, hasMore, nil

}
