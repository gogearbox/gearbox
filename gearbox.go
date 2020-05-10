// Package gearbox is a web framework with a focus on high performance and memory optimization
package gearbox

import (
	"fmt"
	"log"
	"net"

	"github.com/valyala/fasthttp"
)

// Exported constants
const (
	Version = "0.0.2"   // Version of gearbox
	Name    = "Gearbox" // Name of gearbox
	// http://patorjk.com/software/taag/#p=display&f=Big%20Money-ne&t=Gearbox
	banner = `
  /$$$$$$                                /$$                          
 /$$__  $$                              | $$                          
| $$  \__/  /$$$$$$   /$$$$$$   /$$$$$$ | $$$$$$$   /$$$$$$  /$$   /$$
| $$ /$$$$ /$$__  $$ |____  $$ /$$__  $$| $$__  $$ /$$__  $$|  $$ /$$/
| $$|_  $$| $$$$$$$$  /$$$$$$$| $$  \__/| $$  \ $$| $$  \ $$ \  $$$$/ 
| $$  \ $$| $$_____/ /$$__  $$| $$      | $$  | $$| $$  | $$  >$$  $$ 
|  $$$$$$/|  $$$$$$$|  $$$$$$$| $$      | $$$$$$$/|  $$$$$$/ /$$/\  $$
 \______/  \_______/ \_______/|__/      |_______/  \______/ |__/  \__/  %s
Listening on %s
`
)

// HTTP methods were copied from net/http.
const (
	MethodGet     = "GET"     // RFC 7231, 4.3.1
	MethodHead    = "HEAD"    // RFC 7231, 4.3.2
	MethodPost    = "POST"    // RFC 7231, 4.3.3
	MethodPut     = "PUT"     // RFC 7231, 4.3.4
	MethodPatch   = "PATCH"   // RFC 5789
	MethodDelete  = "DELETE"  // RFC 7231, 4.3.5
	MethodConnect = "CONNECT" // RFC 7231, 4.3.6
	MethodOptions = "OPTIONS" // RFC 7231, 4.3.7
	MethodTrace   = "TRACE"   // RFC 7231, 4.3.8
)

// HTTP status codes were copied from net/http.
const (
	StatusContinue           = 100 // RFC 7231, 6.2.1
	StatusSwitchingProtocols = 101 // RFC 7231, 6.2.2
	StatusProcessing         = 102 // RFC 2518, 10.1

	StatusOK                   = 200 // RFC 7231, 6.3.1
	StatusCreated              = 201 // RFC 7231, 6.3.2
	StatusAccepted             = 202 // RFC 7231, 6.3.3
	StatusNonAuthoritativeInfo = 203 // RFC 7231, 6.3.4
	StatusNoContent            = 204 // RFC 7231, 6.3.5
	StatusResetContent         = 205 // RFC 7231, 6.3.6
	StatusPartialContent       = 206 // RFC 7233, 4.1
	StatusMultiStatus          = 207 // RFC 4918, 11.1
	StatusAlreadyReported      = 208 // RFC 5842, 7.1
	StatusIMUsed               = 226 // RFC 3229, 10.4.1

	StatusMultipleChoices   = 300 // RFC 7231, 6.4.1
	StatusMovedPermanently  = 301 // RFC 7231, 6.4.2
	StatusFound             = 302 // RFC 7231, 6.4.3
	StatusSeeOther          = 303 // RFC 7231, 6.4.4
	StatusNotModified       = 304 // RFC 7232, 4.1
	StatusUseProxy          = 305 // RFC 7231, 6.4.5
	_                       = 306 // RFC 7231, 6.4.6 (Unused)
	StatusTemporaryRedirect = 307 // RFC 7231, 6.4.7
	StatusPermanentRedirect = 308 // RFC 7538, 3

	StatusBadRequest                   = 400 // RFC 7231, 6.5.1
	StatusUnauthorized                 = 401 // RFC 7235, 3.1
	StatusPaymentRequired              = 402 // RFC 7231, 6.5.2
	StatusForbidden                    = 403 // RFC 7231, 6.5.3
	StatusNotFound                     = 404 // RFC 7231, 6.5.4
	StatusMethodNotAllowed             = 405 // RFC 7231, 6.5.5
	StatusNotAcceptable                = 406 // RFC 7231, 6.5.6
	StatusProxyAuthRequired            = 407 // RFC 7235, 3.2
	StatusRequestTimeout               = 408 // RFC 7231, 6.5.7
	StatusConflict                     = 409 // RFC 7231, 6.5.8
	StatusGone                         = 410 // RFC 7231, 6.5.9
	StatusLengthRequired               = 411 // RFC 7231, 6.5.10
	StatusPreconditionFailed           = 412 // RFC 7232, 4.2
	StatusRequestEntityTooLarge        = 413 // RFC 7231, 6.5.11
	StatusRequestURITooLong            = 414 // RFC 7231, 6.5.12
	StatusUnsupportedMediaType         = 415 // RFC 7231, 6.5.13
	StatusRequestedRangeNotSatisfiable = 416 // RFC 7233, 4.4
	StatusExpectationFailed            = 417 // RFC 7231, 6.5.14
	StatusTeapot                       = 418 // RFC 7168, 2.3.3
	StatusUnprocessableEntity          = 422 // RFC 4918, 11.2
	StatusLocked                       = 423 // RFC 4918, 11.3
	StatusFailedDependency             = 424 // RFC 4918, 11.4
	StatusUpgradeRequired              = 426 // RFC 7231, 6.5.15
	StatusPreconditionRequired         = 428 // RFC 6585, 3
	StatusTooManyRequests              = 429 // RFC 6585, 4
	StatusRequestHeaderFieldsTooLarge  = 431 // RFC 6585, 5
	StatusUnavailableForLegalReasons   = 451 // RFC 7725, 3

	StatusInternalServerError           = 500 // RFC 7231, 6.6.1
	StatusNotImplemented                = 501 // RFC 7231, 6.6.2
	StatusBadGateway                    = 502 // RFC 7231, 6.6.3
	StatusServiceUnavailable            = 503 // RFC 7231, 6.6.4
	StatusGatewayTimeout                = 504 // RFC 7231, 6.6.5
	StatusHTTPVersionNotSupported       = 505 // RFC 7231, 6.6.6
	StatusVariantAlsoNegotiates         = 506 // RFC 2295, 8.1
	StatusInsufficientStorage           = 507 // RFC 4918, 11.5
	StatusLoopDetected                  = 508 // RFC 5842, 7.2
	StatusNotExtended                   = 510 // RFC 2774, 7
	StatusNetworkAuthenticationRequired = 511 // RFC 6585, 6
)

