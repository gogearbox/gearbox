package gearbox

// Basic Implementation for Ternary Search Tree (TST)

// tst returns Ternary Search Tree
type tst interface {
	Set(word string, value interface{})
	Get(word string) interface{}
}

// Ternary Search Tree node that holds a single character and value if there is
type tstNode struct {
	lower  *tstNode
	higher *tstNode
	equal  *tstNode
	key    byte
	value  interface{}
}

// newTST returns Ternary Search Tree
func newTST() tst {
	return &tstNode{}
}

// Set adds a value to provided string
func (t *tstNode) Set(word string, value interface{}) {
	if len(word) < 1 {
		return
	}
	t.insert(t, word, 0, value)
}

// Get gets the value of provided key if it's existing, otherwise returns nil
func (t *tstNode) Get(word string) interface{} {
	length := len(word)
	if length < 1 || t == nil {
		return nil
	}
	lastElm := length - 1

	n := t
	idx := 0
	key := word[idx]
	for n != nil {
		if key < n.key {
			n = n.lower
		} else if key > n.key {
			n = n.higher
		} else {
			if idx == lastElm {
				return n.value
			}

			idx++
			n = n.equal
			key = word[idx]
		}
	}
	return nil
}

// insert is an internal method for inserting a string with value in TST
func (t *tstNode) insert(n *tstNode, word string, index int, value interface{}) *tstNode {
	key := word[index]
	lastElm := len(word) - 1

	if n == nil {
		n = &tstNode{key: key}
	}

	if key < n.key {
		n.lower = t.insert(n.lower, word, index, value)
	} else if key > n.key {
		n.higher = t.insert(n.higher, word, index, value)
	} else {
		if index == lastElm {
			n.value = value
		} else {
			n.equal = t.insert(n.equal, word, index+1, value)
		}
	}

	return n
}
