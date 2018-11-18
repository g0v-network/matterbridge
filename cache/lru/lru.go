package lru

import "github.com/hashicorp/golang-lru"

type LRU struct {
	cache *lru.Cache
}

func New(size int) (*LRU, error) {
	c, err := lru.New(size)
	cache := &LRU{cache: c}
	return cache, err
}

func (c LRU) Add(key string, value interface{}) bool {
	return c.cache.Add(key, value)
}

func (c LRU) Get(key string) (value interface{}, ok bool) {
	return c.cache.Get(key)
}

func (c LRU) Peek(key string) (value interface{}, ok bool) {
	return c.cache.Peek(key)
}

func (c LRU) Keys() []string {
	// We get back []interface{} which we need as []string
	iKeys := c.cache.Keys()
	sKeys := make([]string, len(iKeys))
	for i := 0; i < len(sKeys); i++ {
		sKeys[i] = iKeys[i].(string)
	}
	return sKeys
}

func (c LRU) Contains(key string) bool {
	return c.cache.Contains(key)
}
