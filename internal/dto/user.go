package dto

type RegisterResult struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
}

type LoginResult struct {
	Token string `json:"token"`
}

type UserProfileResult struct {
	ID        uint   `json:"id"`
	Username  string `json:"username"`
	Nickname  string `json:"nickname"`
	Avatar    string `json:"avatar"`
	Bio       string `json:"bio"`
	CreatedAt string `json:"created_at"`
}

type UserBridfResult struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}
