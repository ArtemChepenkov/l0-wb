package cache

import (
	"github.com/hashicorp/golang-lru"
	"log"
)

var OrdersCache *lru.Cache

func InitCache(size int) {
	var err error
	OrdersCache, err = lru.New(size)
	if err != nil {
		log.Fatalf("Failed to create cache: %v", err)
	}
}