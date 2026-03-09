package messages

import "time"

type Message[T any] struct {
	Payload T `json:"payload"`
}

type DeadLetterMessage[T any] struct {
	OriginalMessage T         `json:"originalMessage"`
	Topic           string    `json:"topic"`
	Partition       int       `json:"partition"`
	Offset          int64     `json:"offset"`
	Key             string    `json:"key"`
	Errors          []string  `json:"errors"`
	FailedAt        time.Time `json:"failedAt"`
	LastError       string    `json:"lastError"`
}
