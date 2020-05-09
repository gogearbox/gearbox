package gearbox

import (
	"fmt"
	"strings"

	"github.com/valyala/fasthttp"
)

type routeNode struct {
	Name     string
	Method   string
	MatchAll bool
	Methods  tst
	Children tst
}

type routeInfo struct {
	Method  string
	Path    string
	Handler func(*fasthttp.RequestCtx)
}

// validateRoutePath makes sure that path complies with path's rules
func validateRoutePath(path string) error {
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
	starIndex := strings.Index(path, "*")
	if starIndex > 0 && starIndex < length-1 && path[starIndex-1] == '/' {
		return fmt.Errorf("* must be in the end of path")
	}

	return nil
}

// registerRoute registers handler with method and path
func (gb *gearbox) registerRoute(method string, path string, handler func(*fasthttp.RequestCtx)) error {
	// Handler is not provided
	if handler == nil {
		return fmt.Errorf("route %s with method %s does not contain any handlers", path, method)
	}

	// Check if path is valid or not
	if err := validateRoutePath(path); err != nil {
		return fmt.Errorf("route %s is not valid! - %s", path, err.Error())
	}

	// Add route to registered routes
	gb.registeredRoutes = append(gb.registeredRoutes, &routeInfo{
		Path:    path,
		Method:  method,
		Handler: handler,
	})
	return nil
}

// createEmptyRouteNode creates a new route node with name
func createEmptyRouteNode(name string) *routeNode {
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
	gb.routingTreeRoot = createEmptyRouteNode("root")

	for _, route := range gb.registeredRoutes {
		currentNode := gb.routingTreeRoot

		// Split path into slices of keywords
		keywords := strings.Split(route.Path, "/")
		keywordsLen := len(keywords)
		for i := 1; i < keywordsLen; i++ {
			keyword := keywords[i]

			// Do not create node if keyword is empty or star
			if keyword == "" || keyword == "*" {
				// Set MatchAll flag for current node
				if keyword == "*" {
					currentNode.MatchAll = true
				}
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
		currentNode.Methods.Set(route.Method, route.Handler)
	}
	return nil
}

// matchRoute matches provided method and path with handler if it's existing
func (gb *gearbox) matchRoute(method string, path string) func(*fasthttp.RequestCtx) {
	// Start with root node
	currentNode := gb.routingTreeRoot

	// Return if root is empty
	if currentNode == nil {
		return nil
	}

	// Used to track if this path matched with match all node during matching
	// loop to fallback
	var lastMatchAll *routeNode
	if currentNode.MatchAll {
		lastMatchAll = currentNode
	}

	// Split path into slices of keywords
	keywords := strings.Split(path, "/")
	keywordsLen := len(keywords)
	for i := 1; i < keywordsLen; i++ {
		keyword := keywords[i]

		// Ignore empty keywords
		if keyword == "" {
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
		if routeHandler, ok := lastMatchAll.Methods.Get(method).(func(*fasthttp.RequestCtx)); ok {
			return routeHandler
		}

		// Otherwise, return nil since there is no match
		return nil
	}

	// Matching with path is done and trying get handler for provided method in
	// currentNode, otherwise return nil
	if routeHandler, ok := currentNode.Methods.Get(method).(func(*fasthttp.RequestCtx)); ok {
		return routeHandler
	}
	return nil
}
