package gearbox

import (
	"net/http"
	"testing"
)

// Test passing the request from middleware to handler
func Test_Next(t *testing.T) {
	// testing routes
	routes := []struct {
		path       string
		middleware handlerFunc
	}{
		{path: "/ok", middleware: emptyMiddleware},
		{path: "/unauthorized", middleware: unAuthorizedHandler},
	}

	// get instance of gearbox
	gb := new(gearbox)
	gb.registeredRoutes = make([]*route, 0)
	gb.settings = &Settings{}

	// register routes according to method
	for _, r := range routes {
		gb.Get(r.path, r.middleware, emptyHandler)
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
