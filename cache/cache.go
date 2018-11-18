package cache

type CacheInterface interface {
	Add(key, value interface{}) bool
	Get(key interface{}) (value interface{}, ok bool)
	Peek(key interface{}) (value interface{}, ok bool)
	Keys() []interface{}
	Contains(key interface{}) bool
}
