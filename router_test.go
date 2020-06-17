package gearbox

import (
	"fmt"
	"testing"
)

// setupGearbox returns instace of gearbox struct
func setupGearbox(settings ...*Settings) *gearbox {
	gb := new(gearbox)
	gb.registeredRoutes = make([]*Route, 0)

	if len(settings) > 0 {
		gb.settings = settings[0]
	} else {
		gb.settings = &Settings{}
	}

	gb.cache = NewCache(defaultCacheSize)
	return gb
}

// TestValidateRoutePath tests if provided paths are valid or not
func TestValidateRoutePath(t *testing.T) {
	// test cases
	tests := []struct {
		input string
		isErr bool
	}{
		{input: "", isErr: true},
		{input: "user", isErr: true},
		{input: "/user", isErr: false},
		{input: "/admin/", isErr: false},
		{input: "/user/*/get", isErr: true},
		{input: "/user/*", isErr: false},
		{input: "/user/:name", isErr: false},
		{input: "/user/:name/:name", isErr: true},
		{input: "/user/:name?/get", isErr: true},
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
	name := "test_node"
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
		method  string
		path    string
		handler handlersChain
		isErr   bool
	}{
		{method: MethodPut, path: "/admin/welcome", handler: emptyHandlersChain, isErr: false},
		{method: MethodPost, path: "/user/add", handler: emptyHandlersChain, isErr: false},
		{method: MethodGet, path: "/account/get", handler: emptyHandlersChain, isErr: false},
		{method: MethodGet, path: "/account/*", handler: emptyHandlersChain, isErr: false},
		{method: MethodGet, path: "/account/*", handler: emptyHandlersChain, isErr: false},
		{method: MethodDelete, path: "/account/delete", handler: emptyHandlersChain, isErr: false},
		{method: MethodDelete, path: "/account/delete", handler: nil, isErr: true},
		{method: MethodGet, path: "/account/*/getAccount", handler: nil, isErr: true},
		{method: MethodGet, path: "/books/:name/:test", handler: emptyHandlersChain, isErr: false},
		{method: MethodGet, path: "/books/:name/:name", handler: nil, isErr: true},
	}

	// counter for valid routes
	validCounter := 0

	for _, tt := range tests {
		// create gearbox instance so old errors don't affect new routes
		gb := setupGearbox()

		gb.registerRoute(tt.method, tt.path, tt.handler)
		err := gb.constructRoutingTree()
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
}

// TestRegisterInvalidRoute tests registering invalid routes
func TestRegisterInvalidRoute(t *testing.T) {
	// create gearbox instance
	gb := setupGearbox()

	// test handler is nil
	gb.registerRoute(MethodGet, "invalid Path", emptyHandlersChain)

	if err := gb.constructRoutingTree(); err == nil {
		t.Errorf("input GET invalid Path find nil expecting error")
	}
}

// TestParseParameter tests parsing parameters into param struct
func TestParseParameter(t *testing.T) {
	tests := []struct {
		path   string
		output *param
	}{
		{path: ":test", output: &param{Name: "test", Type: ptParam}},
		{path: ":test2:[a-z]", output: &param{Name: "test2", Value: "[a-z]", Type: ptRegexp}},
		{path: "*", output: &param{Name: "*", Type: ptMatchAll}},
		{path: "user:[a-z]", output: nil},
		{path: "user", output: nil},
		{path: "", output: nil},
	}

	for _, test := range tests {
		p := parseParameter(test.path)
		if test.output == nil && p == nil {
			continue
		}
		if (test.output == nil && p != nil) ||
			(test.output != nil && p == nil) ||
			(test.output.Name != p.Name) ||
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
			{Type: ptParam, Name: "name"},
			{Type: ptRegexp, Name: "test", Value: "[a-z]"},
			{Type: ptMatchAll, Name: "*"},
		}, output: ptMatchAll},
		{params: []*param{
			{Type: ptParam, Name: "name"},
			{Type: ptMatchAll, Name: "*"},
		}, output: ptMatchAll},
		{params: []*param{
			{Type: ptParam, Name: "name"},
			{Type: ptRegexp, Name: "test3", Value: "[a-z]"},
		}, output: ptParam},
		{params: []*param{
			{Type: ptRegexp, Name: "test3", Value: "[a-z]"},
			{Type: ptMatchAll, Name: "*"},
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
		input  string
		output string
	}{
		{input: "/", output: ""},
		{input: "/test/", output: "test"},
		{input: "test2/", output: "test2"},
		{input: "test2", output: "test2"},
		{input: "/user/test", output: "user/test"},
		{input: "/books/get/test/", output: "books/get/test"},
	}

	for _, test := range tests {
		trimmedPath := trimPath(test.input)
		if trimmedPath != test.output {
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
				{Name: "user", Type: ptParam},
				{Name: "name", Type: ptParam},
			}},
			{Handlers: handlersChain{emptyHandler}, Params: []*param{}},
		}, newEndpoint: &endpoint{Handlers: handlersChain{emptyHandler}, Params: []*param{
			{Name: "test", Type: ptParam},
		}}, output: true},
		{endpoints: []*endpoint{}, newEndpoint: &endpoint{}, output: true},
		{endpoints: []*endpoint{
			{Handlers: handlersChain{emptyHandler}, Params: []*param{
				{Name: "user", Type: ptParam},
			}},
			{Handlers: handlersChain{emptyHandler}, Params: []*param{}},
		}, newEndpoint: &endpoint{Handlers: handlersChain{emptyHandler}, Params: []*param{
			{Name: "test", Type: ptParam},
		}}, output: false},
		{endpoints: []*endpoint{
			{Handlers: handlersChain{emptyHandler}, Params: []*param{
				{Name: "user", Type: ptRegexp, Value: "[a-z]"},
			}},
			{Handlers: handlersChain{emptyHandler}, Params: []*param{}},
		}, newEndpoint: &endpoint{Handlers: handlersChain{emptyHandler}, Params: []*param{
			{Name: "test", Type: ptParam},
		}}, output: true},
		{endpoints: []*endpoint{
			{Handlers: handlersChain{emptyHandler}, Params: []*param{
				{Name: "*", Type: ptMatchAll},
			}},
			{Handlers: handlersChain{emptyHandler}, Params: []*param{}},
		}, newEndpoint: &endpoint{Handlers: handlersChain{emptyHandler}, Params: []*param{
			{Name: "test", Type: ptRegexp, Value: "[a-z]"},
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
	gb := setupGearbox(&Settings{
		CacheSize: 1,
	})

	// testing routes
	routes := []struct {
		method  string
		path    string
		handler handlersChain
	}{
		{method: MethodGet, path: "/articles/search", handler: emptyHandlersChain},
		{method: MethodGet, path: "/articles/test", handler: emptyHandlersChain},
		{method: MethodGet, path: "/articles/204", handler: emptyHandlersChain},
		{method: MethodGet, path: "/posts", handler: emptyHandlersChain},
		{method: MethodGet, path: "/post/502", handler: emptyHandlersChain},
		{method: MethodGet, path: "/post/a23011a", handler: emptyHandlersChain},
		{method: MethodGet, path: "/user/204", handler: emptyHandlersChain},
		{method: MethodGet, path: "/user/205/", handler: emptyHandlersChain},
		{method: MethodPost, path: "/user/204/setting", handler: emptyHandlersChain},
		{method: MethodGet, path: "/users/*", handler: emptyHandlersChain},
		{method: MethodGet, path: "/books/get/:name", handler: emptyHandlersChain},
		{method: MethodGet, path: "/books/get/*", handler: emptyHandlersChain},
		{method: MethodGet, path: "/books/search/:pattern:([a-z]+", handler: emptyHandlersChain},
		{method: MethodGet, path: "/books/search/:pattern", handler: emptyHandlersChain},
		{method: MethodGet, path: "/books/search/:pattern1/:pattern2/:pattern3", handler: emptyHandlersChain},
		{method: MethodGet, path: "/books//search/*", handler: emptyHandlersChain},
		{method: MethodGet, path: "/account/:name?", handler: emptyHandlersChain},
		{method: MethodGet, path: "/profile/:name:([a-z]+)?", handler: emptyHandlersChain},
		{method: MethodGet, path: "/order/:name1/:name2:([a-z]+)?", handler: emptyHandlersChain},
		{method: MethodGet, path: "/", handler: emptyHandlersChain},
	}

	// register routes
	for _, r := range routes {
		gb.registerRoute(r.method, r.path, r.handler)
	}

	gb.constructRoutingTree()

	// requests test cases
	requests := []struct {
		method string
		path   string
		params map[string]string
		match  bool
	}{
		{method: MethodPut, path: "/admin/welcome", match: false, params: make(map[string]string)},
		{method: MethodGet, path: "/articles/search", match: true, params: make(map[string]string)},
		{method: MethodGet, path: "/articles/test", match: true, params: make(map[string]string)},
		{method: MethodGet, path: "/articles/test", match: true, params: make(map[string]string)},
		{method: MethodGet, path: "/articles/test", match: true, params: make(map[string]string)},
		{method: MethodGet, path: "/articles/204", match: true, params: make(map[string]string)},
		{method: MethodGet, path: "/posts", match: true, params: make(map[string]string)},
		{method: MethodGet, path: "/post/502", match: true, params: make(map[string]string)},
		{method: MethodGet, path: "/post/a23011a", match: true, params: make(map[string]string)},
		{method: MethodPost, path: "/post/a23011a", match: false, params: make(map[string]string)},
		{method: MethodGet, path: "/user/204", match: true, params: make(map[string]string)},
		{method: MethodGet, path: "/user/205", match: true, params: make(map[string]string)},
		{method: MethodPost, path: "/user/204/setting", match: true, params: make(map[string]string)},
		{method: MethodGet, path: "/users/ahmed", match: true, params: make(map[string]string)},
		{method: MethodGet, path: "/users/ahmed/ahmed", match: true, params: make(map[string]string)},
		{method: MethodPut, path: "/users/ahmed/ahmed", match: false, params: make(map[string]string)},
		{method: MethodGet, path: "/books/get/test", match: true, params: map[string]string{"name": "test"}},
		{method: MethodGet, path: "/books/search/test", match: true, params: make(map[string]string)},
		{method: MethodGet, path: "/books/search//test", match: true, params: make(map[string]string)},
		{method: MethodGet, path: "/books/search/123", match: true, params: map[string]string{"pattern": "123"}},
		{method: MethodGet, path: "/books/search/test1/test2/test3", match: true, params: map[string]string{"pattern1": "test1", "pattern2": "test2", "pattern3": "test3"}},
		{method: MethodGet, path: "/books/search/test/test2", match: true, params: make(map[string]string)},
		{method: MethodGet, path: "/books/search/test/test2", match: true, params: make(map[string]string)},
		{method: MethodGet, path: "/account/testuser", match: true, params: map[string]string{"name": "testuser"}},
		{method: MethodGet, path: "/account", match: true, params: make(map[string]string)},
		{method: MethodPut, path: "/account/test1/test2", match: false, params: make(map[string]string)},
		{method: MethodGet, path: "/profile/testuser", match: true, params: map[string]string{"name": "testuser"}},
		{method: MethodGet, path: "/profile", match: true, params: make(map[string]string)},
		{method: MethodGet, path: "/order/test1", match: true, params: map[string]string{"name1": "test1"}},
		{method: MethodGet, path: "/order/test1/test2/", match: true, params: map[string]string{"name1": "test1", "name2": "test2"}},
		{method: MethodPut, path: "/order/test1/test2/test3", match: false, params: make(map[string]string)},
		{method: MethodGet, path: "/", match: true, params: make(map[string]string)},
	}

	// test matching routes
	for _, rq := range requests {
		handler, params := gb.matchRoute(rq.method, rq.path)
		if (handler != nil && !rq.match) || (handler == nil && rq.match) {
			t.Errorf("input %s %s find nil expecting handler", rq.method, rq.path)
		}
		for paramKey, expectedParam := range rq.params {
			if actualParam, ok := params[paramKey]; !ok || actualParam != expectedParam {
				if !ok {
					actualParam = "nil"
				}
				for k, w := range params {
					fmt.Println(k, string(w))
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
	gb := setupGearbox()

	// register route
	gb.registerRoute(MethodGet, "/*", emptyHandlersChain)

	// test handler is nil
	if handler, _ := gb.matchRoute(MethodGet, "/hello/world"); handler != nil {
		t.Errorf("input GET /hello/world find handler expecting nil")
	}
}

// TestMatchAll tests matching all requests with one handler
func TestMatchAll(t *testing.T) {
	// create gearbox instance
	gb := setupGearbox()

	// register route
	gb.registerRoute(MethodGet, "/*", emptyHandlersChain)
	gb.constructRoutingTree()

	// test handler is not nil
	if handler, _ := gb.matchRoute(MethodGet, "/hello/world"); handler == nil {
		t.Errorf("input GET /hello/world find nil expecting handler")
	}

	if handler, _ := gb.matchRoute(MethodGet, "//world"); handler == nil {
		t.Errorf("input GET //world find nil expecting handler")
	}
}

// TestConstructRoutingTree tests constructing routing tree with two handlers
// for the same path and method
func TestConstructRoutingTreeConflict(t *testing.T) {
	// create gearbox instance
	gb := setupGearbox()

	// register routes
	gb.registerRoute(MethodGet, "/articles/test", emptyHandlersChain)
	gb.registerRoute(MethodGet, "/articles/test", emptyHandlersChain)

	if err := gb.constructRoutingTree(); err == nil {
		t.Fatalf("invalid listener passed")
	}
}

// TestNoRegisteredFallback tests that if no registered fallback is available
// matchRoute() returns nil
func TestNoRegisteredFallback(t *testing.T) {
	// create gearbox instance
	gb := setupGearbox()

	// register routes
	gb.registerRoute(MethodGet, "/articles", emptyHandlersChain)
	gb.constructRoutingTree()

	// attempt to match route that cannot match
	if handler, _ := gb.matchRoute(MethodGet, "/fail"); handler != nil {
		t.Errorf("input GET /fail found a valid handler, expecting nil")
	}
}

// TestFallback tests that if a registered fallback is available
// matchRoute() returns the non-nil registered fallback handler
func TestFallback(t *testing.T) {
	// create gearbox instance
	gb := setupGearbox()

	// register routes
	gb.registerRoute(MethodGet, "/articles", emptyHandlersChain)
	if err := gb.registerFallback(emptyHandlersChain); err != nil {
		t.Errorf("invalid fallback: %s", err.Error())
	}
	gb.constructRoutingTree()

	// attempt to match route that cannot match
	if handler, _ := gb.matchRoute(MethodGet, "/fail"); handler == nil {
		t.Errorf("input GET /fail did not find a valid handler, expecting valid fallback handler")
	}
}

// TestInvalidFallback tests that a fallback cannot be registered
// with a nil handler
func TestInvalidFallback(t *testing.T) {
	// create gearbox instance
	gb := setupGearbox()

	// attempt to register an invalid (nil) fallback handler
	if err := gb.registerFallback(nil); err == nil {
		t.Errorf("registering an invalid fallback did not return an error, expected error")
	}
}

// TestGroupRouting tests that you can do group routing
func TestGroupRouting(t *testing.T) {
	// create gearbox instance
	gb := setupGearbox()
	routes := []*Route{gb.Get("/id", emptyHandler), gb.Post("/abc", emptyHandler), gb.Post("/abcd", emptyHandler)}
	gb.Group("/account", routes)
	// attempt to register an invalid (nil) fallback handler
	if err := gb.constructRoutingTree(); err != nil {
		t.Errorf("Grout routing failed, error: %v", err)
	}
}

// TestNestedGroupRouting tests that you can do group routing inside a group routing
func TestNestedGroupRouting(t *testing.T) {
	// create gearbox instance
	gb := setupGearbox()
	routes := []*Route{gb.Get("/id", emptyHandler), gb.Post("/abc", emptyHandler), gb.Post("/abcd", emptyHandler)}
	gb.Group("/account", gb.Group("/api", routes))
	// attempt to register an invalid (nil) fallback handler
	if err := gb.constructRoutingTree(); err != nil {
		t.Errorf("Grout routing failed, error: %v", err)
	}
}
