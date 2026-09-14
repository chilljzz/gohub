package request

type ListConversationQuery struct {
	Limit int `form:"limit"`

	Cursor string `form:"cursor"`
}
