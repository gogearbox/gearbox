package gearbox

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"testing"

	"github.com/valyala/fasthttp"
)

// Mocked conn interface for testing
// https://golang.org/src/net/net.go#L113
type fakeConn struct {
	net.Conn
	r bytes.Buffer
	w bytes.Buffer
}

// Close closes the connection
func (c *fakeConn) Close() error {
	return nil
}

// Read reads data from the connection
func (c *fakeConn) Read(b []byte) (int, error) {
	return c.r.Read(b)
}

// Write writes data to the connection
func (c *fakeConn) Write(b []byte) (int, error) {
	return c.w.Write(b)
}

// startGearbox constructs routing tree and creates server
func startGearbox(gb *gearbox) {
	gb.constructRoutingTree()
	gb.httpServer = &fasthttp.Server{
		Handler:      gb.handler,
		Logger:       nil,
		LogAllErrors: false,
	}
}

// makeRequest makes an http request to http server and returns response or error
func makeRequest(request *http.Request, gb *gearbox) (*http.Response, error) {
	// Dump request to send it
	dumpRequest, err := httputil.DumpRequest(request, true)
	if err != nil {
		return nil, err
	}

	// Write request to the connection
	c := &fakeConn{}
	if _, err = c.r.Write(dumpRequest); err != nil {
		return nil, err
	}

	// Handling connection
	ch := make(chan error)
	go func() {
		ch <- gb.httpServer.ServeConn(c)
	}()

	select {
	case err = <-ch:
		if err != nil {
			return nil, err
		}
	}

	// Parse response
	buffer := bufio.NewReader(&c.w)
	resp, err := http.ReadResponse(buffer, request)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// handler just an empty handler
var handler = func(c *fasthttp.RequestCtx) {}

// unAuthorizedHandler sets status unauthorized in response
var unAuthorizedHandler = func(c *fasthttp.RequestCtx) {
	c.SetStatusCode(StatusUnauthorized)
}

// pingHandler returns string pong in response body
var pingHandler = func(c *fasthttp.RequestCtx) {
	c.Response.SetBodyString("pong")
}

// TestMethods tests creating gearbox instance, registering routes, making
// requests and getting proper responses
func TestMethods(t *testing.T) {
	// testing routes
	routes := []struct {
		method  string
		path    string
		handler func(*fasthttp.RequestCtx)
	}{
		{method: MethodGet, path: "/articles/search", handler: emptyHandler},
		{method: MethodHead, path: "/articles/test", handler: emptyHandler},
		{method: MethodPost, path: "/articles/204", handler: emptyHandler},
		{method: MethodPost, path: "/articles/205", handler: unAuthorizedHandler},
		{method: MethodGet, path: "/ping", handler: pingHandler},
		{method: MethodPut, path: "/posts", handler: emptyHandler},
		{method: MethodPatch, path: "/post/502", handler: emptyHandler},
		{method: MethodDelete, path: "/post/a23011a", handler: emptyHandler},
		{method: MethodConnect, path: "/user/204", handler: emptyHandler},
		{method: MethodOptions, path: "/user/204/setting", handler: emptyHandler},
		{method: MethodTrace, path: "/users/*", handler: emptyHandler},
		{method: MethodTrace, path: "/users/test", handler: emptyHandler},
	}

	// get instance of gearbox
	gb := new(gearbox)
	gb.registeredRoutes = make([]*routeInfo, 0)

	// register routes according to method
	for _, r := range routes {
		switch r.method {
		case MethodGet:
			gb.Get(r.path, r.handler)
		case MethodHead:
			gb.Head(r.path, r.handler)
		case MethodPost:
			gb.Post(r.path, r.handler)
		case MethodPut:
			gb.Put(r.path, r.handler)
		case MethodPatch:
			gb.Patch(r.path, r.handler)
		case MethodDelete:
			gb.Delete(r.path, r.handler)
		case MethodConnect:
			gb.Connect(r.path, r.handler)
		case MethodOptions:
			gb.Options(r.path, r.handler)
		case MethodTrace:
			gb.Trace(r.path, r.handler)
		}
	}

	// start serving
	startGearbox(gb)

	// Requests that will be tested
	testCases := []struct {
		method     string
		path       string
		statusCode int
		body       string
	}{
		{method: MethodGet, path: "/articles/search", statusCode: StatusOK},
		{method: MethodPost, path: "/articles/search", statusCode: StatusNotFound},
		{method: MethodGet, path: "/articles/searching", statusCode: StatusNotFound},
		{method: MethodHead, path: "/articles/test", statusCode: StatusOK},
		{method: MethodPost, path: "/articles/204", statusCode: StatusOK},
		{method: MethodPost, path: "/articles/205", statusCode: StatusUnauthorized},
		{method: MethodPost, path: "/articles/206", statusCode: StatusNotFound},
		{method: MethodGet, path: "/ping", statusCode: StatusOK, body: "pong"},
		{method: MethodPut, path: "/posts", statusCode: StatusOK},
		{method: MethodPatch, path: "/post/502", statusCode: StatusOK},
		{method: MethodDelete, path: "/post/a23011a", statusCode: StatusOK},
		{method: MethodConnect, path: "/user/204", statusCode: StatusOK},
		{method: MethodOptions, path: "/user/204/setting", statusCode: StatusOK},
		{method: MethodTrace, path: "/users/testing", statusCode: StatusOK},
	}

	for _, tc := range testCases {
		// create and make http request
		req, _ := http.NewRequest(tc.method, tc.path, nil)
		response, err := makeRequest(req, gb)

		if err != nil {
			t.Fatalf("%s(%s): %s", tc.method, tc.path, err.Error())
		}

		// check status code
		if response.StatusCode != tc.statusCode {
			t.Fatalf("%s(%s): returned %d expected %d", tc.method, tc.path, response.StatusCode, tc.statusCode)
		}

		// read body from response
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			t.Fatalf("%s(%s): %s", tc.method, tc.path, err.Error())
		}

		// check response body
		if string(body) != tc.body {
			t.Fatalf("%s(%s): returned %s expected %s", tc.method, tc.path, body, tc.body)
		}
	}
}

// TestStop tests stop service method
func TestStop(t *testing.T) {
	gearbox := New()
	gearbox.Stop()
}
