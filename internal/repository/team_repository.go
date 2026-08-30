package repository

import (
	"errors"

	"github.com/chilljzz/gohub/internal/database"
	"github.com/chilljzz/gohub/internal/model"
	"gorm.io/gorm"
)

type TeamRepository struct {
}

func NewTeamRepository() *TeamRepository {
	return &TeamRepository{}
}

func (r *TeamRepository) CreateWithOwner(
	team *model.Team,
) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(team).Error; err != nil {
			return err
		}

		ownerMeber := model.TeamMember{
			TeamID: team.ID,
			UserID: team.OwnerID,
			Role:   model.TeamRoleOwner,
		}
		if err := tx.Create(&ownerMeber).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *TeamRepository) ListMemberships(
	UserID uint,
) ([]model.TeamMember, error) {
	var members []model.TeamMember
	err := database.DB.
		Preload("Team").
		Where("user_id = ?", UserID).
		Order("created_at DESC").
		Find(&members).Error

	return members, err
}

func (r *TeamRepository) FindByID(
	teamID uint,
) (*model.Team, error) {
	var team model.Team
	err := database.DB.
		Find(&team, teamID).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &team, err
}

func (r *TeamRepository) FindMember(
	teamID uint,
	userID uint,
) (*model.TeamMember, error) {
	var teamMember model.TeamMember
	err := database.DB.
		Where(
			"team_id = ? AND user_id = ?",
			teamID,
			userID,
		).First(&teamMember).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {

		return nil, err
	}
	return &teamMember, nil
}

func (r *TeamRepository) CreateMember(
	member *model.TeamMember,
) error {
	return database.DB.Create(member).Error
}

func (r *TeamRepository) ListMembers(
	teamID uint,
) ([]model.TeamMember, error) {
	var members []model.TeamMember
	err := database.DB.
		Preload("User").
		Where("team_id = ?", teamID).
		Order("created_at ASC").
		Find(&members).Error

	return members, err
}
