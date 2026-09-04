package request

type CreateDirectConversationRequest struct {
	UserID uint `json:"user_id" binding:"required"`
}
