package request

type CreateTeamRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=50"`
	Description string `json:"Description" binding:"omitempty,max=2000"`
}

type AddTeamMemberRequest struct {
	Username string `json:"username" binging:"required,min=3,max=20"`
}
