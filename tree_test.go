package gearbox

import "testing"

func catchPanic(f func()) (recv interface{}) {
	defer func() {
		recv = recover()
	}()

	f()
	return
}

type testRoute struct {
	path     string
	conflict bool
}

func TestAddRoute(t *testing.T) {
	tree := createRootNode()

	routes := []testRoute{
		{"/cmd/:tool/:sub", false},
		{"/cmd/vet", true},
		{"/src/*", false},
		{"/src/*", true},
		{"/src/test", true},
		{"/src/:test", true},
		{"/src/", false},
		{"/src1/", false},
		{"/src1/*", false},
		{"/search/:query", false},
		{"/search/invalid", true},
		{"/user_:name", false},
		{"/user_x", false},
		{"/id:id", false},
		{"/id/:id", false},
		{"/id/:value", true},
		{"/id/:id/settings", false},
		{"/id/:id/:type", true},
		{"/*", true},
		{"books/*/get", true},
		{"/file/test", false},
		{"/file/test", true},
		{"/file/:test", true},
		{"/orders/:id/settings/:id", true},
		{"/accounts/*/settings", true},
		{"/results/*", false},
		{"/results/*/view", true},
	}
	for _, route := range routes {
		recv := catchPanic(func() {
			tree.addRoute(route.path, emptyHandlersChain)
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

type testRequests []struct {
	path   string
	match  bool
	params map[string]string
}

func TestMatchRoute(t *testing.T) {
	tree := createRootNode()

	routes := [...]string{
		"/hi",
		"/contact",
		"/users/:id/",
		"/books/*",
		"/search/:item1/settings/:item2",
		"/co",
		"/c",
		"/a",
		"/ab",
		"/doc/",
		"/doc/go_faq.html",
		"/doc/go1.html",
		"/α",
		"/β",
	}
	for _, route := range routes {
		tree.addRoute(route, emptyHandlersChain)
	}

	requests := testRequests{
		{"/a", true, nil},
		{"/", false, nil},
		{"/hi", true, nil},
		{"/contact", true, nil},
		{"/co", true, nil},
		{"/con", false, nil},  // key mismatch
		{"/cona", false, nil}, // key mismatch
		{"/no", false, nil},   // no matching child
		{"/ab", true, nil},
		{"/α", true, nil},
		{"/β", true, nil},
		{"/users/test", true, map[string]string{"id": "test"}},
		{"/books/title", true, nil},
		{"/search/test1/settings/test2", true, map[string]string{"item1": "test1", "item2": "test2"}},
		{"/search/test1", false, nil},
		{"test", false, nil},
	}
	for _, request := range requests {
		ctx := &context{paramValues: make(map[string]string)}
		handler := tree.matchRoute(request.path, ctx)

		if handler == nil {
			if request.match {
				t.Errorf("handle mismatch for route '%s': Expected non-nil handle", request.path)
			}
		} else if !request.match {
			t.Errorf("handle mismatch for route '%s': Expected nil handle", request.path)
		}

		for expectedKey, expectedValue := range request.params {
			actualValue := ctx.Param(expectedKey)
			if actualValue != expectedValue {
				t.Errorf(" mismatch for route '%s' parameter '%s' actual '%s', expected '%s'",
					request.path, expectedKey, actualValue, expectedValue)
			}
		}
	}
}
