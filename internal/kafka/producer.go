package kafka

import (
	"log"
	"github.com/Shopify/sarama"
	data "l0-wb/internal/db"
	"encoding/json"
)

const (
	topic = "upload-order-topic"
)

func StartProducer(order data.Order, ch chan struct{}) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Retry.Max = 5
	config.Producer.RequiredAcks = sarama.WaitForAll
	brokers := []string{"kafka:9092"}
	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		log.Fatalf("cannot create producer: %v\n", err)
	}
	defer producer.Close()
	
	messageBytes, err := json.Marshal(order);

	message := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder("key"),
		Value: sarama.ByteEncoder(messageBytes),
	}
	partition, offset, err := producer.SendMessage(message)

	if err != nil {
		log.Fatalf("Send error: %v\n", err)
	} else {
		ch <- struct{}{}
		log.Printf("Sent to partition %d at offset %d\n", partition, offset)
	}
}
