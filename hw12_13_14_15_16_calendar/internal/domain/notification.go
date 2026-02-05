package domain

import "time"

type Notification struct {
	ID     string
	Title  string
	Date   time.Time
	UserID string
}
