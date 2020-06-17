package gearbox

import (
	"container/list"
	"sync"
)

// Implementation of LRU caching using doubly linked list and tst

// Cache returns LRU cache
type Cache interface {
	Set(key string, value interface{})
	Get(key string) interface{}
}

// lruCache holds info used for caching internally
type lruCache struct {
	capacity int
	list     *list.List
	store    map[string]interface{}
	mutex    sync.RWMutex
}

// pair contains key and value of element
type pair struct {
	key   string
	value interface{}
}

// NewCache returns LRU cache
func NewCache(capacity int) Cache {
	// minimum is 1
	if capacity <= 0 {
		capacity = 1
	}

	return &lruCache{
		capacity: capacity,
		list:     new(list.List),
		store:    make(map[string]interface{}),
	}
}

// Get returns value of provided key if it's existing
func (c *lruCache) Get(key string) interface{} {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	// check if list node exists
	if node, ok := c.store[key].(*list.Element); ok {
		c.list.MoveToFront(node)

		return node.Value.(*pair).value
	}
	return nil
}

// Set adds a value to provided key in cache
func (c *lruCache) Set(key string, value interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// update the value if key is existing
	if node, ok := c.store[key].(*list.Element); ok {
		c.list.MoveToFront(node)

		node.Value.(*pair).value = value
		return
	}

	// remove last node if cache is full
	if c.list.Len() == c.capacity {
		lastNode := c.list.Back()

		// delete key's value
		delete(c.store, lastNode.Value.(*pair).key)

		c.list.Remove(lastNode)
	}

	c.store[key] = c.list.PushFront(&pair{
		key:   key,
		value: value,
	})
}
