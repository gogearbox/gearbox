package gearbox

import (
	"net/http"
	"testing"
)

// Test passing the request from middleware to handler
func TestNext(t *testing.T) {
	// testing routes
	routes := []struct {
		path       string
		middleware handlerFunc
		handler    handlerFunc
	}{
		{path: "/ok", middleware: emptyMiddleware, handler: emptyMiddlewareHandler},
		{path: "/unauthorized", middleware: unAuthorizedHandler, handler: emptyHandler},
	}

	// get instance of gearbox
	gb := setupGearbox()

	// register routes according to method
	for _, r := range routes {
		gb.Get(r.path, r.middleware, r.handler)
	}

	// start serving
	startGearbox(gb)

	// Requests that will be tested
	testCases := []struct {
		path       string
		statusCode int
	}{
		{path: "/ok", statusCode: StatusOK},
		{path: "/unauthorized", statusCode: StatusUnauthorized},
	}

	for _, tc := range testCases {
		// create and make http request
		req, _ := http.NewRequest(MethodGet, tc.path, nil)
		response, err := makeRequest(req, gb)

		if err != nil {
			t.Fatalf("%s(%s): %s", MethodGet, tc.path, err.Error())
		}

		// check status code
		if response.StatusCode != tc.statusCode {
			t.Fatalf("%s(%s): returned %d expected %d", MethodGet, tc.path, response.StatusCode, tc.statusCode)
		}
	}
}
