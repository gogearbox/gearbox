// Package gearbox is a web framework with a focus on high performance and memory optimization
package gearbox

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/prefork"
)

// Exported constants
const (
	Version = "1.0.3"   // Version of gearbox
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

const (
	// defaultCacheSize is number of entries that can cache hold
	defaultCacheSize = 1000

	// defaultConcurrency is the maximum number of concurrent connections
	defaultConcurrency = 256 * 1024

	// defaultMaxRequestBodySize is the maximum request body size the server
	defaultMaxRequestBodySize = 4 * 1024 * 1024

	// defaultMaxRequestParamsCount is the maximum number of request params
	defaultMaxRequestParamsCount = 1024

	// defaultMaxRequestURLLength is the maximum request url length
	defaultMaxRequestURLLength = 2048
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
	Get(path string, handlers ...handlerFunc) *Route
	Head(path string, handlers ...handlerFunc) *Route
	Post(path string, handlers ...handlerFunc) *Route
	Put(path string, handlers ...handlerFunc) *Route
	Patch(path string, handlers ...handlerFunc) *Route
	Delete(path string, handlers ...handlerFunc) *Route
	Connect(path string, handlers ...handlerFunc) *Route
	Options(path string, handlers ...handlerFunc) *Route
	Trace(path string, handlers ...handlerFunc) *Route
	Method(method, path string, handlers ...handlerFunc) *Route
	Fallback(handlers ...handlerFunc) error
	Use(middlewares ...handlerFunc)
	Group(prefix string, routes []*Route) []*Route
}

// gearbox implements Gearbox interface
type gearbox struct {
	httpServer         *fasthttp.Server
	routingTreeRoot    *routeNode
	registeredRoutes   []*Route
	address            string // server address
	handlers           handlersChain
	registeredFallback *routerFallback
	cache              Cache
	settings           *Settings
}

// Settings struct holds server settings
type Settings struct {
	// Enable case sensitive routing
	CaseSensitive bool // default false

	// Maximum size of LRU cache that will be used in routing if it's enabled
	CacheSize int // default 1000

	// ServerName for sending in response headers
	ServerName string // default ""

	// Maximum request body size
	MaxRequestBodySize int // default 4 * 1024 * 1024

	// Maximum request params count
	MaxRequestParamsCount int // default 1024

	// Max request url length
	MaxRequestURLLength int // default 2048

	// Maximum number of concurrent connections
	Concurrency int // default 256 * 1024

	// This will spawn multiple Go processes listening on the same port
	// Default: false
	Prefork bool

	// LRU caching used to speed up routing
	DisableCaching bool // default false

	// Disable printing gearbox banner
	DisableStartupMessage bool // default false

	// Disable keep-alive connections, the server will close incoming connections after sending the first response to client
	DisableKeepalive bool // default false

	// When set to true causes the default date header to be excluded from the response
	DisableDefaultDate bool // default false

	// When set to true, causes the default Content-Type header to be excluded from the Response
	DisableDefaultContentType bool // default false

	// By default all header names are normalized: conteNT-tYPE -> Content-Type
	DisableHeaderNormalizing bool // default false

	// The amount of time allowed to read the full request including body
	ReadTimeout time.Duration // default unlimited

	// The maximum duration before timing out writes of the response
	WriteTimeout time.Duration // default unlimited

	// The maximum amount of time to wait for the next request when keep-alive is enabled
	IdleTimeout time.Duration // default unlimited

	// Enable TLS or not
	TLSEnabled bool // default false

	// The path of the TLS certificate
	TLSCertPath string // default ""

	// The path of the TLS key
	TLSKeyPath string // default ""
}

// Route struct which holds each route info
type Route struct {
	Method   string
	Path     string
	Handlers handlersChain
}

// New creates a new instance of gearbox
func New(settings ...*Settings) Gearbox {
	gb := new(gearbox)
	gb.registeredRoutes = make([]*Route, 0)

	if len(settings) > 0 {
		gb.settings = settings[0]
	} else {
		gb.settings = &Settings{}
	}

	// set default settings for settings that don't have values set
	if gb.settings.CacheSize <= 0 {
		gb.settings.CacheSize = defaultCacheSize
	}

	if gb.settings.MaxRequestBodySize <= 0 {
		gb.settings.MaxRequestBodySize = defaultMaxRequestBodySize
	}

	if gb.settings.MaxRequestParamsCount <= 0 {
		gb.settings.MaxRequestParamsCount = defaultMaxRequestParamsCount
	}

	if gb.settings.MaxRequestURLLength <= 0 || gb.settings.MaxRequestURLLength > defaultMaxRequestURLLength {
		gb.settings.MaxRequestURLLength = defaultMaxRequestURLLength
	}

	if gb.settings.Concurrency <= 0 {
		gb.settings.Concurrency = defaultConcurrency
	}

	gb.httpServer = gb.newHTTPServer()

	return gb
}

// Start handling requests
func (gb *gearbox) Start(address string) error {
	// Construct routing tree
	if err := gb.constructRoutingTree(); err != nil {
		return fmt.Errorf("unable to construct routing %s", err.Error())
	}

	gb.cache = NewCache(gb.settings.CacheSize)

	if gb.settings.Prefork {
		if !gb.settings.DisableStartupMessage {
			printStartupMessage(address)
		}

		pf := prefork.New(gb.httpServer)
		pf.Reuseport = true
		pf.Network = "tcp4"

		if gb.settings.TLSEnabled {
			return pf.ListenAndServeTLS(address, gb.settings.TLSCertPath, gb.settings.TLSKeyPath)
		}
		return pf.ListenAndServe(address)
	}

	ln, err := net.Listen("tcp4", address)
	if err != nil {
		return err
	}
	gb.address = address

	if !gb.settings.DisableStartupMessage {
		printStartupMessage(address)
	}

	if gb.settings.TLSEnabled {
		return gb.httpServer.ServeTLS(ln, gb.settings.TLSCertPath, gb.settings.TLSKeyPath)
	}
	return gb.httpServer.Serve(ln)
}

// customLogger Customized logger used to filter logging messages
type customLogger struct{}

func (dl *customLogger) Printf(format string, args ...interface{}) {
}

// newHTTPServer returns a new instance of fasthttp server
func (gb *gearbox) newHTTPServer() *fasthttp.Server {
	return &fasthttp.Server{
		Handler:                       gb.handler,
		Logger:                        &customLogger{},
		LogAllErrors:                  false,
		Name:                          gb.settings.ServerName,
		Concurrency:                   gb.settings.Concurrency,
		NoDefaultDate:                 gb.settings.DisableDefaultDate,
		NoDefaultContentType:          gb.settings.DisableDefaultContentType,
		DisableHeaderNamesNormalizing: gb.settings.DisableHeaderNormalizing,
		DisableKeepalive:              gb.settings.DisableKeepalive,
		NoDefaultServerHeader:         gb.settings.ServerName == "",
		ReadTimeout:                   gb.settings.ReadTimeout,
		WriteTimeout:                  gb.settings.WriteTimeout,
		IdleTimeout:                   gb.settings.IdleTimeout,
	}
}

// Stop serving
func (gb *gearbox) Stop() error {
	err := gb.httpServer.Shutdown()

	// check if shutdown was ok and server had valid address
	if err == nil && gb.address != "" {
		log.Printf("%s stopped listening on %s", Name, gb.address)
		return nil
	}

	return err
}

// Get registers an http relevant method
func (gb *gearbox) Get(path string, handlers ...handlerFunc) *Route {
	return gb.registerRoute(string(MethodGet), string(path), handlers)
}

// Head registers an http relevant method
func (gb *gearbox) Head(path string, handlers ...handlerFunc) *Route {
	return gb.registerRoute(string(MethodHead), string(path), handlers)
}

// Post registers an http relevant method
func (gb *gearbox) Post(path string, handlers ...handlerFunc) *Route {
	return gb.registerRoute(string(MethodPost), string(path), handlers)
}

// Put registers an http relevant method
func (gb *gearbox) Put(path string, handlers ...handlerFunc) *Route {
	return gb.registerRoute(string(MethodPut), string(path), handlers)
}

// Patch registers an http relevant method
func (gb *gearbox) Patch(path string, handlers ...handlerFunc) *Route {
	return gb.registerRoute(string(MethodPatch), string(path), handlers)
}

// Delete registers an http relevant method
func (gb *gearbox) Delete(path string, handlers ...handlerFunc) *Route {
	return gb.registerRoute(string(MethodDelete), string(path), handlers)
}

// Connect registers an http relevant method
func (gb *gearbox) Connect(path string, handlers ...handlerFunc) *Route {
	return gb.registerRoute(string(MethodConnect), string(path), handlers)
}

// Options registers an http relevant method
func (gb *gearbox) Options(path string, handlers ...handlerFunc) *Route {
	return gb.registerRoute(string(MethodOptions), string(path), handlers)
}

// Trace registers an http relevant method
func (gb *gearbox) Trace(path string, handlers ...handlerFunc) *Route {
	return gb.registerRoute(string(MethodTrace), string(path), handlers)
}

// Trace registers an http relevant method
func (gb *gearbox) Method(method, path string, handlers ...handlerFunc) *Route {
	return gb.registerRoute(string(method), string(path), handlers)
}

// Fallback registers an http handler only fired when no other routes match with request
func (gb *gearbox) Fallback(handlers ...handlerFunc) error {
	return gb.registerFallback(handlers)
}

// Use attaches a global middleware to the gearbox object.
// included in the handlers chain for all matched requests.
// it will always be executed before the handler and/or middlewares for the matched request
// For example, this is the right place for a logger or some security check or permission checking.
func (gb *gearbox) Use(middlewares ...handlerFunc) {
	gb.handlers = append(gb.handlers, middlewares...)
}

// Group appends a prefix to registered routes.
func (gb *gearbox) Group(prefix string, routes []*Route) []*Route {
	for _, route := range routes {
		route.Path = prefix + route.Path
	}
	return routes
}

// Handles all incoming requests and route them to proper handler according to
// method and path
func (gb *gearbox) handler(ctx *fasthttp.RequestCtx) {
	if handlers, params := gb.matchRoute(
		GetString(ctx.Request.Header.Method()),
		GetString(ctx.URI().Path())); handlers != nil {
		context := Context{
			RequestCtx: ctx,
			Params:     params,
			handlers:   append(gb.handlers, handlers...),
			index:      0,
		}
		context.handlers[0](&context)
		return
	}

	ctx.SetStatusCode(StatusNotFound)
}

func printStartupMessage(addr string) {
	if prefork.IsChild() {
		log.Printf("Started child proc #%v\n", os.Getpid())
	} else {
		log.Printf(banner, Version, addr)
	}
}
