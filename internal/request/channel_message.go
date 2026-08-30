package request

type SyncChannelMessageQuery struct {
	AfterID uint `form:"after_id" `
	Limit   int  `form:"limit" blinding:"omitempty,min=1,max=500"`
}
