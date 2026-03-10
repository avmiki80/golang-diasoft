package domain

import "time"

type Notification struct {
	ID        string    `db:"id" json:"id"`
	CreatedAt time.Time `db:"created_at" json:"-"`
	Title     string    `db:"title" json:"title"`
	StartDate time.Time `db:"start_date" json:"startDate"`
	EndDate   time.Time `db:"end_date" json:"endDate"`
	EventID   string    `db:"event_id" json:"eventId"`
	UserID    string    `db:"user_id" json:"userId"`
}
