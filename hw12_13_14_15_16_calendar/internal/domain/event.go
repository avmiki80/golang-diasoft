package domain

import "time"

type Event struct {
	ID          string        `db:"id" json:"id"`
	CreatedAt   time.Time     `db:"created_at" json:"-"`
	UpdatedAt   time.Time     `db:"updated_at" json:"-"`
	Title       string        `db:"title" json:"title"`
	StartDate   time.Time     `db:"start_date" json:"startDate"`
	EndDate     time.Time     `db:"end_date" json:"endDate"`
	Description string        `db:"description" json:"description"`
	UserID      string        `db:"user_id" json:"userId"`
	OffsetTime  time.Duration `db:"offset_time" json:"offsetTime"`
}
