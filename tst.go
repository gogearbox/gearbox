package gearbox

import "sync"

// Basic Implementation for Ternary Search Tree (TST)

// tst returns Ternary Search Tree
type tst interface {
	Set(word []byte, value interface{})
	Get(word []byte) interface{}
	GetString(word string) interface{}
}

type tstImpl struct {
	root  *tstNode
	mutex sync.Mutex
}

// Ternary Search Tree node that holds a single character and value if there is
type tstNode struct {
	lower  *tstNode
	higher *tstNode
	equal  *tstNode
	char   byte
	value  interface{}
}

// newTST returns Ternary Search Tree
func newTST() tst {
	return &tstImpl{}
}

// Set adds a value to provided key
func (t *tstImpl) Set(key []byte, value interface{}) {
	if len(key) < 1 {
		return
	}

	t.mutex.Lock()
	t.root = t.insert(t.root, key, 0, value)
	t.mutex.Unlock()
}

// Get gets the value of provided key if it's existing, otherwise returns nil
func (t *tstImpl) Get(key []byte) interface{} {
	length := len(key)
	if length < 1 || t.root == nil {
		return nil
	}
	lastElm := length - 1

	n := t.root
	idx := 0
	char := key[idx]
	for n != nil {
		if char < n.char {
			n = n.lower
		} else if char > n.char {
			n = n.higher
		} else {
			if idx == lastElm {
				return n.value
			}

			idx++
			n = n.equal
			char = key[idx]
		}
	}
	return nil
}

// Get gets the value of provided key (string) if it's existing, otherwise returns nil
func (t *tstImpl) GetString(key string) interface{} {
	return t.Get([]byte(key))
}

// insert is an internal method for inserting a []byte with value in TST
func (t *tstImpl) insert(n *tstNode, key []byte, index int, value interface{}) *tstNode {
	char := key[index]
	lastElm := len(key) - 1

	if n == nil {
		n = &tstNode{char: char}
	}

	if char < n.char {
		n.lower = t.insert(n.lower, key, index, value)
	} else if char > n.char {
		n.higher = t.insert(n.higher, key, index, value)
	} else {
		if index == lastElm {
			n.value = value
		} else {
			n.equal = t.insert(n.equal, key, index+1, value)
		}
	}

	return n
}
