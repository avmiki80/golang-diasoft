package events

import (
	"github.com/segmentio/kafka-go"
)

func GetHeaderValue(headers []kafka.Header, key string) (string, bool) {
	for _, header := range headers {
		if header.Key == key {
			return string(header.Value), true
		}
	}
	return "", false
}
