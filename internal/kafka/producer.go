package kafka

import (
	_ "fmt"
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
		log.Printf("cannot create producer: %v\n", err)
		panic("cannot create producer")
	}
	defer producer.Close()
	
	messageBytes, err := json.Marshal(order);
log.Printf("bbbbbbbbbbbbb\n")
	message := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder("key"),
		Value: sarama.ByteEncoder(messageBytes),
	}
	partition, offset, err := producer.SendMessage(message)
	log.Printf("aaaaaaaaaaaaaaaaaaa\n")
	if err != nil {
		log.Printf("send error: %v\n", err)
		panic("send error")
	} else {
		ch <- struct{}{}
		log.Printf("✅ Sent to partition %d at offset %d\n", partition, offset)
	}
}
