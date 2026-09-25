package dto

type DirectReceiptStatrResult struct {
	ConversationID uint `json:"conversation_id"`

	PeerUserID uint `json:"peer_user_id"`

	LastDeliveredMessageID uint `json:"last_delivered_message_id"`

	LastReadMessageID uint `json:"last_read_message_id"`
}
