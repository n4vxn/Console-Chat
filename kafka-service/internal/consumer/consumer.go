package consumer

import "github.com/confluentinc/confluent-kafka-go/v2/kafka"

func CreateProducer() (*kafka.Producer, error) {
	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092",
	})
}