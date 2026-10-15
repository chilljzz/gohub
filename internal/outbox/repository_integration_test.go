//go:build integration

package outbox

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/chilljzz/gohub/internal/model"
	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func isolatedDB(t *testing.T) *gorm.DB {
	t.Helper()
	_ = godotenv.Load(filepath.Join("..", "..", ".env.test.local"))
	dsn := os.Getenv("GOHUB_TEST_MYSQL_DSN")
	if dsn == "" {
		t.Fatal("GOHUB_TEST_MYSQL_DSN must target a disposable and isolated MySQL database")
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{TranslateError: true})
	if err != nil {
		t.Fatal(err)
	}
	if !db.Migrator().HasColumn("outbox_events", "lease_token") {
		t.Fatal("run M02 goose migration on test DB first")
	}
	var waiting int64

	if err := db.Model(&Record{}).
		Where("status IN ?",
			[]string{StatusPending, StatusProcessing},
		).Count(&waiting).Error; err != nil {
		t.Fatal(err)
	}
	if waiting != 0 {
		t.Skipf("test database already has %d in-flight jobs; use isolated database", waiting)
	}
	return db
}

func fixture(t *testing.T, db *gorm.DB) {
	t.Helper()
	eventID := fmt.Sprintf("m02-test-%d", time.Now().UnixNano())
	e := model.OutBoxEvent{EventID: eventID, AggregateType: "conversation", AggregateID: 1, EventType: "message.created", Payload: `{"event_id":"test"}`, Status: "pending"}
	if err := db.Create(&e).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Unscoped().Where("event_id = ?", eventID).Delete(&model.OutBoxEvent{}) })
}

func TestConcurrentClaimAndLeaseFencingMySQL(t *testing.T) {
	db := isolatedDB(t)
	fixture(t, db)
	repo := NewRepository(db)
	ctx := context.Background()
	var wg sync.WaitGroup
	var won atomic.Int64
	var first *Record
	var mu sync.Mutex

	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			e, err := repo.ClaimOne(ctx, 200*time.Millisecond, 4)
			if err != nil {
				t.Errorf("claim: %v", err)
				return
			}
			if e != nil {
				won.Add(1)
				mu.Lock()
				first = e
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	if won.Load() != 1 {
		t.Fatalf("claimed %d times, expected 1", won.Load())
	}
	time.Sleep(350 * time.Millisecond)
	second, err := repo.ClaimOne(ctx, time.Second, 4)
	if err != nil || second == nil {
		t.Fatalf("recovery claim: %v, %#v", err, second)
	}
	if first.LeaseToken == second.LeaseToken {
		t.Fatal("lease tokens must differ")
	}
	if err := repo.MarkPublished(ctx, first); !errors.Is(err, ErrLeaseLost) {
		t.Fatalf("stale worker ack should fail: %v", err)
	}
	if err := repo.MarkPublished(ctx, second); err != nil {
		t.Fatalf("new worker ack: %v", err)
	}
}

func TestBackoffRetryMySQL(t *testing.T) {
	db := isolatedDB(t)
	fixture(t, db)
	repo := NewRepository(db)
	ctx := context.Background()
	first, err := repo.ClaimOne(ctx, 3*time.Second, 4)
	if err != nil || first == nil {
		t.Fatalf("claim: %v", err)
	}
	if err := repo.MarkFailed(ctx, first, errors.New("kafka down"), time.Second, 4); err != nil {
		t.Fatal(err)
	}
	immediate, err := repo.ClaimOne(ctx, 3*time.Second, 4)
	if err != nil || immediate != nil {
		t.Fatalf("backoff not enforced: %v, %#v", err, immediate)
	}
	time.Sleep(1200 * time.Millisecond)
	second, err := repo.ClaimOne(ctx, 3*time.Second, 4)
	if err != nil || second == nil || second.RetryCount != 2 {
		t.Fatalf("retry claim: %v, %#v", err, second)
	}
	if err := repo.MarkPublished(ctx, second); err != nil {
		t.Fatal(err)
	}
}