// Gearbox interface
type Gearbox interface {
	Start(address string) error
	Stop() error
	Get(path string, handler func(*fasthttp.RequestCtx)) error
	Head(path string, handler func(*fasthttp.RequestCtx)) error
	Post(path string, handler func(*fasthttp.RequestCtx)) error
	Put(path string, handler func(*fasthttp.RequestCtx)) error
	Patch(path string, handler func(*fasthttp.RequestCtx)) error
	Delete(path string, handler func(*fasthttp.RequestCtx)) error
	Connect(path string, handler func(*fasthttp.RequestCtx)) error
	Options(path string, handler func(*fasthttp.RequestCtx)) error
	Trace(path string, handler func(*fasthttp.RequestCtx)) error
	Fallback(handler func(*fasthttp.RequestCtx)) error
}

// gearbox implements Gearbox interface
type gearbox struct {
	httpServer         *fasthttp.Server
	routingTreeRoot    *routeNode
	registeredRoutes   []*routeInfo
	address            string // server address
	registeredFallback *routerFallback
}

// New creates a new instance of gearbox
func New() Gearbox {
	gb := new(gearbox)
	gb.registeredRoutes = make([]*routeInfo, 0)
	gb.httpServer = gb.newHTTPServer()
	return gb
}

// Start handling requests
func (gb *gearbox) Start(address string) error {
	// Construct routing tree
	if err := gb.constructRoutingTree(); err != nil {
		return fmt.Errorf("Unable to construct routing %s", err.Error())
	}

	ln, err := net.Listen("tcp4", address)
	if err != nil {
		return err
	}
	gb.address = address
	err = gb.httpServer.Serve(ln)
	if err == nil {
		log.Printf(banner, Version, gb.address)
	}
	return err
}

// newHTTPServer returns a new instance of fasthttp server
func (gb *gearbox) newHTTPServer() *fasthttp.Server {
	return &fasthttp.Server{
		Handler:      gb.handler,
		Logger:       nil,
		LogAllErrors: false,
	}
}

// Stop serving
func (gb *gearbox) Stop() error {
	err := gb.httpServer.Shutdown()
	if err == nil && gb.address != "" { // check if shutdown was ok and server had valid address
		log.Printf("%s stopped listening on %s", Name, gb.address)
		return nil
	}
	return err
}

// Get registers an http relevant method
func (gb *gearbox) Get(path string, handler func(*fasthttp.RequestCtx)) error {
	return gb.registerRoute(MethodGet, path, handler)
}

// Head registers an http relevant method
func (gb *gearbox) Head(path string, handler func(*fasthttp.RequestCtx)) error {
	return gb.registerRoute(MethodHead, path, handler)
}

// Post registers an http relevant method
func (gb *gearbox) Post(path string, handler func(*fasthttp.RequestCtx)) error {
	return gb.registerRoute(MethodPost, path, handler)
}

// Put registers an http relevant method
func (gb *gearbox) Put(path string, handler func(*fasthttp.RequestCtx)) error {
	return gb.registerRoute(MethodPut, path, handler)
}

// Patch registers an http relevant method
func (gb *gearbox) Patch(path string, handler func(*fasthttp.RequestCtx)) error {
	return gb.registerRoute(MethodPatch, path, handler)
}

// Delete registers an http relevant method
func (gb *gearbox) Delete(path string, handler func(*fasthttp.RequestCtx)) error {
	return gb.registerRoute(MethodDelete, path, handler)
}

// Connect registers an http relevant method
func (gb *gearbox) Connect(path string, handler func(*fasthttp.RequestCtx)) error {
	return gb.registerRoute(MethodConnect, path, handler)
}

// Options registers an http relevant method
func (gb *gearbox) Options(path string, handler func(*fasthttp.RequestCtx)) error {
	return gb.registerRoute(MethodOptions, path, handler)
}

// Trace registers an http relevant method
func (gb *gearbox) Trace(path string, handler func(*fasthttp.RequestCtx)) error {
	return gb.registerRoute(MethodTrace, path, handler)
}

// Fallback registers an http handler only fired when no other routes match with request
func (gb *gearbox) Fallback(handler func(*fasthttp.RequestCtx)) error {
	return gb.registerFallback(handler)
}

// Handles all incoming requests and route them to proper handler according to
// method and path
func (gb *gearbox) handler(ctx *fasthttp.RequestCtx) {
	if handler := gb.matchRoute(getString(ctx.Request.Header.Method()), getString(ctx.URI().Path())); handler != nil {
		handler(ctx)
		return
	}

	ctx.SetStatusCode(StatusNotFound)
}
