package gearbox

import (
	"testing"
)

// TestValidateRoutePath tests if provided paths are valid or not
func TestValidateRoutePath(t *testing.T) {
	// test cases
	tests := []struct {
		input []byte
		isErr bool
	}{
		{input: []byte(""), isErr: true},
		{input: []byte("user"), isErr: true},
		{input: []byte("/user"), isErr: false},
		{input: []byte("/admin/"), isErr: false},
		{input: []byte("/user/*/get"), isErr: true},
		{input: []byte("/user/*"), isErr: false},
	}

	for _, tt := range tests {
		err := validateRoutePath(tt.input)
		if (err != nil && !tt.isErr) || (err == nil && tt.isErr) {
			errMsg := ""

			// get error message if there is
			if err != nil {
				errMsg = err.Error()
			}

			t.Errorf("input %s find error %t %s expecting error %t", tt.input, err == nil, errMsg, tt.isErr)
		}
	}
}

// TestCreateEmptyNode tests creating route node with specific name
func TestCreateEmptyNode(t *testing.T) {
	name := []byte("test_node")
	node := createEmptyRouteNode(name)

	if node == nil {
		// node.Name != name {
		t.Errorf("find name %s expecting name %s", node.Name, name)
	}
}

// emptyHandler just an empty handler
var emptyHandler = func(ctx *Context) {}

// empty Handlers chain is just an empty array
var emptyHandlersChain = HandlersChain{}

// TestRegisterRoute tests registering routes after validating it
func TestRegisterRoute(t *testing.T) {
	// test cases
	tests := []struct {
		method  []byte
		path    []byte
		handler HandlersChain
		isErr   bool
	}{
		{method: []byte(MethodPut), path: []byte("/admin/welcome"), handler: emptyHandlersChain, isErr: false},
		{method: []byte(MethodPost), path: []byte("/user/add"), handler: emptyHandlersChain, isErr: false},
		{method: []byte(MethodGet), path: []byte("/account/get"), handler: emptyHandlersChain, isErr: false},
		{method: []byte(MethodGet), path: []byte("/account/*"), handler: emptyHandlersChain, isErr: false},
		{method: []byte(MethodGet), path: []byte("/account/*"), handler: emptyHandlersChain, isErr: false},
		{method: []byte(MethodDelete), path: []byte("/account/delete"), handler: emptyHandlersChain, isErr: false},
		{method: []byte(MethodDelete), path: []byte("/account/delete"), handler: nil, isErr: true},
		{method: []byte(MethodGet), path: []byte("/account/*/getAccount"), handler: nil, isErr: true},
	}

	// create gearbox instance
	gb := new(gearbox)
	gb.registeredRoutes = make([]*routeInfo, 0)

	// counter for valid routes
	validCounter := 0

	for _, tt := range tests {
		err := gb.registerRoute(tt.method, tt.path, tt.handler)
		if (err != nil && !tt.isErr) || (err == nil && tt.isErr) {
			errMsg := ""

			// get error message if there is
			if err != nil {
				errMsg = err.Error()
			}

			t.Errorf("input %v find error %t %s expecting error %t", tt, err == nil, errMsg, tt.isErr)
		}

		if !tt.isErr {
			validCounter++
		}
	}

	// check valid counter is the same as count of registered routes
	currentCount := len(gb.registeredRoutes)
	if validCounter != currentCount {
		t.Errorf("input %d find %d expecting %d", validCounter, currentCount, validCounter)
	}
}

// TestRegisterInvalidRoute tests registering invalid routes
func TestRegisterInvalidRoute(t *testing.T) {
	// create gearbox instance
	gb := new(gearbox)
	gb.registeredRoutes = make([]*routeInfo, 0)

	// test handler is nil
	if err := gb.registerRoute([]byte(MethodGet), []byte("invalid Path"), emptyHandlersChain); err == nil {
		t.Errorf("input GET invalid Path find nil expecting error")
	}
}

