package hw04lrucache

import "sync"

type Key string

type Cache interface {
	Set(key Key, value interface{}) bool
	Get(key Key) (interface{}, bool)
	Clear()
}

type lruCache struct {
	capacity int
	queue    List
	items    map[Key]*ListItem
	keys     map[*ListItem]Key
	mutex    sync.RWMutex
}

func (cache *lruCache) Get(key Key) (interface{}, bool) {
	cache.mutex.RLock()
	item, ok := cache.items[key]
	cache.mutex.RUnlock()
	if !ok {
		return nil, false
	}
	cache.mutex.Lock()
	cache.queue.MoveToFront(item)
	cache.mutex.Unlock()

	return item.Value, true
}

func (cache *lruCache) Set(key Key, value interface{}) bool {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	item, exists := cache.items[key]
	if exists {
		item.Value = value
		cache.queue.MoveToFront(item)
		return true
	}

	item = cache.queue.PushFront(value)
	cache.items[key] = item
	cache.keys[item] = key

	if cache.queue.Len() > cache.capacity {
		lastItem := cache.queue.Back()
		if lastItem != nil {
			if lastKey, ok := cache.keys[lastItem]; ok {
				delete(cache.items, lastKey)
				delete(cache.keys, lastItem)
				cache.queue.Remove(lastItem)
			}
		}
	}
	return false
}

func (cache *lruCache) Clear() {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()
	cache.items = make(map[Key]*ListItem)
	cache.queue = NewList()
	cache.keys = make(map[*ListItem]Key)
}

func NewCache(capacity int) Cache {
	return &lruCache{
		capacity: capacity,
		queue:    NewList(),
		items:    make(map[Key]*ListItem, capacity),
		keys:     make(map[*ListItem]Key, capacity),
	}
}
