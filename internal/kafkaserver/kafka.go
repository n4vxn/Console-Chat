package kafkaserver

import (
	"fmt"
	"log"

	"github.com/IBM/sarama"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type KafkaConfig struct {
	Broker string
	Topic  string
}

func StartProducer(kafkaConfig *KafkaConfig) (sarama.SyncProducer, error) {
	// Set up Sarama configuration for the producer
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll // Wait for acknowledgment from all replicas
	config.Producer.Retry.Max = 5                    // Max retries if message delivery fails
	config.Producer.Return.Successes = true          // Receive successes asynchronously

	// Create a new sync producer instance
	producer, err := sarama.NewSyncProducer([]string{kafkaConfig.Broker}, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %v", err)
	}

	return producer, nil
}

func StartConsumer(kafkaConfig *KafkaConfig) (*kafka.Consumer, error) {
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": kafkaConfig.Broker,
		"group.id":          "chat-consumer-group",
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer: %v", err)
	}
	err = consumer.Subscribe(kafkaConfig.Topic, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to topic: %v", err)
	}
	return consumer, nil
}

func ConsumeMessages(consumer *kafka.Consumer, handler func([]byte)) {
	for {
		msg, err := consumer.ReadMessage(-1)
		if err != nil {
			log.Printf("Error reading message: %v\n", err)
			continue
		}
		handler(msg.Value)
	}
}