// TestConstructRoutingTree tests constructing routing tree and matching routes properly
func TestConstructRoutingTree(t *testing.T) {
	// create gearbox instance
	gb := new(gearbox)
	gb.registeredRoutes = make([]*routeInfo, 0)

	// testing routes
	routes := []struct {
		method  []byte
		path    []byte
		handler HandlersChain
	}{
		{method: []byte(MethodGet), path: []byte("/articles/search"), handler: emptyHandlersChain},
		{method: []byte(MethodGet), path: []byte("/articles/test"), handler: emptyHandlersChain},
		{method: []byte(MethodGet), path: []byte("/articles/204"), handler: emptyHandlersChain},
		{method: []byte(MethodGet), path: []byte("/posts"), handler: emptyHandlersChain},
		{method: []byte(MethodGet), path: []byte("/post/502"), handler: emptyHandlersChain},
		{method: []byte(MethodGet), path: []byte("/post/a23011a"), handler: emptyHandlersChain},
		{method: []byte(MethodGet), path: []byte("/user/204"), handler: emptyHandlersChain},
		{method: []byte(MethodGet), path: []byte("/user/205/"), handler: emptyHandlersChain},
		{method: []byte(MethodPost), path: []byte("/user/204/setting"), handler: emptyHandlersChain},
		{method: []byte(MethodGet), path: []byte("/users/*"), handler: emptyHandlersChain},
	}

	// register routes
	for _, r := range routes {
		gb.registerRoute(r.method, r.path, r.handler)
	}

	gb.constructRoutingTree()

	// requests test cases
	requests := []struct {
		method []byte
		path   []byte
		match  bool
	}{
		{method: []byte(MethodPut), path: []byte("/admin/welcome"), match: false},
		{method: []byte(MethodGet), path: []byte("/articles/search"), match: true},
		{method: []byte(MethodGet), path: []byte("/articles/test"), match: true},
		{method: []byte(MethodGet), path: []byte("/articles/204"), match: true},
		{method: []byte(MethodGet), path: []byte("/posts"), match: true},
		{method: []byte(MethodGet), path: []byte("/post/502"), match: true},
		{method: []byte(MethodGet), path: []byte("/post/a23011a"), match: true},
		{method: []byte(MethodPost), path: []byte("/post/a23011a"), match: false},
		{method: []byte(MethodGet), path: []byte("/user/204"), match: true},
		{method: []byte(MethodGet), path: []byte("/user/205"), match: true},
		{method: []byte(MethodPost), path: []byte("/user/204/setting"), match: true},
		{method: []byte(MethodGet), path: []byte("/users/ahmed"), match: true},
		{method: []byte(MethodGet), path: []byte("/users/ahmed/ahmed"), match: true},
		{method: []byte(MethodPut), path: []byte("/users/ahmed/ahmed"), match: false},
		{method: []byte(MethodPut), path: []byte(""), match: false},
	}

	// test matching routes
	for _, rq := range requests {
		handler := gb.matchRoute(rq.method, rq.path)
		if (handler != nil && !rq.match) || (handler == nil && rq.match) {
			t.Errorf("input %s %s find nil expecting handler", rq.method, rq.path)
		}
	}
}

// TestNullRoutingTree tests matching with null routing tree
func TestNullRoutingTree(t *testing.T) {
	// create gearbox instance
	gb := new(gearbox)
	gb.registeredRoutes = make([]*routeInfo, 0)

	// register route
	gb.registerRoute([]byte(MethodGet), []byte("/*"), emptyHandlersChain)

	// test handler is nil
	if handler := gb.matchRoute([]byte(MethodGet), []byte("/hello/world")); handler != nil {
		t.Errorf("input GET /hello/world find handler expecting nil")
	}
}

// TestMatchAll tests matching all requests with one handler
func TestMatchAll(t *testing.T) {
	// create gearbox instance
	gb := new(gearbox)
	gb.registeredRoutes = make([]*routeInfo, 0)

	// register route
	gb.registerRoute([]byte(MethodGet), []byte("/*"), emptyHandlersChain)
	gb.constructRoutingTree()

	// test handler is not nil
	if handler := gb.matchRoute([]byte(MethodGet), []byte("/hello/world")); handler == nil {
		t.Errorf("input GET /hello/world find nil expecting handler")
	}

	if handler := gb.matchRoute([]byte(MethodGet), []byte("//world")); handler == nil {
		t.Errorf("input GET //world find nil expecting handler")
	}
}

// TestConstructRoutingTree tests constructing routing tree with two handlers
// for the same path and method
func TestConstructRoutingTreeConflict(t *testing.T) {
	// create gearbox instance
	gb := new(gearbox)
	gb.registeredRoutes = make([]*routeInfo, 0)

	// register routes
	gb.registerRoute([]byte(MethodGet), []byte("/articles/test"), emptyHandlersChain)
	gb.registerRoute([]byte(MethodGet), []byte("/articles/test"), emptyHandlersChain)

	if err := gb.constructRoutingTree(); err == nil {
		t.Fatalf("invalid listener passed")
	}
}

// TestNoRegisteredFallback tests that if no registered fallback is available
// matchRoute() returns nil
func TestNoRegisteredFallback(t *testing.T) {
	// create gearbox instance
	gb := new(gearbox)
	gb.registeredRoutes = make([]*routeInfo, 0)

	// register routes
	gb.registerRoute([]byte(MethodGet), []byte("/articles"), emptyHandlersChain)
	gb.constructRoutingTree()

	// attempt to match route that cannot match
	if handler := gb.matchRoute([]byte(MethodGet), []byte("/fail")); handler != nil {
		t.Errorf("input GET /fail found a valid handler, expecting nil")
	}
}

// TestFallback tests that if a registered fallback is available
// matchRoute() returns the non-nil registered fallback handler
func TestFallback(t *testing.T) {
	// create gearbox instance
	gb := new(gearbox)
	gb.registeredRoutes = make([]*routeInfo, 0)

	// register routes
	gb.registerRoute([]byte(MethodGet), []byte("/articles"), emptyHandlersChain)
	if err := gb.registerFallback(emptyHandlersChain); err != nil {
		t.Errorf("invalid fallback: %s", err.Error())
	}
	gb.constructRoutingTree()

	// attempt to match route that cannot match
	if handler := gb.matchRoute([]byte(MethodGet), []byte("/fail")); handler == nil {
		t.Errorf("input GET /fail did not find a valid handler, expecting valid fallback handler")
	}
}

// TestInvalidFallback tests that a fallback cannot be registered
// with a nil handler
func TestInvalidFallback(t *testing.T) {
	// create gearbox instance
	gb := new(gearbox)

	// attempt to register an invalid (nil) fallback handler
	if err := gb.registerFallback(nil); err == nil {
		t.Errorf("registering an invalid fallback did not return an error, expected error")
	}
}
