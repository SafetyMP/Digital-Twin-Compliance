package consumer

import "github.com/segmentio/kafka-go"

type Runner struct {
	writer *kafka.Writer
}
