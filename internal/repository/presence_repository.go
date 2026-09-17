package repository

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const presnceLastSeenKey = "gohub:presence:last_seen"

type PresenceRepository struct {
	client *redis.Client
}

func NewPresenceRepository(
	client *redis.Client,
) *PresenceRepository {
	return &PresenceRepository{
		client: client,
	}
}

func (r *PresenceRepository) Touch(
	ctx context.Context,
	userID uint,
	at time.Time,
) error {
	member := strconv.FormatUint(
		uint64(userID),
		10,
	)

	return r.client.ZAdd(
		ctx,
		presnceLastSeenKey,
		redis.Z{
			Score:  float64(at.UnixMilli()),
			Member: member,
		},
	).Err()
}

func (r *PresenceRepository) GetLastSeenMany(
	ctx context.Context,
	userIDs []uint,
) (map[uint]time.Time, error) {
	result := make(map[uint]time.Time, len(userIDs))

	if len(userIDs) == 0 {
		return result, nil
	}

	pipe := r.client.Pipeline()

	cmds := make(
		map[uint]*redis.FloatCmd,
		len(userIDs),
	)

	for _, userID := range userIDs {

		if _, exists := cmds[userID]; exists {
			continue
		}

		member := strconv.FormatUint(uint64(userID), 10)

		cmds[userID] = pipe.ZScore(ctx, presnceLastSeenKey, member)

	}

	_, err := pipe.Exec(ctx)

	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, err
	}

	for userID, cmd := range cmds {

		score, err := cmd.Result()

		if errors.Is(err, redis.Nil) {
			continue
		}

		if err != nil {
			return nil, err
		}

		result[userID] = time.UnixMilli(int64(score))
	}

	return result, nil

}
