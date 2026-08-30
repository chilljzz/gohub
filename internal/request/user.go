package request

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=20"`
	Password string `json:"password" binding:"required,min=6,max=32"`
	Nickname string `json:"nickname" binding:"omitempty,max=30"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required,min=3,max=20"`
	Password string `json:"password" binding:"required,min=6,max=32"`
}

type UpdateProfileRequest struct {
	Nickname *string `json:"nickname" binding:"omitempty,min=1,max=30"`
	Avatar   *string `json:"avater" binding:"omitempty,max=255"`
	Bio      *string `json:"bio" binging:"omitempty,max=255"`
}

type SendFriendRequest struct {
	Username string `json:"username" binding:"required,min=3,max=20"`
}
