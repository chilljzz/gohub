package model

type OutBoxEvent struct {
	ID uint64 `gorm:"primaryKey"`

	EventID string `gorm:"column:event_id"`

	AggregateType string `gorm:"column:aggregate_type"`

	AggregateID uint64 `gorm:"column:aggregate_id"`

	EventType string `gorm:"column:event_type"`

	Payload string `gorm:"column:payload;type:json"`

	Status string `gorm:"column:status"`
}

func (OutBoxEvent) TableName() string {
	return "outbox_events"
}
