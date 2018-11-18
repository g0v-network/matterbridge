package cache

type CacheInterface interface {
	Add(key string, value interface{}) bool
	Get(key string) (value interface{}, ok bool)
	Peek(key string) (value interface{}, ok bool)
	Keys() []string
	Contains(key string) bool
}
