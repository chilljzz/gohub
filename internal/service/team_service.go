package service

import (
	"errors"
	"strings"

	"github.com/chilljzz/gohub/internal/dto"
	"github.com/chilljzz/gohub/internal/model"
	"github.com/chilljzz/gohub/internal/repository"
	"github.com/chilljzz/gohub/internal/request"
)

var (
	ErrTeamNotFound      = errors.New("team not found")
	ErrNotTeamOwner      = errors.New("not team owner")
	ErrNotTeamMember     = errors.New("not team member")
	ErrAlreadyTeamMember = errors.New("already team member")
	ErrInvalidParam      = errors.New("InvalidParam")
)

type TeamService struct {
	teamRepo *repository.TeamRepository
	userRepo *repository.UserRepository
}

func NewTeamService(
	teamRepo *repository.TeamRepository,
	userRepo *repository.UserRepository,
) *TeamService {
	return &TeamService{
		teamRepo: teamRepo,
		userRepo: userRepo,
	}
}

func (s *TeamService) CreateTeam(
	ownerID uint,
	req request.CreateTeamRequest,
) (*dto.TeamResult, error) {
	name := strings.TrimSpace(req.Name)
	description := strings.TrimSpace(req.Description)
	if name == "" {
		return nil, ErrInvalidParam
	}
	team := &model.Team{
		Name:        name,
		Description: description,
		OwnerID:     ownerID,
	}
	if err := s.teamRepo.CreateWithOwner(team); err != nil {
		return nil, err
	}

	return &dto.TeamResult{
		ID:          team.ID,
		Name:        team.Name,
		Description: team.Description,
		OwnerID:     team.OwnerID,
		Role:        model.TeamRoleOwner,
		CreatedAt:   team.CreatedAt,
	}, nil
}

func (s *TeamService) ListMyTeams(
	userID uint,
) ([]dto.TeamResult, error) {
	memberships, err := s.teamRepo.ListMemberships(userID)
	if err != nil {
		return nil, err
	}
	results := make(
		[]dto.TeamResult,
		0,
		len(memberships),
	)
	for _, membership := range memberships {
		results = append(results, dto.TeamResult{
			ID:          membership.Team.ID,
			Name:        membership.Team.Name,
			Description: membership.Team.Description,
			OwnerID:     membership.Team.OwnerID,
			Role:        membership.Role,
			CreatedAt:   membership.Team.CreatedAt,
		})
	}
	return results, nil
}

func (s *TeamService) ListMembers(
	currentUserID uint,
	teamID uint,
) ([]dto.TeamMemberResult, error) {

	team, err := s.teamRepo.FindByID(teamID)
	if err != nil {
		return nil, err
	}

	if team == nil {
		return nil, ErrTeamNotFound
	}

	currentMember, err := s.teamRepo.FindMember(teamID, currentUserID)
	if err != nil {
		return nil, err
	}

	if currentMember == nil {
		return nil, ErrNotTeamMember
	}

	members, err := s.teamRepo.ListMembers(teamID)
	if err != nil {
		return nil, err
	}
	results := make(
		[]dto.TeamMemberResult,
		0,
		len(members),
	)
	for _, member := range members {
		results = append(results, dto.TeamMemberResult{
			ID:       member.User.ID,
			Username: member.User.Username,
			Nickname: member.User.Nickname,
			Avatar:   member.User.Avatar,
			Role:     member.Role,
			JoinedAt: member.CreatedAt,
		})
	}
	return results, nil

}

func (s *TeamService) AddMember(
	currentUserID uint,
	teamID uint,
	req request.AddTeamMemberRequest,
) (*dto.TeamMemberResult, error) {
	team, err := s.teamRepo.FindByID(teamID)
	if err != nil {
		return nil, err
	}

	if team == nil {
		return nil, ErrTeamNotFound
	}

	if team.OwnerID != currentUserID {
		return nil, ErrNotTeamOwner
	}

	username := strings.TrimSpace(req.Username)
	user, err := s.userRepo.FindByUsername(username)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, ErrUserNotFound
	}

	existingMember, err := s.teamRepo.FindMember(teamID, user.ID)
	if err != nil {
		return nil, err
	}

	if existingMember != nil {
		return nil, ErrAlreadyTeamMember
	}

	member := &model.TeamMember{
		TeamID: teamID,
		UserID: user.ID,
		Role:   model.TeamRoleMember,
	}

	if err := s.teamRepo.CreateMember(member); err != nil {
		return nil, err
	}

	return &dto.TeamMemberResult{
		ID:       user.ID,
		Username: user.Username,
		Nickname: user.Nickname,
		Avatar:   user.Avatar,
		Role:     member.Role,
		JoinedAt: member.CreatedAt,
	}, nil

}
