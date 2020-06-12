package gearbox

import (
	"container/list"
	"sync"
)

// Implementation of LRU caching using doubly linked list and tst

// cache returns LRU cache
type cache interface {
	Set(key []byte, value interface{})
	Get(key []byte) interface{}
}

// lruCache holds info used for caching internally
type lruCache struct {
	capacity int
	list     *list.List
	store    tst
	mutex    sync.RWMutex
}

// pair contains key and value of element
type pair struct {
	key   []byte
	value interface{}
}

// newCache returns LRU cache
func newCache(capacity int) cache {
	return &lruCache{
		capacity: capacity,
		list:     new(list.List),
		store:    newTST(),
	}
}

// Get returns value of provided key if it's existing
func (c *lruCache) Get(key []byte) interface{} {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// check if list node exists
	if node, ok := c.store.Get(key).(*list.Element); ok {
		c.list.MoveToFront(node)

		return node.Value.(*pair).value
	}
	return nil
}

// Set adds a value to provided key in cache
func (c *lruCache) Set(key []byte, value interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// update the value if key is existing
	if node, ok := c.store.Get(key).(*list.Element); ok {
		c.list.MoveToFront(node)
		node.Value.(*pair).value = value

		return
	}

	// remove last node if cache is full
	if c.list.Len() == c.capacity {
		lastKey := c.list.Back().Value.(*pair).key

		// delete key's value
		c.store.Set(lastKey, nil)

		c.list.Remove(c.list.Back())
	}

	newValue := &pair{
		key:   key,
		value: value,
	}
	c.store.Set(key, c.list.PushFront(newValue))
}
