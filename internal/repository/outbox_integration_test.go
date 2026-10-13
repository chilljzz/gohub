//go:build integration

package repository_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/chilljzz/gohub/internal/event"
	"github.com/chilljzz/gohub/internal/model"
	"github.com/chilljzz/gohub/internal/repository"
	"github.com/chilljzz/gohub/internal/service"
	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	path := filepath.Join(
		"..", "..", ".env.test.local",
	)
	if err := godotenv.Load(path); err != nil {
		t.Fatalf("load test config: %v", err)
	}

	dsn := os.Getenv("GOHUB_TEST_MYSQL_DSN")
	if dsn == "" {
		t.Fatal("GOHUB_TEST_MYSQL_DSN is required")
	}

	db, err := gorm.Open(
		mysql.Open(dsn),
		&gorm.Config{
			TranslateError: true,
		},
	)
	if err != nil {
		t.Fatalf("open MySQL: %v", err)
	}

	// 测试库是独立、可丢弃的 MySQL。
	// 旧业务表按项目原有 GORM Model 初始化。
	// outbox_events 必须由 goose Migration 创建。
	if err := db.AutoMigrate(
		&model.User{},
		&model.Conversation{},
		&model.ConversationMember{},
		&model.Message{},
	); err != nil {
		t.Fatalf("initialize test schema: %v", err)
	}

	if !db.Migrator().HasTable("outbox_events") {
		t.Fatal("outbox_events migration is missing")
	}

	return db
}

func seedDirectConversation(
	t *testing.T,
	db *gorm.DB,
) (uint, uint) {
	t.Helper()

	username := fmt.Sprintf(
		"day3_%d", time.Now().UnixNano(),
	)

	user := model.User{
		Username: username,
		Password: "integration-test-only",
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}

	conv := model.Conversation{
		Type: model.ConversationTypeDirect,
	}
	if err := db.Create(&conv).Error; err != nil {
		t.Fatal(err)
	}

	member := model.ConversationMember{
		ConversationID: conv.ID,
		UserID:         user.ID,
	}
	if err := db.Create(&member).Error; err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		// 按引用依赖顺序清理测试数据。
		db.Exec(
			"DELETE FROM outbox_events WHERE "+
				"aggregate_type = ? AND aggregate_id = ?",
			"conversation",
			conv.ID,
		)
		db.Where(
			"conversation_id = ?", conv.ID,
		).Delete(&model.Message{})
		db.Where(
			"conversation_id = ?", conv.ID,
		).Delete(&model.ConversationMember{})
		db.Delete(&conv)
		db.Delete(&user)
	})

	return user.ID, conv.ID
}

func TestTransactionalOutboxMySQL(t *testing.T) {
	db := openTestDB(t)
	userID, conversationID := seedDirectConversation(t, db)

	messageRepo := repository.NewMessageRepository(db)

	conversationService := service.NewConversationService(
		repository.NewConversationRepository(db),
		nil,
		nil,
		nil,
	)

	messageService := service.NewConversationMessageService(
		messageRepo,
		conversationService,
		nil,
	)

	clientID := fmt.Sprintf(
		"day3-client-%d", time.Now().UnixNano(),
	)

	// Case 1：首次发送成功。
	first, err := messageService.CreateMessage(
		userID,
		conversationID,
		"hello outbox",
		clientID,
	)
	if err != nil {
		t.Fatalf("first create: %v", err)
	}
	if !first.Created {
		t.Fatal("first request should create message")
	}

	eventID := fmt.Sprintf(
		"message.created:%d",
		first.Message.ID,
	)

	var stored model.OutBoxEvent
	if err := db.Where(
		"event_id = ?", eventID,
	).First(&stored).Error; err != nil {
		t.Fatalf("outbox event missing: %v", err)
	}

	if stored.Status != "pending" {
		t.Fatalf("status=%s, want pending", stored.Status)
	}

	var payload event.MessageCreated
	if err := json.Unmarshal(
		[]byte(stored.Payload),
		&payload,
	); err != nil {
		t.Fatalf("invalid event payload: %v", err)
	}

	if payload.MessageID != first.Message.ID ||
		payload.ConversationID != conversationID {
		t.Fatalf("wrong event payload: %+v", payload)
	}

	// Case 2：相同 client_message_id 重复发送。
	repeated, err := messageService.CreateMessage(
		userID,
		conversationID,
		"hello outbox",
		clientID,
	)
	if err != nil {
		t.Fatalf("repeated request: %v", err)
	}

	if repeated.Created ||
		repeated.Message.ID != first.Message.ID {
		t.Fatalf(
			"idempotency failed: %+v",
			repeated,
		)
	}

	var eventCount int64
	if err := db.Model(&model.OutBoxEvent{}).
		Where("event_id = ?", eventID).
		Count(&eventCount).Error; err != nil {
		t.Fatal(err)
	}

	if eventCount != 1 {
		t.Fatalf("event count=%d, want 1", eventCount)
	}

	// Case 3：同一个幂等键被用于不同内容。
	_, err = messageService.CreateMessage(
		userID,
		conversationID,
		"different content",
		clientID,
	)

	if !errors.Is(
		err,
		service.ErrClientMessageConflict,
	) {
		t.Fatalf("want conflict, got %v", err)
	}

	// Case 4：预先制造 Outbox event_id 唯一键冲突。
	// 手动指定一个仅用于隔离测试的消息 ID，
	// 确保第一条 INSERT 可以成功，
	// 而第二条 Outbox INSERT 必然违反唯一约束。

	failedMessageID := uint(
		1000000000 + time.Now().UnixNano()%200000000,
	)
	duplicateEventID := fmt.Sprintf(
		"message.created:%d",
		failedMessageID,
	)

	existingOutbox := model.OutBoxEvent{
		EventID:       duplicateEventID,
		AggregateType: "conversation",
		AggregateID:   uint64(conversationID),
		EventType:     event.TypeMessageCreated,
		Payload:       "{}",
		Status:        "pending",
	}
	if err := db.Create(&existingOutbox).Error; err != nil {
		t.Fatalf("prepare conflict: %v", err)
	}

	failedMessage := &model.Message{
		ID:              failedMessageID,
		ConversationID:  conversationID,
		SenderID:        userID,
		ClientMessageID: "day3-rollback-test",
		Content:         "must rollback",
	}

	err = messageRepo.CreateWithOutbox(failedMessage)

	if !errors.Is(err, gorm.ErrDuplicatedKey) {
		t.Fatalf(
			"want duplicate event error, got %v",
			err,
		)
	}

	var messageCount int64
	if err := db.Model(&model.Message{}).
		Where("id = ?", failedMessageID).
		Count(&messageCount).Error; err != nil {
		t.Fatal(err)
	}

	if messageCount != 0 {
		t.Fatalf(
			"rollback failed: message count=%d",
			messageCount,
		)
	}
}
