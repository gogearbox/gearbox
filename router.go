package gearbox

import (
	"bytes"
	"fmt"
)

type routeNode struct {
	Name     []byte
	Method   []byte
	MatchAll bool
	Methods  tst
	Children tst
}

type routeInfo struct {
	Method   []byte
	Path     []byte
	Handlers HandlersChain
}

type routerFallback struct {
	Handlers HandlersChain
}

// validateRoutePath makes sure that path complies with path's rules
func validateRoutePath(path []byte) error {
	// Check length of the path
	length := len(path)
	if length == 0 {
		return fmt.Errorf("length is zero")
	}

	// Make sure path starts with /
	if path[0] != '/' {
		return fmt.Errorf("path must start with /")
	}

	// Make sure that star is in the end of path if it's existing
	starIndex := bytes.Index(path, []byte("*"))
	if starIndex > 0 && starIndex < length-1 && path[starIndex-1] == '/' {
		return fmt.Errorf("* must be in the end of path")
	}

	return nil
}

// registerRoute registers handler with method and path
func (gb *gearbox) registerRoute(method []byte, path []byte, handlers HandlersChain) error {
	// Handler is not provided
	if handlers == nil {
		return fmt.Errorf("route %s with method %s does not contain any handlers", path, method)
	}

	// Check if path is valid or not
	if err := validateRoutePath(path); err != nil {
		return fmt.Errorf("route %s is not valid! - %s", path, err.Error())
	}

	// Add route to registered routes
	gb.registeredRoutes = append(gb.registeredRoutes, &routeInfo{
		Path:     path,
		Method:   method,
		Handlers: handlers,
	})
	return nil
}

// registerFallback registers a single handler that will match only if all other routes fail to match
func (gb *gearbox) registerFallback(handlers HandlersChain) error {
	// Handler is not provided
	if handlers == nil {
		return fmt.Errorf("fallback does not contain a handler")
	}

	gb.registeredFallback = &routerFallback{Handlers: handlers}
	return nil
}

// createEmptyRouteNode creates a new route node with name
func createEmptyRouteNode(name []byte) *routeNode {
	return &routeNode{
		Name:     name,
		Children: newTST(),
		Methods:  newTST(),
		MatchAll: false,
	}
}

// constructRoutingTree constructs routing tree according to provided routes
func (gb *gearbox) constructRoutingTree() error {
	// Firstly, create root node
	gb.routingTreeRoot = createEmptyRouteNode([]byte("root"))

	for _, route := range gb.registeredRoutes {
		currentNode := gb.routingTreeRoot

		// Split path into slices of keywords
		keywords := bytes.Split(route.Path, []byte("/"))

		keywordsLen := len(keywords)
		for i := 1; i < keywordsLen; i++ {
			keyword := keywords[i]

			// Do not create node if keyword is empty
			if len(keyword) == 0 {
				continue
			}

			// Set MatchAll flag for current node
			if keyword[0] == '*' {
				currentNode.MatchAll = true
				continue
			}

			// Try to get a child of current node with keyword, otherwise
			//creates a new node and make it current node
			keywordNode, ok := currentNode.Children.Get(keyword).(*routeNode)
			if !ok {
				keywordNode = createEmptyRouteNode(keyword)
				currentNode.Children.Set(keyword, keywordNode)
			}
			currentNode = keywordNode
		}

		// Make sure that current node does not have a handler for route's method
		if routeHandler := currentNode.Methods.Get(route.Method); routeHandler != nil {
			return fmt.Errorf("there already registered method %s for %s", route.Method, route.Path)
		}

		// Save handler to route's method for current node
		currentNode.Methods.Set(route.Method, route.Handlers)
	}
	return nil
}

// matchRoute matches provided method and path with handler if it's existing
func (gb *gearbox) matchRoute(method []byte, path []byte) HandlersChain {
	if handlers := gb.matchRouteAgainstRegistered(method, path); handlers != nil {
		return handlers
	}

	if gb.registeredFallback != nil && gb.registeredFallback.Handlers != nil {
		return gb.registeredFallback.Handlers
	}

	return nil
}

// getKeywordEnd gets index of last byte before next '/' starting from index start
func getKeywordEnd(start int, path *[]byte, len int) int {
	for i := start; i < len; i++ {
		if (*path)[i] == '/' {
			return i
		}
	}
	return len
}

func (gb *gearbox) matchRouteAgainstRegistered(method []byte, path []byte) HandlersChain {
	// Start with root node
	currentNode := gb.routingTreeRoot

	// Return if root is empty
	if currentNode == nil || len(path) == 0 || path[0] != '/' {
		return nil
	}

	// Used to track if this path matched with match all node during matching
	// loop to fallback
	var lastMatchAll *routeNode
	if currentNode.MatchAll {
		lastMatchAll = currentNode
	}

	pathLen := len(path)
	lastPathByte := pathLen - 1

	// Start from the second byte as first is '/'
	start := 1
	for end := getKeywordEnd(start, &path, pathLen); start < lastPathByte; end = getKeywordEnd(start, &path, pathLen) {
		keyword := path[start:end]
		start = end + 1

		// Ignore empty keywords
		if len(keyword) == 0 {
			continue
		}

		// Try to match keyword with a child of current node
		if keywordNode, ok := currentNode.Children.Get(keyword).(*routeNode); ok {

			// set matched node as current node
			currentNode = keywordNode

			// Set lastMatchAll node if MatchAll flag is set in current node
			if currentNode.MatchAll {
				lastMatchAll = currentNode
			}
			continue
		}

		// There is no match for keyword with a child of current node
		// Check if lastMatchAll node is set to fallback
		if lastMatchAll == nil {
			return nil
		}

		// Try to get handler for provided method in lastMatchAll node and return it
		if routeHandler, ok := lastMatchAll.Methods.Get(method).(HandlersChain); ok {
			return routeHandler
		}

		// Otherwise, return nil since there is no match
		return nil
	}

	// Matching with path is done and trying get handler for provided method in
	// currentNode, otherwise return nil
	if routeHandler, ok := currentNode.Methods.Get(method).(HandlersChain); ok {
		return routeHandler
	}
	return nil
}
