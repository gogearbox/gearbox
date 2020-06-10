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
		{input: []byte("/user/:name"), isErr: false},
		{input: []byte("/user/:name/:name"), isErr: true},
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
var emptyHandlersChain = handlersChain{}

// TestRegisterRoute tests registering routes after validating it
func TestRegisterRoute(t *testing.T) {
	// test cases
	tests := []struct {
		method  []byte
		path    []byte
		handler handlersChain
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
		{method: []byte(MethodGet), path: []byte("/books/:name/:test"), handler: emptyHandlersChain, isErr: false},
		{method: []byte(MethodGet), path: []byte("/books/:name/:name"), handler: nil, isErr: true},
	}

	// create gearbox instance
	gb := new(gearbox)
	gb.registeredRoutes = make([]*route, 0)

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
	gb.registeredRoutes = make([]*route, 0)

	// test handler is nil
	if err := gb.registerRoute([]byte(MethodGet), []byte("invalid Path"), emptyHandlersChain); err == nil {
		t.Errorf("input GET invalid Path find nil expecting error")
	}
}

// TestParseParameter tests parsing parameters into param struct
func TestParseParameter(t *testing.T) {
	tests := []struct {
		path   []byte
		output *param
	}{
		{path: []byte(":test"), output: &param{Name: []byte("test"), Type: ptParam}},
		{path: []byte(":test2:[a-z]"), output: &param{Name: []byte("test2"), Value: "[a-z]", Type: ptRegexp}},
		{path: []byte("*"), output: &param{Name: []byte("*"), Type: ptMatchAll}},
		{path: []byte("user:[a-z]"), output: nil},
		{path: []byte("user"), output: nil},
		{path: []byte(""), output: nil},
	}

	for _, test := range tests {
		p := parseParameter(test.path)
		if test.output == nil && p == nil {
			continue
		}
		if (test.output == nil && p != nil) ||
			(test.output != nil && p == nil) ||
			(string(test.output.Name) != string(p.Name)) ||
			(test.output.Type != p.Type) ||
			(test.output.Value != p.Value) {
			t.Errorf("path %s, find %v expected %v", test.path, p, test.output)
		}
	}
}

// TestGetLeastStrictParamType test
func TestGetLeastStrictParamType(t *testing.T) {
	tests := []struct {
		params []*param
		output paramType
	}{
		{params: []*param{}, output: ptNoParam},
		{params: []*param{
			{Type: ptParam, Name: []byte("name")},
			{Type: ptRegexp, Name: []byte("test"), Value: "[a-z]"},
			{Type: ptMatchAll, Name: []byte("*")},
		}, output: ptMatchAll},
		{params: []*param{
			{Type: ptParam, Name: []byte("name")},
			{Type: ptMatchAll, Name: []byte("*")},
		}, output: ptMatchAll},
		{params: []*param{
			{Type: ptParam, Name: []byte("name")},
			{Type: ptRegexp, Name: []byte("test3"), Value: "[a-z]"},
		}, output: ptParam},
		{params: []*param{
			{Type: ptRegexp, Name: []byte("test3"), Value: "[a-z]"},
			{Type: ptMatchAll, Name: []byte("*")},
		}, output: ptMatchAll},
	}

	for _, test := range tests {
		paramType := getLeastStrictParamType(test.params)
		if paramType != test.output {
			t.Errorf("params %v, find %d expected %d", test.params, paramType, test.output)
		}
	}
}

// TestTrimPath test
func TestTrimPath(t *testing.T) {
	tests := []struct {
		input  []byte
		output []byte
	}{
		{input: []byte("/"), output: []byte("")},
		{input: []byte("/test/"), output: []byte("test")},
		{input: []byte("test2/"), output: []byte("test2")},
		{input: []byte("test2"), output: []byte("test2")},
		{input: []byte("/user/test"), output: []byte("user/test")},
		{input: []byte("/books/get/test/"), output: []byte("books/get/test")},
	}

	for _, test := range tests {
		trimmedPath := trimPath(test.input)
		if string(trimmedPath) != string(test.output) {
			t.Errorf("path %s, find %s expected %s", test.input, trimmedPath, test.output)
		}
	}
}

