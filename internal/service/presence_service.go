package service

import (
	"context"
	"time"
)

type PresenceStore interface {
	Touch(
		ctx context.Context,
		userID uint,
		at time.Time,
	) error

	GetLastSeenMany(
		ctx context.Context,
		userIDs []uint,
	) (map[uint]time.Time, error)
}

type PresenceStatus struct {
	Online     bool
	LastSeenAt *time.Time
}

type PresenceService struct {
	store   PresenceStore
	timeout time.Duration
}

func NewPresenceService(
	store PresenceStore,
	timeout time.Duration,
) *PresenceService {
	return &PresenceService{
		store:   store,
		timeout: timeout,
	}

}

func (s *PresenceService) Touch(
	ctx context.Context,
	userID uint,
) error {
	return s.store.Touch(
		ctx,
		userID,
		time.Now(),
	)
}

func (s *PresenceService) GetMany(
	ctx context.Context,
	userIDs []uint,
) (map[uint]PresenceStatus, error) {
	lastSeenMap, err := s.store.GetLastSeenMany(ctx, userIDs)

	if err != nil {
		return nil, err
	}

	result := make(map[uint]PresenceStatus, len(userIDs))

	now := time.Now()

	for _, userID := range userIDs {
		lastSeen, ok := lastSeenMap[userID]

		if !ok {
			result[userID] = PresenceStatus{Online: false}
			continue
		}

		online := now.Sub(lastSeen) <= s.timeout

		result[userID] = PresenceStatus{
			Online:     online,
			LastSeenAt: &lastSeen,
		}

	}
	return result, nil

}
