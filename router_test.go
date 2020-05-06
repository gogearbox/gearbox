package gearbox

import (
	"testing"

	"github.com/valyala/fasthttp"
)

func TestValidateRoutePath(t *testing.T) {
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
	}

	for _, tt := range tests {
		err := validateRoutePath(tt.input)
		if (err != nil && !tt.isErr) || (err == nil && tt.isErr) {
			errMsg := ""
			if err != nil {
				errMsg = err.Error()
			}
			t.Errorf("input %s find error %t %s expecting error %t", tt.input, err == nil, errMsg, tt.isErr)
		}
	}
}

func TestCreateEmptyNode(t *testing.T) {
	name := "test_node"
	node := createEmptyRouteNode(name)

	if node == nil || node.Name != name {
		t.Errorf("find name %s expecting name %s", node.Name, name)
	}
}

var emptyHandler = func(ctx *fasthttp.RequestCtx) {}

func TestRegisterRoute(t *testing.T) {

	tests := []struct {
		method  string
		path    string
		handler func(*fasthttp.RequestCtx)
		isErr   bool
	}{
		{method: MethodPut, path: "/admin/welcome", handler: emptyHandler, isErr: false},
		{method: MethodPost, path: "/user/add", handler: emptyHandler, isErr: false},
		{method: MethodGet, path: "/account/get", handler: emptyHandler, isErr: false},
		{method: MethodGet, path: "/account/*", handler: emptyHandler, isErr: false},
		{method: MethodDelete, path: "/account/delete", handler: emptyHandler, isErr: false},
		{method: MethodDelete, path: "/account/delete", handler: nil, isErr: true},
		{method: MethodGet, path: "/account/*/getAccount", handler: nil, isErr: true},
	}

	gearbox := new(gearboxApp)
	gearbox.registeredRoutes = make([]*routeInfo, 0)

	validCounter := 0
	for _, tt := range tests {
		err := gearbox.registerRoute(tt.method, tt.path, tt.handler)
		if (err != nil && !tt.isErr) || (err == nil && tt.isErr) {
			errMsg := ""
			if err != nil {
				errMsg = err.Error()
			}
			t.Errorf("input %v find error %t %s expecting error %t", tt, err == nil, errMsg, tt.isErr)
		}
		if !tt.isErr {
			validCounter++
		}
	}

	currentCount := len(gearbox.registeredRoutes)
	if validCounter != currentCount {
		t.Errorf("input %d find %d expecting %d", validCounter, currentCount, validCounter)
	}
}

func TestConstructRoutingTree(t *testing.T) {
	gearbox := new(gearboxApp)
	gearbox.registeredRoutes = make([]*routeInfo, 0)

	routes := []struct {
		method  string
		path    string
		handler func(*fasthttp.RequestCtx)
	}{
		{method: MethodGet, path: "/articles/search", handler: emptyHandler},
		{method: MethodGet, path: "/articles/test", handler: emptyHandler},
		{method: MethodGet, path: "/articles/204", handler: emptyHandler},
		{method: MethodGet, path: "/posts", handler: emptyHandler},
		{method: MethodGet, path: "/post/502", handler: emptyHandler},
		{method: MethodGet, path: "/post/a23011a", handler: emptyHandler},
		{method: MethodGet, path: "/user/204", handler: emptyHandler},
		{method: MethodPost, path: "/user/204/setting", handler: emptyHandler},
		{method: MethodGet, path: "/users/*", handler: emptyHandler},
	}

	for _, r := range routes {
		gearbox.registerRoute(r.method, r.path, r.handler)
	}

	gearbox.constructRoutingTree()

	for _, r := range routes {
		handler := gearbox.matchRoute(r.method, r.path)
		if handler == nil {
			t.Errorf("input %s %s find nil expecting handler", r.method, r.path)
		}
	}

	requests := []struct {
		method string
		path   string
		match  bool
	}{
		{method: MethodPut, path: "/admin/welcome", match: false},
		{method: MethodGet, path: "/articles/search", match: true},
		{method: MethodGet, path: "/articles/test", match: true},
		{method: MethodGet, path: "/articles/204", match: true},
		{method: MethodGet, path: "/posts", match: true},
		{method: MethodGet, path: "/post/502", match: true},
		{method: MethodGet, path: "/post/a23011a", match: true},
		{method: MethodPost, path: "/post/a23011a", match: false},
		{method: MethodGet, path: "/user/204", match: true},
		{method: MethodPost, path: "/user/204/setting", match: true},
		{method: MethodGet, path: "/users/ahmed", match: true},
		{method: MethodGet, path: "/users/ahmed/ahmed", match: true},
		{method: MethodPut, path: "/users/ahmed/ahmed", match: false},
	}

	for _, rq := range requests {
		handler := gearbox.matchRoute(rq.method, rq.path)
		if (handler != nil && !rq.match) || (handler == nil && rq.match) {
			t.Errorf("input %s %s find nil expecting handler", rq.method, rq.path)
		}
	}
}
