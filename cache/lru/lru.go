package lru

import "github.com/hashicorp/golang-lru"

type LRU struct {
	*lru.Cache
}

func New(size int) (*lru.Cache, error) {
	return lru.New(size)
}

func (cache LRU) Add(key, value interface{}) bool {
	return cache.Add(key, value)
}

func (cache LRU) Get(key interface{}) (value interface{}, ok bool) {
	return cache.Get(key)
}

func (cache LRU) Peek(key interface{}) (value interface{}, ok bool) {
	return cache.Peek(key)
}

func (cache LRU) Keys() []interface{} {
	return cache.Keys()
}

func (cache LRU) Contains(key interface{}) bool {
	return cache.Contains(key)
}
