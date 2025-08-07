package kafka

import (
	"encoding/json"
	"github.com/Shopify/sarama"
	"time"
	"l0-wb/internal/cache"
	"database/sql"
	"log"
	"fmt"
	data "l0-wb/internal/db"
)

const (
	startTryes = 100
)

func parseMessage(msg []byte) *data.Order {
	var order data.Order
	err := json.Unmarshal(msg, &order)
	if err != nil {
		log.Fatalf("Failed to parse message: %v", err)
	}

	return &order
}

func saveOrder(order data.Order, db *sql.DB) {
	err := data.LoadFullINfo(db, order)
	if err != nil {
		log.Fatalf("Failed to save to DB: %v", err)
	}
	cache.OrdersCache.Add(order.OrderUID, order)
}

func StartConsumer(db *sql.DB) {
	brokers := []string{"kafka:9092"}
	var err error
	var consumer sarama.Consumer
	for i := 0; i < startTryes; i++ {
		consumer, err = sarama.NewConsumer(brokers, nil)
		if err == nil {
			log.Println("consumer started")
			break
		}
		log.Println("Retry consumer start")
		time.Sleep(time.Second)
	}
	if err != nil {
		log.Fatalf("cannot create consumer: %v\n", err)
	} 
	defer consumer.Close()

	partitionConsumer, err := consumer.ConsumePartition(topic, 0, sarama.OffsetNewest)

	if err != nil {
		log.Fatalf("cannot consume partition: %v\n", err)
	}
	
	defer partitionConsumer.Close()

	for {
		msg := <-partitionConsumer.Messages()
		fmt.Printf("📨Received message: %s (offset %d)\n", string(msg.Value), msg.Offset)
		curOrder := parseMessage(msg.Value)
		saveOrder(*curOrder, db)
	}
}