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
	length := len(path)
	if length == 0 {
		return fmt.Errorf("length is zero")
	}

	if path[0] != '/' {
		return fmt.Errorf("path must start with /")
	}

	starIndex := strings.Index(path, "*")
	if starIndex > 0 && starIndex < length-1 && path[starIndex-1] == '/' {
		return fmt.Errorf("* must be in the end of path")
	}

	return nil
}

// registerRoute registers handler with method and path
func (gb *gearboxApp) registerRoute(method string, path string, handler func(*fasthttp.RequestCtx)) error {
	if handler == nil {
		return fmt.Errorf("route %s with method %s does not contain any handlers", path, method)
	}

	if err := validateRoutePath(path); err != nil {
		return fmt.Errorf("route %s is not valid! - %s", path, err.Error())
	}

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
func (gb *gearboxApp) constructRoutingTree() error {
	gb.routingTreeRoot = createEmptyRouteNode("root")
	for _, route := range gb.registeredRoutes {
		currentNode := gb.routingTreeRoot

		keywords := strings.Split(route.Path, "/")
		keywordsLen := len(keywords)
		for i := 1; i < keywordsLen; i++ {
			keyword := keywords[i]
			if keyword == "" || keyword == "*" {
				if keyword == "*" {
					currentNode.MatchAll = true
				}
				continue
			}

			keywordNode, ok := currentNode.Children.Get(keyword).(*routeNode)
			if !ok {
				keywordNode = createEmptyRouteNode(keyword)
				currentNode.Children.Set(keyword, keywordNode)
			}
			currentNode = keywordNode
		}

		if routeHandler := currentNode.Methods.Get(route.Method); routeHandler != nil {
			return fmt.Errorf("there already registered method %s for %s", route.Method, route.Path)
		}
		currentNode.Methods.Set(route.Method, route.Handler)
	}
	return nil
}

// matchRoute matches provided method and path with handler if it's existing
func (gb *gearboxApp) matchRoute(method string, path string) func(*fasthttp.RequestCtx) {
	currentNode := gb.routingTreeRoot
	var lastMatchAll *routeNode
	if currentNode.MatchAll {
		lastMatchAll = currentNode
	}

	keywords := strings.Split(path, "/")
	keywordsLen := len(keywords)
	for i := 1; i < keywordsLen; i++ {
		keyword := keywords[i]
		if keyword == "" {
			continue
		}

		if keywordNode, ok := currentNode.Children.Get(keyword).(*routeNode); ok {
			currentNode = keywordNode
			if currentNode.MatchAll {
				lastMatchAll = currentNode
			}
			continue
		}

		if lastMatchAll == nil {
			return nil
		}

		if routeHandler, ok := lastMatchAll.Methods.Get(method).(func(*fasthttp.RequestCtx)); ok {
			return routeHandler
		}

		return nil
	}

	if routeHandler, ok := currentNode.Methods.Get(method).(func(*fasthttp.RequestCtx)); ok {
		return routeHandler
	}
	return nil
}
