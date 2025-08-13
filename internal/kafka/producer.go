package kafka

import (
	"encoding/json"
	_ "log"

	sarama "github.com/IBM/sarama"
	"l0-wb/internal/model"
)

type SyncProducer struct {
	prod sarama.SyncProducer
}

func NewProducer(brokers []string) (*SyncProducer, error) {
	cfg := sarama.NewConfig()
	cfg.Producer.Return.Successes = true
	cfg.Producer.RequiredAcks = sarama.WaitForAll
	p, err := sarama.NewSyncProducer(brokers, cfg)
	if err != nil {
		return nil, err
	}
	return &SyncProducer{prod: p}, nil
}

func (p *SyncProducer) Send(o model.Order) error {
	b, err := json.Marshal(o)
	if err != nil {
		return err
	}
	msg := &sarama.ProducerMessage{
		Topic: "upload-order-topic",
		Key:   sarama.StringEncoder(o.OrderUID),
		Value: sarama.ByteEncoder(b),
	}
	_, _, err = p.prod.SendMessage(msg)
	return err
}

func (p *SyncProducer) Close() error { return p.prod.Close() }
