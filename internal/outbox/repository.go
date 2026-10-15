package outbox

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	StatusPending    = "pending"
	StatusProcessing = "processing"
	StatusPublished  = "published"
	StatusDead       = "dead"
)

var ErrLeaseLost = errors.New("outbox lease ownership lost")

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func newToken() (string, error) {
	var b [16]byte

	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("create lease token: %w ", err)
	}

	return hex.EncodeToString(b[:]), nil
}

func (r *Repository) ClaimOne(
	ctx context.Context,
	leaseTTL time.Duration,
	maxAttempts int,
) (*Record, error) {
	var claimed *Record
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var e Record
		err := tx.Clauses(clause.Locking{
			Strength: "UPDATE",
			Options:  "SKIP LOCKED",
		}).Where(
			"(status = ? AND next_retry_at <= NOW(3)) "+
				"OR"+
				"(status = ? AND locked_until <= NOW(3))",
			StatusPending,
			StatusProcessing,
		).
			Order("id ASC").
			Take(&e).Error

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		if err != nil {
			return err
		}

		if e.RetryCount >= maxAttempts {
			return tx.Model(&Record{}).
				Where("id=?", e.ID).
				Updates(map[string]any{
					"status":       StatusDead,
					"lease_token":  nil,
					"locked_until": nil,
					"last_error":   "max attempts exhausted(including crashes)",
				}).Error
		}
		token, err := newToken()
		if err != nil {
			return err
		}

		if err := tx.Model(&Record{}).
			Where("id = ?", e.ID).
			Updates(map[string]any{
				"status":      StatusProcessing,
				"lease_token": token,
				"locked_until": gorm.Expr(
					"TIMESTAMPADD(MICROSECOND,?,NOW(3))",
					leaseTTL.Microseconds(),
				),
				"retry_count": gorm.Expr("retry_count+1"),
			}).Error; err != nil {
			return nil
		}

		e.LeaseToken = token
		e.RetryCount++
		claimed = &e
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("claim outbox: %w", err)
	}
	return claimed, nil
}

func (r *Repository) MarkPublished(ctx context.Context, e *Record) error {
	res := r.db.WithContext(ctx).Model(&Record{}).
		Where("id = ? AND status = ? AND lease_token = ? AND locked_until > NOW(3)",
			e.ID,
			StatusProcessing,
			e.LeaseToken,
		).Updates(
		map[string]any{
			"status":       StatusPublished,
			"published_at": gorm.Expr("NOW(3)"),
			"lease_token":  nil,
			"locked_until": nil,
			"last_error":   nil,
		},
	)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected != 1 {
		return ErrLeaseLost
	}

	return nil
}

func (r *Repository) MarkFailed(
	ctx context.Context,
	e *Record,
	publishErr error,
	delay time.Duration,
	maxAttempts int,
) error {
	status := StatusPending
	if e.RetryCount >= maxAttempts {
		status = StatusDead
	}

	runes := []rune(publishErr.Error())
	if len(runes) > 400 {
		runes = runes[:400]
	}

	res := r.db.WithContext(ctx).Model(&Record{}).
		Where("id = ? AND status = ? AND lease_token = ? AND locked_until > NOW(3)",
			e.ID,
			StatusProcessing,
			e.LeaseToken,
		).Updates(map[string]any{
		"status":        status,
		"lease_token":   nil,
		"locked_until":  nil,
		"last_error":    string(runes),
		"next_retry_at": gorm.Expr("TIMESTAMPADD(MICROSECOND, ?, NOW(3))", delay.Microseconds()),
	})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected != 1 {
		return ErrLeaseLost
	}
	return nil

}
