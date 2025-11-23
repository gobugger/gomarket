package view

import (
	"sync"
)

const (
	CacheKeyProducts string = "products"
)

var cache = map[string]any{}
var mtx = sync.Mutex{}

func CachePut(key string, value any) {
	mtx.Lock()
	defer mtx.Unlock()
	cache[key] = value
}

func CacheGet(key string) any {
	mtx.Lock()
	defer mtx.Unlock()
	return cache[key]
}

func CacheDelete(key string) {
	mtx.Lock()
	defer mtx.Unlock()
	delete(cache, key)
}
