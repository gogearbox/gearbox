package gearbox

import (
	"container/list"
	"sync"
)

// Implementation of LRU caching using doubly linked list and tst

// cache returns LRU cache
type cache interface {
	Set(key string, value interface{})
	Get(key string) interface{}
}

// lruCache holds info used for caching internally
type lruCache struct {
	capacity int
	list     *list.List
	store    sync.Map
	mutex    sync.RWMutex
}

// pair contains key and value of element
type pair struct {
	key   string
	value interface{}
}

// newCache returns LRU cache
func newCache(capacity int) cache {
	// minimum is 1
	if capacity <= 0 {
		capacity = 1
	}

	return &lruCache{
		capacity: capacity,
		list:     new(list.List),
	}
}

// Get returns value of provided key if it's existing
func (c *lruCache) Get(key string) interface{} {
	// check if list node exists
	if node, ok := c.store.Load(key); ok {
		nnode := node.(*list.Element)
		c.mutex.RLock()
		c.list.MoveToFront(nnode)
		c.mutex.RUnlock()

		return nnode.Value.(*pair).value
	}
	return nil
}

// Set adds a value to provided key in cache
func (c *lruCache) Set(key string, value interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// update the value if key is existing
	if node, ok := c.store.Load(key); ok {
		nnode := node.(*list.Element)
		c.list.MoveToFront(nnode)

		nnode.Value.(*pair).value = value

		return
	}

	// remove last node if cache is full
	if c.list.Len() == c.capacity {
		lastKey := c.list.Back().Value.(*pair).key

		// delete key's value
		c.store.Delete(lastKey)

		c.list.Remove(c.list.Back())
	}

	c.store.Store(key, c.list.PushFront(&pair{
		key:   key,
		value: value,
	}))
}
