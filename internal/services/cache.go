package services

import (
	"container/list"
	"sync"
)

type LRUCache struct {
	capacity int
	cache    map[string]*list.Element
	list     *list.List
	mutex    sync.RWMutex
}

type entry struct {
	key   string
	value string
}

func NewLRUCache(capacity int) *LRUCache {
	return &LRUCache{
		capacity: capacity,
		cache:    make(map[string]*list.Element),
		list:     list.New(),
	}
}

func (c *LRUCache) Get(key string) (string, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	if elem, exists := c.cache[key]; exists {
		c.list.MoveToFront(elem)
		return elem.Value.(*entry).value, true
	}
	return "", false
}

func (c *LRUCache) Put(key, value string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	// Don't store anything if capacity is 0
	if c.capacity <= 0 {
		return
	}
	
	if elem, exists := c.cache[key]; exists {
		c.list.MoveToFront(elem)
		elem.Value.(*entry).value = value
		return
	}
	
	if c.list.Len() >= c.capacity {
		oldest := c.list.Back()
		if oldest != nil {
			c.list.Remove(oldest)
			delete(c.cache, oldest.Value.(*entry).key)
		}
	}
	
	newEntry := &entry{key: key, value: value}
	elem := c.list.PushFront(newEntry)
	c.cache[key] = elem
}