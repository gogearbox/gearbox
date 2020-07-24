package gearbox

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

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

// setupGearbox returns instace of gearbox struct
func setupGearbox(settings ...*Settings) *gearbox {
	gb := new(gearbox)
	gb.registeredRoutes = make([]*Route, 0)

	if len(settings) > 0 {
		gb.settings = settings[0]
	} else {
		gb.settings = &Settings{}
	}

	gb.router = &router{
		settings: gb.settings,
		cache:    make(map[string]*matchResult),
		pool: sync.Pool{
			New: func() interface{} {
				return new(context)
			},
		},
	}

	return gb
}

// startGearbox constructs routing tree and creates server
func startGearbox(gb *gearbox) {
	gb.setupRouter()
	gb.httpServer = &fasthttp.Server{
		Handler:      gb.router.Handler,
		Logger:       nil,
		LogAllErrors: false,
	}
}

// emptyHandler just an empty handler
var emptyHandler = func(ctx Context) {}

// empty Handlers chain is just an empty array
var emptyHandlersChain = handlersChain{}

var fakeHandlersChain = handlersChain{emptyHandler}

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

	if err = <-ch; err != nil {
		return nil, err
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
var handler = func(ctx Context) {}

// errorHandler contains buggy code
var errorHandler = func(ctx Context) {
	m := make(map[string]int)
	m["a"] = 0
	ctx.SendString(string(5 / m["a"]))
}

// headerHandler echos header's value of key "my-header"
var headerHandler = func(ctx Context) {
	ctx.Set("custom", ctx.Get("my-header"))
}

// queryHandler answers with query's value of key "name"
var queryHandler = func(ctx Context) {
	ctx.SendString(ctx.Query("name"))
}

// bodyHandler answers with request body
var bodyHandler = func(ctx Context) {
	ctx.Context().Response.SetBodyString(ctx.Body())
}

// unAuthorizedHandler sets status unauthorized in response
var unAuthorizedHandler = func(ctx Context) {
	ctx.Status(StatusUnauthorized)
}

// pingHandler returns string pong in response body
var pingHandler = func(ctx Context) {
	ctx.SendString("pong")
}

// fallbackHandler returns not found status with custom fallback handler in response body
var fallbackHandler = func(ctx Context) {
	ctx.Status(StatusNotFound).SendString("custom fallback handler")
}

// emptyMiddleware does not stop the request and passes it to the next middleware/handler
var emptyMiddleware = func(ctx Context) {
	ctx.Next()
}

// registerRoute matches with register route request with available methods and calls it
func registerRoute(gb Gearbox, method, path string, handler func(ctx Context)) {
	switch method {
	case MethodGet:
		gb.Get(path, handler)
	case MethodHead:
		gb.Head(path, handler)
	case MethodPost:
		gb.Post(path, handler)
	case MethodPut:
		gb.Put(path, handler)
	case MethodPatch:
		gb.Patch(path, handler)
	case MethodDelete:
		gb.Delete(path, handler)
	case MethodConnect:
		gb.Connect(path, handler)
	case MethodOptions:
		gb.Options(path, handler)
	case MethodTrace:
		gb.Trace(path, handler)
	}
}

// TestMethods tests creating gearbox instance, registering routes, making
// requests and getting proper responses
func TestMethods(t *testing.T) {
	// testing routes
	routes := []struct {
		method  string
		path    string
		handler func(ctx Context)
	}{
		{method: MethodGet, path: "/order/get", handler: queryHandler},
		{method: MethodPost, path: "/order/add", handler: bodyHandler},
		{method: MethodGet, path: "/books/find", handler: emptyHandler},
		{method: MethodGet, path: "/articles/search", handler: emptyHandler},
		{method: MethodPut, path: "/articles/search", handler: emptyHandler},
		{method: MethodHead, path: "/articles/test", handler: emptyHandler},
		{method: MethodPost, path: "/articles/204", handler: emptyHandler},
		{method: MethodPost, path: "/articles/205", handler: unAuthorizedHandler},
		{method: MethodGet, path: "/ping", handler: pingHandler},
		{method: MethodPut, path: "/posts", handler: emptyHandler},
		{method: MethodPatch, path: "/post/502", handler: emptyHandler},
		{method: MethodDelete, path: "/post/a23011a", handler: emptyHandler},
		{method: MethodConnect, path: "/user/204", handler: headerHandler},
		{method: MethodOptions, path: "/user/204/setting", handler: errorHandler},
		{method: MethodTrace, path: "/users/*", handler: emptyHandler},
	}

	// get instance of gearbox
	gb := setupGearbox(&Settings{
		CaseInSensitive:        true,
		AutoRecover:            true,
		HandleOPTIONS:          true,
		HandleMethodNotAllowed: true,
	})

	// register routes according to method
	for _, r := range routes {
		registerRoute(gb, r.method, r.path, r.handler)
	}

	// start serving
	startGearbox(gb)

	// Requests that will be tested
	testCases := []struct {
		method      string
		path        string
		statusCode  int
		requestBody string
		body        string
		headers     map[string]string
	}{
		{method: MethodGet, path: "/order/get?name=art123", statusCode: StatusOK, body: "art123"},
		{method: MethodPost, path: "/order/add", requestBody: "testOrder", statusCode: StatusOK, body: "testOrder"},
		{method: MethodPost, path: "/books/find", statusCode: StatusMethodNotAllowed, body: "Method Not Allowed", headers: map[string]string{"Allow": "GET, OPTIONS"}},
		{method: MethodGet, path: "/articles/search", statusCode: StatusOK},
		{method: MethodGet, path: "/articles/search", statusCode: StatusOK},
		{method: MethodGet, path: "/Articles/search", statusCode: StatusOK},
		{method: MethodOptions, path: "/articles/search", statusCode: StatusOK},
		{method: MethodOptions, path: "*", statusCode: StatusOK},
		{method: MethodOptions, path: "/*", statusCode: StatusOK},
		{method: MethodGet, path: "/articles/searching", statusCode: StatusNotFound, body: "Not Found"},
		{method: MethodHead, path: "/articles/test", statusCode: StatusOK},
		{method: MethodPost, path: "/articles/204", statusCode: StatusOK},
		{method: MethodPost, path: "/articles/205", statusCode: StatusUnauthorized},
		{method: MethodPost, path: "/Articles/205", statusCode: StatusUnauthorized},
		{method: MethodPost, path: "/articles/206", statusCode: StatusNotFound, body: "Not Found"},
		{method: MethodGet, path: "/ping", statusCode: StatusOK, body: "pong"},
		{method: MethodPut, path: "/posts", statusCode: StatusOK},
		{method: MethodPatch, path: "/post/502", statusCode: StatusOK},
		{method: MethodDelete, path: "/post/a23011a", statusCode: StatusOK},
		{method: MethodConnect, path: "/user/204", statusCode: StatusOK, headers: map[string]string{"custom": "testing"}},
		{method: MethodOptions, path: "/user/204/setting", statusCode: StatusInternalServerError, body: "Internal Server Error"},
		{method: MethodTrace, path: "/users/testing", statusCode: StatusOK},
	}

	for _, tc := range testCases {
		// create and make http request
		req, _ := http.NewRequest(tc.method, tc.path, strings.NewReader(tc.requestBody))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Add("Content-Length", strconv.Itoa(len(tc.requestBody)))
		req.Header.Set("my-header", "testing")

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

		for expectedKey, expectedValue := range tc.headers {
			actualValue := response.Header.Get(expectedKey)
			if actualValue != expectedValue {
				t.Errorf(" mismatch for route '%s' parameter '%s' actual '%s', expected '%s'",
					tc.path, expectedKey, actualValue, expectedValue)
			}
		}
	}
}

// TestStartWithPrefork tests start service method
func TestStartWithPrefork(t *testing.T) {
	gb := New(&Settings{
		Prefork: true,
	})

	go func() {
		time.Sleep(1000 * time.Millisecond)
		gb.Stop()
	}()

	gb.Start(":3000")
}

// TestStart tests start service method
func TestStart(t *testing.T) {
	gb := New()

	go func() {
		time.Sleep(1000 * time.Millisecond)
		gb.Stop()
	}()

	gb.Start(":3010")
}

// TestStartWithTLS tests start service method
func TestStartWithTLS(t *testing.T) {
	gb := New(&Settings{
		TLSKeyPath:  "ssl-cert-snakeoil.key",
		TLSCertPath: "ssl-cert-snakeoil.crt",
		TLSEnabled:  true,
	})

	// use a channel to hand off the error ( if any )
	errs := make(chan error, 1)

	go func() {
		_, err := tls.DialWithDialer(
			&net.Dialer{
				Timeout: time.Second * 3,
			},
			"tcp",
			"localhost:3050",
			&tls.Config{
				InsecureSkipVerify: true,
			})
		errs <- err
		gb.Stop()
	}()

	gb.Start(":3050")

	// wait for an error
	err := <-errs
	if err != nil {
		t.Fatalf("StartWithSSL failed to connect with TLS error: %s", err)
	}
}

// TestStartInvalidListener tests start with invalid listener
func TestStartInvalidListener(t *testing.T) {
	gb := New()

	go func() {
		time.Sleep(1000 * time.Millisecond)
		gb.Stop()
	}()

	if err := gb.Start("invalid listener"); err == nil {
		t.Fatalf("invalid listener passed")
	}
}

// TestStop tests stop service method
func TestStop(t *testing.T) {
	gb := New()

	go func() {
		time.Sleep(1000 * time.Millisecond)
		gb.Stop()
	}()

	gb.Start("")
}

// TestRegisterFallback tests router fallback handler
func TestNotFound(t *testing.T) {
	// get instance of gearbox
	gb := setupGearbox()

	// register valid route
	gb.Get("/ping", pingHandler)

	// register not found handlers
	gb.NotFound(fallbackHandler)

	// start serving
	startGearbox(gb)

	// One valid request, one invalid
	testCases := []struct {
		method     string
		path       string
		statusCode int
		body       string
	}{
		{method: MethodGet, path: "/ping", statusCode: StatusOK, body: "pong"},
		{method: MethodGet, path: "/error", statusCode: StatusNotFound, body: "custom fallback handler"},
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

// TestGroupRouting tests that you can do group routing
func TestGroupRouting(t *testing.T) {
	// create gearbox instance
	gb := setupGearbox()
	routes := []*Route{
		gb.Get("/id", emptyHandler),
		gb.Post("/abc", emptyHandler),
		gb.Post("/abcd", emptyHandler),
	}
	gb.Group("/account", gb.Group("/api", routes))

	// start serving
	startGearbox(gb)

	// One valid request, one invalid
	testCases := []struct {
		method     string
		path       string
		statusCode int
		body       string
	}{
		{method: MethodGet, path: "/account/api/id", statusCode: StatusOK},
		{method: MethodPost, path: "/account/api/abc", statusCode: StatusOK},
		{method: MethodPost, path: "/account/api/abcd", statusCode: StatusOK},
		{method: MethodGet, path: "/id", statusCode: StatusNotFound, body: "Not Found"},
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

// TestUse tries to register middlewares that work before all routes
func TestUse(t *testing.T) {
	// get instance of gearbox
	gb := setupGearbox()

	// register valid route
	gb.Get("/ping", pingHandler)

	// Use authorized middleware for all the application
	gb.Use(unAuthorizedHandler)

	// start serving
	startGearbox(gb)

	req, _ := http.NewRequest(MethodGet, "/ping", nil)
	response, err := makeRequest(req, gb)

	if err != nil {
		t.Fatalf("%s(%s): %s", MethodGet, "/ping", err.Error())
	}

	// check status code
	if response.StatusCode != StatusUnauthorized {
		t.Fatalf("%s(%s): returned %d expected %d", MethodGet, "/ping", response.StatusCode, StatusUnauthorized)
	}
}
