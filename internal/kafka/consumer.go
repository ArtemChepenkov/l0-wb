package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"
	"database/sql"

	sarama "github.com/IBM/sarama"
	"l0-wb/internal/cache"
	data "l0-wb/internal/model"
	"l0-wb/internal/repo"
)

type Consumer struct {
	brokers []string
	topic   string
	dbConn  *sql.DB
}

func NewConsumer(brokers []string, topic string, dbConn *sql.DB) *Consumer {
	return &Consumer{brokers: brokers, topic: topic, dbConn: dbConn}
}

func (c *Consumer) Run(ctx context.Context) {
	var consumer sarama.Consumer
	var err error
	for i := 0; i < 60; i++ {
		consumer, err = sarama.NewConsumer(c.brokers, nil)
		if err == nil {
			break
		}
		log.Println("Retry consumer start")
		time.Sleep(time.Second)
	}
	if err != nil {
		log.Printf("cannot create consumer: %v\n", err)
		return
	}
	defer func() { _ = consumer.Close() }()

	partitionConsumer, err := consumer.ConsumePartition(c.topic, 0, sarama.OffsetNewest)
	if err != nil {
		log.Printf("cannot consume partition: %v\n", err)
		return
	}
	defer func() { _ = partitionConsumer.Close() }()

	for {
		select {
		case <-ctx.Done():
			log.Println("consumer: ctx done, stopping")
			return
		case msg, ok := <-partitionConsumer.Messages():
			if !ok {
				log.Println("consumer: messages closed")
				return
			}
			log.Printf("Received: %s (offset %d)\n", string(msg.Value), msg.Offset)
			var ord data.Order
			if err := json.Unmarshal(msg.Value, &ord); err != nil {
				log.Printf("invalid json: %v", err)
				continue
			}
			if err := repo.NewPostgresRepo(c.dbConn).SaveOrder(ord); err != nil {
				log.Printf("save order err: %v", err)
				continue
			}
			cache.Self().Set(ord.OrderUID, ord)
		}
	}
}
