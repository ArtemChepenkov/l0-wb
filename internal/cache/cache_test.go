package cache

import (
	"testing"

	"l0-wb/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestCache_SetAndGet(t *testing.T) {
	err := InitCache(2)
	assert.NoError(t, err)

	c := Self()

	order := model.Order{OrderUID: "123"}
	c.Set("123", order)

	got, ok := c.Get("123")
	assert.True(t, ok)
	assert.Equal(t, "123", got.OrderUID)

	// Проверим отсутствующий ключ
	_, ok = c.Get("not-exist")
	assert.False(t, ok)
}

func TestCache_Len(t *testing.T) {
	_ = InitCache(3)
	c := Self()

	assert.Equal(t, 0, c.Len())

	c.Set("1", model.Order{OrderUID: "1"})
	c.Set("2", model.Order{OrderUID: "2"})
	assert.Equal(t, 2, c.Len())
}
