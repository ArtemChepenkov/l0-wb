package service

import (
	"errors"
	"l0-wb/internal/cache"
	"l0-wb/internal/model"
	"l0-wb/internal/repo"

	kafkaProd "l0-wb/internal/kafka"
)

type OrderDTO = model.Order

type OrderService struct {
	repo     *repo.PostgresRepo
	cache    *cache.CacheAdapter
	producer *kafkaProd.SyncProducer
}

func NewOrderService(r *repo.PostgresRepo, c *cache.CacheAdapter, p *kafkaProd.SyncProducer) *OrderService {
	return &OrderService{repo: r, cache: c, producer: p}
}

func (s *OrderService) GetByUID(uid string) (model.Order, error) {
	if o, ok := s.cache.Get(uid); ok {
		return o, nil
	}
	o, err := s.repo.GetFull(uid)
	if err != nil {
		return model.Order{}, err
	}
	s.cache.Set(uid, o)
	return o, nil
}

func (s *OrderService) Create(o OrderDTO) error {
	if o.OrderUID == "" {
		return errors.New("order_uid required")
	}
	if err := s.repo.SaveOrder(o); err != nil {
		return err
	}
	s.cache.Set(o.OrderUID, o)

	if s.producer != nil {
		_ = s.producer.Send(o)
	}
	return nil
}
