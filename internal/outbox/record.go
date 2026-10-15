package outbox

import "time"

type Record struct {
	ID            uint64     `gorm:"primaryKey"`
	EventID       string     `gorm:"column:event_id"`
	AggregateType string     `gorm:"column:aggregate_type"`
	AggregateID   uint64     `gorm:"column:aggregate_id"`
	EventType     string     `gorm:"column:event_type"`
	Payload       string     `gorm:"column:payload;type:json"`
	Status        string     `gorm:"column:status"`
	RetryCount    int        `gorm:"column:retry_count"`
	NextRetryAt   time.Time  `gorm:"column:next_retry_at"`
	LockedUntil   *time.Time `gorm:"column:locked_until"`
	LeaseToken    string     `gorm:"column:lease_token"`
	LastError     *string    `gorm:"column:last_error"`
	PublishedAt   *time.Time `gorm:"column:published_at"`
}

func (Record) TableName() string { return "outbox_events" }
