package outbox

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

type fakeStore struct {
	mu        sync.Mutex
	once      sync.Once
	claimed   bool
	published int
	failed    int
	done      chan struct{}
}

func (f *fakeStore) ClaimOne(
	ctx context.Context,
	lease time.Duration,
	maxAttempts int,
) (*Record, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.claimed {
		return nil, nil
	}

	f.claimed = true

	return &Record{
		ID:         1,
		EventID:    "message.created:1",
		RetryCount: 1,
		LeaseToken: "test-token",
	}, nil
}

func (f *fakeStore) MarkPublished(
	_ context.Context,
	_ *Record,
) error {
	f.mu.Lock()
	f.published++
	f.mu.Unlock()

	f.once.Do(func() {
		close(f.done)
	})
	return nil
}

func (f *fakeStore) MarkFailed(
	_ context.Context,
	_ *Record,
	_ error,
	_ time.Duration,
	_ int,
) error {
	f.mu.Lock()
	f.failed++
	f.mu.Unlock()

	f.once.Do(func() {
		close(f.done)
	})
	return nil
}

type fakePublisher struct{ err error }

func (f fakePublisher) PublishOutbox(_ context.Context, _ Record) error { return f.err }

func TestWorkerSuccessAndFailure(t *testing.T) {
	for _, tc := range []struct {
		name                      string
		pubErr                    error
		wantPublished, wantFailed int
	}{
		{name: "success", wantPublished: 1},
		{name: "failure", pubErr: errors.New("broker unreachable"), wantFailed: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			store := &fakeStore{done: make(chan struct{})}
			w, err := NewWorker(
				store,
				fakePublisher{err: tc.pubErr},
				Config{
					PollInterval:   time.Millisecond,
					LeaseTTL:       time.Second,
					PublishTimeout: 100 * time.Millisecond,
					MaxAttempts:    8,
				},
			)
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithCancel(context.Background())
			finished := make(chan error, 1)
			go func() {
				finished <- w.Run(ctx)
			}()

			select {
			case <-store.done:
			case <-time.After(time.Second):
				t.Fatal("worker did not process event")
			}
			cancel()

			select {
			case err := <-finished:
				if err != nil {
					t.Fatal(err)
				}
			case <-time.After(time.Second):
				t.Fatal("worker did not stop")
			}

			store.mu.Lock()
			defer store.mu.Unlock()

			if store.published != tc.wantPublished ||
				store.failed != tc.wantFailed {

				t.Fatalf(
					"published=%d failed=%d",
					store.published,
					store.failed,
				)
			}
		})
	}
}

func TestBackoff(t *testing.T) {
	if backoff(1) != time.Second || backoff(2) != 2*time.Second || backoff(7) != 60*time.Second {
		t.Fatal("unexpected backoff")
	}
}

func TestConfigValidation(t *testing.T) {
	_, err := NewWorker(
		&fakeStore{},
		fakePublisher{},
		Config{
			PollInterval:   time.Second,
			LeaseTTL:       time.Second,
			PublishTimeout: time.Second,
			MaxAttempts:    1,
		})
	if err == nil {
		t.Fatal("publish timeout must be shorter than lease")
	}
}
