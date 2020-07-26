package gearbox

import (
	"sync"
	"testing"
)

func TestHandle(t *testing.T) {
	// testing routes
	routes := []struct {
		method   string
		path     string
		conflict bool
		handlers handlersChain
	}{
		{method: MethodGet, path: "/articles/search", conflict: false, handlers: fakeHandlersChain},
		{method: MethodGet, path: "/articles/test", conflict: false, handlers: fakeHandlersChain},
		{method: MethodGet, path: "", conflict: true, handlers: fakeHandlersChain},
		{method: "", path: "/articles/test", conflict: true, handlers: fakeHandlersChain},
		{method: MethodGet, path: "orders/test", conflict: true, handlers: fakeHandlersChain},
		{method: MethodGet, path: "/books/test", conflict: true, handlers: emptyHandlersChain},
	}

	router := &router{
		settings: &Settings{},
		cache:    make(map[string]*matchResult),
		pool: sync.Pool{
			New: func() interface{} {
				return new(context)
			},
		},
	}

	for _, route := range routes {
		recv := catchPanic(func() {
			router.handle(route.method, route.path, route.handlers)
		})

		if route.conflict {
			if recv == nil {
				t.Errorf("no panic for conflicting route '%s'", route.path)
			}
		} else if recv != nil {
			t.Errorf("unexpected panic for route '%s': %v", route.path, recv)
		}
	}

}

func TestHandler(t *testing.T) {
	routes := []struct {
		method   string
		path     string
		handlers handlersChain
	}{
		{method: MethodGet, path: "/articles/search", handlers: fakeHandlersChain},
		{method: MethodGet, path: "/articles/test", handlers: fakeHandlersChain},
	}

	router := &router{
		settings: &Settings{},
		cache:    make(map[string]*matchResult),
		pool: sync.Pool{
			New: func() interface{} {
				return new(context)
			},
		},
	}

	for _, route := range routes {
		router.handle(route.method, route.path, route.handlers)
	}

}