// TestIsValidEndpoint test
func TestIsValidEndpoint(t *testing.T) {
	tests := []struct {
		endpoints   []*endpoint
		newEndpoint *endpoint
		output      bool
	}{
		{endpoints: []*endpoint{}, newEndpoint: &endpoint{}, output: true},
		{endpoints: []*endpoint{
			{Handlers: handlersChain{emptyHandler}, Params: []*param{
				{Name: []byte("user"), Type: ptParam},
				{Name: []byte("name"), Type: ptParam},
			}},
			{Handlers: handlersChain{emptyHandler}, Params: []*param{}},
		}, newEndpoint: &endpoint{Handlers: handlersChain{emptyHandler}, Params: []*param{
			{Name: []byte("test"), Type: ptParam},
		}}, output: true},
		{endpoints: []*endpoint{}, newEndpoint: &endpoint{}, output: true},
		{endpoints: []*endpoint{
			{Handlers: handlersChain{emptyHandler}, Params: []*param{
				{Name: []byte("user"), Type: ptParam},
			}},
			{Handlers: handlersChain{emptyHandler}, Params: []*param{}},
		}, newEndpoint: &endpoint{Handlers: handlersChain{emptyHandler}, Params: []*param{
			{Name: []byte("test"), Type: ptParam},
		}}, output: false},
		{endpoints: []*endpoint{
			{Handlers: handlersChain{emptyHandler}, Params: []*param{
				{Name: []byte("user"), Type: ptRegexp, Value: "[a-z]"},
			}},
			{Handlers: handlersChain{emptyHandler}, Params: []*param{}},
		}, newEndpoint: &endpoint{Handlers: handlersChain{emptyHandler}, Params: []*param{
			{Name: []byte("test"), Type: ptParam},
		}}, output: true},
		{endpoints: []*endpoint{
			{Handlers: handlersChain{emptyHandler}, Params: []*param{
				{Name: []byte("*"), Type: ptMatchAll},
			}},
			{Handlers: handlersChain{emptyHandler}, Params: []*param{}},
		}, newEndpoint: &endpoint{Handlers: handlersChain{emptyHandler}, Params: []*param{
			{Name: []byte("test"), Type: ptRegexp, Value: "[a-z]"},
		}}, output: true},
	}

	for _, test := range tests {
		isValid := isValidEndpoint(test.endpoints, test.newEndpoint)
		if isValid != test.output {
			t.Errorf("endpoints %v, new endpoint %v find %t expected %t", test.endpoints,
				test.newEndpoint, isValid, test.output)
		}
	}
}

// TestConstructRoutingTree tests constructing routing tree and matching routes properly
func TestConstructRoutingTree(t *testing.T) {
	// create gearbox instance
	gb := new(gearbox)
	gb.registeredRoutes = make([]*route, 0)

	// testing routes
	routes := []struct {
		method  []byte
		path    []byte
		handler handlersChain
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
		{method: []byte(MethodGet), path: []byte("/books/get/:name"), handler: emptyHandlersChain},
		{method: []byte(MethodGet), path: []byte("/books/get/*"), handler: emptyHandlersChain},
		{method: []byte(MethodGet), path: []byte("/books/search/:pattern:([a-z]+)"), handler: emptyHandlersChain},
		{method: []byte(MethodGet), path: []byte("/books/search/:pattern"), handler: emptyHandlersChain},
		{method: []byte(MethodGet), path: []byte("/books/search/:pattern1/:pattern2/:pattern3"), handler: emptyHandlersChain},
		{method: []byte(MethodGet), path: []byte("/books/search/*"), handler: emptyHandlersChain},
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
		params map[string]string
		match  bool
	}{
		{method: []byte(MethodPut), path: []byte("/admin/welcome"), match: false, params: make(map[string]string)},
		{method: []byte(MethodGet), path: []byte("/articles/search"), match: true, params: make(map[string]string)},
		{method: []byte(MethodGet), path: []byte("/articles/test"), match: true, params: make(map[string]string)},
		{method: []byte(MethodGet), path: []byte("/articles/204"), match: true, params: make(map[string]string)},
		{method: []byte(MethodGet), path: []byte("/posts"), match: true, params: make(map[string]string)},
		{method: []byte(MethodGet), path: []byte("/post/502"), match: true, params: make(map[string]string)},
		{method: []byte(MethodGet), path: []byte("/post/a23011a"), match: true, params: make(map[string]string)},
		{method: []byte(MethodPost), path: []byte("/post/a23011a"), match: false, params: make(map[string]string)},
		{method: []byte(MethodGet), path: []byte("/user/204"), match: true, params: make(map[string]string)},
		{method: []byte(MethodGet), path: []byte("/user/205"), match: true, params: make(map[string]string)},
		{method: []byte(MethodPost), path: []byte("/user/204/setting"), match: true, params: make(map[string]string)},
		{method: []byte(MethodGet), path: []byte("/users/ahmed"), match: true, params: make(map[string]string)},
		{method: []byte(MethodGet), path: []byte("/users/ahmed/ahmed"), match: true, params: make(map[string]string)},
		{method: []byte(MethodPut), path: []byte("/users/ahmed/ahmed"), match: false, params: make(map[string]string)},
		{method: []byte(MethodGet), path: []byte("/books/get/test"), match: true, params: map[string]string{"name": "test"}},
		{method: []byte(MethodGet), path: []byte("/books/search/test"), match: true, params: make(map[string]string)},
		{method: []byte(MethodGet), path: []byte("/books/search//test"), match: true, params: make(map[string]string)},
		{method: []byte(MethodGet), path: []byte("/books/search/123"), match: true, params: map[string]string{"pattern": "123"}},
		{method: []byte(MethodGet), path: []byte("/books/search/test1/test2/test3"), match: true, params: map[string]string{"pattern1": "test1", "pattern2": "test2", "pattern3": "test3"}},
		{method: []byte(MethodGet), path: []byte("/books/search/test/test2"), match: true, params: make(map[string]string)},
	}

	// test matching routes
	for _, rq := range requests {
		handler, params := gb.matchRoute(rq.method, rq.path)
		if (handler != nil && !rq.match) || (handler == nil && rq.match) {
			t.Errorf("input %s %s find nil expecting handler", rq.method, rq.path)
		}
		for paramKey, expectedParam := range rq.params {
			actualParam, ok := params.GetString(paramKey).([]byte)
			if !ok || string(actualParam) != expectedParam {
				if !ok {
					actualParam = []byte("nil")
				}

				t.Errorf("input %s %s parameter %s find %s expecting %s",
					rq.method, rq.path, paramKey, actualParam, expectedParam)
			}
		}
	}
}

