package service

import "github.com/chilljzz/gohub/internal/repository"

type RealtimeService struct {
	channelRepo *repository.ChannelRepository
	teamRepo    *repository.TeamRepository
}

func NewRealtimeService(
	channelRepo *repository.ChannelRepository,
	teamRepo *repository.TeamRepository,
) *RealtimeService {
	return &RealtimeService{
		channelRepo: channelRepo,
		teamRepo:    teamRepo,
	}
}

func (s *RealtimeService) CheckChannelAccess(
	userID uint,
	channelID uint,
) error {
	channel, err := s.channelRepo.FindByID(channelID)
	if err != nil {
		return err
	}

	if channel == nil {
		return ErrChannelNotFound
	}

	member, err := s.teamRepo.FindMember(
		channel.TeamID,
		userID,
	)
	if err != nil {
		return err
	}
	if member == nil {
		return ErrNotTeamMember
	}
	return nil
}
