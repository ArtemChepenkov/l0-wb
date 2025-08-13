package cache

import (
	"context"
	"log"
	"sync"
	_ "time"
	"database/sql"

	lru "github.com/hashicorp/golang-lru"
	"l0-wb/internal/model"
	"l0-wb/internal/repo"
)

var lc *lru.Cache
var mu sync.RWMutex

func InitCache(size int) error {
	var err error
	lc, err = lru.New(size)
	return err
}

func Self() *CacheAdapter {
	return &CacheAdapter{}
}

type CacheAdapter struct{}

func (c *CacheAdapter) Get(key string) (model.Order, bool) {
	mu.RLock()
	defer mu.RUnlock()
	if lc == nil {
		return model.Order{}, false
	}
	v, ok := lc.Get(key)
	if !ok {
		return model.Order{}, false
	}
	return v.(model.Order), true
}

func (c *CacheAdapter) Set(key string, val model.Order) {
	mu.Lock()
	defer mu.Unlock()
	if lc == nil {
		return
	}
	_ = lc.Add(key, val)
}

func (c *CacheAdapter) Len() int {
	mu.RLock()
	defer mu.RUnlock()
	if lc == nil {
		return 0
	}
	return lc.Len()
}

func Preload(ctx context.Context, dbConn *sql.DB, limit int) {
	log.Printf("Starting cache preload (limit=%d)\n", limit)
	rows, err := dbConn.Query(`SELECT order_uid FROM orders ORDER BY date_created DESC LIMIT $1`, limit)
	if err != nil {
		log.Printf("Failed to preload cache: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		select {
		case <-ctx.Done():
			log.Println("preload cancelled")
			return
		default:
		}
		var uid string
		if err := rows.Scan(&uid); err != nil {
			continue
		}
		order, err := repo.NewPostgresRepo(dbConn).GetFull(uid)
		if err == nil {
			Self().Set(uid, order)
		}
	}
	log.Printf("Cache preload done, size=%d\n", Self().Len())
}
