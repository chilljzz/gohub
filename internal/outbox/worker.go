package outbox

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

type Store interface {
	ClaimOne(context.Context, time.Duration, int) (*Record, error)
	MarkPublished(context.Context, *Record) error
	MarkFailed(context.Context, *Record, error, time.Duration, int) error
}

type Publisher interface {
	PublishOutbox(context.Context, Record) error
}

type Config struct {
	PollInterval   time.Duration
	LeaseTTL       time.Duration
	PublishTimeout time.Duration
	MaxAttempts    int
}

type Worker struct {
	store     Store
	publisher Publisher
	cfg       Config
}

func NewWorker(store Store, publisher Publisher, cfg Config) (*Worker, error) {
	if store == nil || publisher == nil || cfg.PollInterval <= 0 ||
		cfg.LeaseTTL <= 0 || cfg.PublishTimeout <= 0 ||
		cfg.PublishTimeout >= cfg.LeaseTTL || cfg.MaxAttempts <= 0 {
		return nil, fmt.Errorf("invalid outbox worker config")
	}

	return &Worker{store: store, publisher: publisher, cfg: cfg}, nil
}

func backoff(attempt int) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	if attempt > 7 {
		attempt = 7
	}

	d := time.Second << (attempt - 1)
	if d > 60*time.Second {
		return 60 * time.Second
	}
	return d
}

func wait(ctx context.Context, duration time.Duration) bool {
	t := time.NewTimer(duration)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-t.C:
		return true
	}
}

func (w *Worker) Run(ctx context.Context) error {
	for ctx.Err() == nil {
		e, err := w.store.ClaimOne(ctx, w.cfg.LeaseTTL, w.cfg.MaxAttempts)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}

			slog.Error("outbox claim failed", "error", err)
			if !wait(ctx, w.cfg.PollInterval) {
				return nil
			}
			continue
		}
		if e == nil {
			if !wait(ctx, w.cfg.PollInterval) {
				return nil
			}
			continue
		}

		pubCtx, pubCancel := context.WithTimeout(ctx, w.cfg.PublishTimeout)
		publishErr := w.publisher.PublishOutbox(pubCtx, *e)
		pubCancel()

		finalCtx, finalCancel := context.WithTimeout(context.Background(), 3*time.Second)
		if publishErr == err {
			err = w.store.MarkPublished(finalCtx, e)
			if err != nil {
				slog.Error("outbox confirmation failed; lease will recover",
					"event_id", e.EventID,
					"error", err,
				)
			} else {
				slog.Info("outbox published",
					"event_id", e.EventID,
					"attempts", e.RetryCount,
				)
			}
		} else {
			err := w.store.MarkFailed(finalCtx, e, publishErr, backoff(e.RetryCount), w.cfg.MaxAttempts)
			if err != nil {
				slog.Error("outbox failure recording failed",
					"event_id", e.EventID,
					"error", err,
				)
			}
			slog.Warn("outbox publish failed",
				"event_id", e.EventID,
				"error", publishErr,
			)
		}
		finalCancel()

	}

	return nil
}
