package domain

import "time"

type DLQAlert struct {
	ID              string    `db:"id" json:"id"`
	Topic           string    `db:"topic" json:"topic"`
	OriginalTopic   string    `db:"original_topic" json:"originalTopic"`
	Partition       int       `db:"partition" json:"partition"`
	MessageOffset   int64     `db:"message_offset" json:"messageOffset"`
	MessageKey      string    `db:"message_key" json:"messageKey"`
	ErrorMessage    string    `db:"error_message" json:"errorMessage"`
	ErrorCount      int       `db:"error_count" json:"errorCount"`
	OriginalMessage string    `db:"original_message" json:"originalMessage"`
	FailedAt        time.Time `db:"failed_at" json:"failedAt"`
	CreatedAt       time.Time `db:"created_at" json:"createdAt"`
	Status          string    `db:"status" json:"status"` // new, processing, resolved, failed
}
