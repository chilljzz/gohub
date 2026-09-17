package ws

import "context"

type PresenceTracker interface {
	Touch(
		ctx context.Context,
		userID uint,
	) error
}
