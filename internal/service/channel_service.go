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
	ErrChannelNotFound = errors.New("channel not found")
	ErrChannelExists   = errors.New("channel already exists")
)

type ChannelService struct {
	channelRepo *repository.ChannelRepository
	teamRepo    *repository.TeamRepository
}

func NewChannelService(
	channelRepo *repository.ChannelRepository,
	teamRepo *repository.TeamRepository,
) *ChannelService {
	return &ChannelService{
		channelRepo: channelRepo,
		teamRepo:    teamRepo,
	}
}

func (s *ChannelService) CreateChannel(
	currentUserID uint,
	teamID uint,
	req request.CreateChannelRequest,
) (*dto.ChannelResult, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, ErrInvalidParam
	}
	description := strings.TrimSpace(req.Description)

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

	findChannel, err := s.channelRepo.FindByTeamIDAndName(teamID, name)
	if err != nil {
		return nil, err
	}
	if findChannel != nil {
		return nil, ErrChannelExists
	}

	channel := &model.Channel{
		TeamID:      teamID,
		Name:        name,
		Description: description,
		CreatedBy:   currentUserID,
	}

	if err := s.channelRepo.CreateWithConversation(channel); err != nil {
		return nil, err
	}

	return &dto.ChannelResult{
		ID:          channel.ID,
		TeamID:      channel.TeamID,
		Name:        channel.Name,
		Description: channel.Description,
		CreatedBy:   channel.CreatedBy,
		CreatedAt:   channel.CreatedAt,
	}, nil

}

func (s *ChannelService) ListChannels(
	currentUserID uint,
	teamID uint,
) ([]dto.ChannelResult, error) {
	team, err := s.teamRepo.FindByID(teamID)
	if err != nil {
		return nil, err
	}
	if team == nil {
		return nil, ErrTeamNotFound
	}

	existingMember, err := s.teamRepo.FindMember(teamID, currentUserID)
	if err != nil {
		return nil, err
	}
	if existingMember == nil {
		return nil, ErrNotTeamMember
	}

	channels, err := s.channelRepo.ListByTeamID(teamID)
	if err != nil {
		return nil, err
	}
	results := make([]dto.ChannelResult, 0, len(channels))
	for _, channel := range channels {
		results = append(results, dto.ChannelResult{
			ID:          channel.ID,
			TeamID:      channel.TeamID,
			Name:        channel.Name,
			Description: channel.Description,
			CreatedBy:   channel.CreatedBy,
			CreatedAt:   channel.CreatedAt,
		})
	}

	return results, nil

}
