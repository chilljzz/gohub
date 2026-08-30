package request

type CreateChannelRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=50"`
	Description string `json:"description" binding:"omitempty,max=200"`
}

type MarkChannelReadRequest struct {
	MessageID uint `json:"message_id" binding:"required"`
}
