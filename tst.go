package gearbox

// Basic Implementation for Ternary Search Tree (TST)

// TST holds a pointer to the root element of Ternary Search Tree
type TST struct {
	root    *tstNode
	counter uint32
}

type tstNode struct {
	lower  *tstNode
	higher *tstNode
	equal  *tstNode
	key    byte
	value  interface{}
}

// Set adds a value to provided key
func (t *TST) Set(word string, value interface{}) {
	if len(word) < 1 {
		return
	}
	t.root = t.insert(t.root, word, 0, value)
	t.counter++
}

// Get gets the value of provided key if it's existing, otherwise returns nil
func (t *TST) Get(word string) interface{} {
	length := len(word)
	if length < 1 {
		return nil
	}
	lastElm := length - 1

	n := t.root
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

// Count gets counter of current values
func (t *TST) Count() uint32 {
	return t.counter
}

func (t *TST) insert(n *tstNode, word string, index int, value interface{}) *tstNode {
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