// TestNullRoutingTree tests matching with null routing tree
func TestNullRoutingTree(t *testing.T) {
	// create gearbox instance
	gb := new(gearbox)
	gb.registeredRoutes = make([]*route, 0)

	// register route
	gb.registerRoute([]byte(MethodGet), []byte("/*"), emptyHandlersChain)

	// test handler is nil
	if handler, _ := gb.matchRoute([]byte(MethodGet), []byte("/hello/world")); handler != nil {
		t.Errorf("input GET /hello/world find handler expecting nil")
	}
}

// TestMatchAll tests matching all requests with one handler
func TestMatchAll(t *testing.T) {
	// create gearbox instance
	gb := new(gearbox)
	gb.registeredRoutes = make([]*route, 0)

	// register route
	gb.registerRoute([]byte(MethodGet), []byte("/*"), emptyHandlersChain)
	gb.constructRoutingTree()

	// test handler is not nil
	if handler, _ := gb.matchRoute([]byte(MethodGet), []byte("/hello/world")); handler == nil {
		t.Errorf("input GET /hello/world find nil expecting handler")
	}

	if handler, _ := gb.matchRoute([]byte(MethodGet), []byte("//world")); handler == nil {
		t.Errorf("input GET //world find nil expecting handler")
	}
}

// TestConstructRoutingTree tests constructing routing tree with two handlers
// for the same path and method
func TestConstructRoutingTreeConflict(t *testing.T) {
	// create gearbox instance
	gb := new(gearbox)
	gb.registeredRoutes = make([]*route, 0)

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
	gb.registeredRoutes = make([]*route, 0)

	// register routes
	gb.registerRoute([]byte(MethodGet), []byte("/articles"), emptyHandlersChain)
	gb.constructRoutingTree()

	// attempt to match route that cannot match
	if handler, _ := gb.matchRoute([]byte(MethodGet), []byte("/fail")); handler != nil {
		t.Errorf("input GET /fail found a valid handler, expecting nil")
	}
}

// TestFallback tests that if a registered fallback is available
// matchRoute() returns the non-nil registered fallback handler
func TestFallback(t *testing.T) {
	// create gearbox instance
	gb := new(gearbox)
	gb.registeredRoutes = make([]*route, 0)

	// register routes
	gb.registerRoute([]byte(MethodGet), []byte("/articles"), emptyHandlersChain)
	if err := gb.registerFallback(emptyHandlersChain); err != nil {
		t.Errorf("invalid fallback: %s", err.Error())
	}
	gb.constructRoutingTree()

	// attempt to match route that cannot match
	if handler, _ := gb.matchRoute([]byte(MethodGet), []byte("/fail")); handler == nil {
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
