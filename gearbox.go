// Package gearbox is a web framework with a focus on high performance and memory optimization
package gearbox

import (
	"fmt"
	"net"

	"github.com/valyala/fasthttp"
)

// gearbox interface
type gearbox interface {
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
}

type gearboxApp struct {
	httpServer       *fasthttp.Server
	routingTreeRoot  *routeNode
	registeredRoutes []*routeInfo
}

// New creates a new instance
func New() gearbox {
	gb := new(gearboxApp)
	gb.registeredRoutes = make([]*routeInfo, 0)
	return gb
}

// Start handling requests
func (gb *gearboxApp) Start(address string) error {
	gb.constructRoutingTree()
	gb.httpServer = &fasthttp.Server{
		Handler:      gb.handler,
		Logger:       nil,
		LogAllErrors: false,
	}

	ln, err := net.Listen("tcp4", address)
	if err != nil {
		return err
	}

	return gb.httpServer.Serve(ln)
}

// Stop serving
func (gb *gearboxApp) Stop() error {
	if gb.httpServer == nil {
		return fmt.Errorf("Service is not running")
	}
	return gb.httpServer.Shutdown()
}

// Get registers an http relevant method
func (gb *gearboxApp) Get(path string, handler func(*fasthttp.RequestCtx)) error {
	return gb.registerRoute(MethodGet, path, handler)
}

// Head registers an http relevant method
func (gb *gearboxApp) Head(path string, handler func(*fasthttp.RequestCtx)) error {
	return gb.registerRoute(MethodHead, path, handler)
}

// Post registers an http relevant method
func (gb *gearboxApp) Post(path string, handler func(*fasthttp.RequestCtx)) error {
	return gb.registerRoute(MethodPost, path, handler)
}

// Put registers an http relevant method
func (gb *gearboxApp) Put(path string, handler func(*fasthttp.RequestCtx)) error {
	return gb.registerRoute(MethodPut, path, handler)
}

// Patch registers an http relevant method
func (gb *gearboxApp) Patch(path string, handler func(*fasthttp.RequestCtx)) error {
	return gb.registerRoute(MethodPatch, path, handler)
}

// Delete registers an http relevant method
func (gb *gearboxApp) Delete(path string, handler func(*fasthttp.RequestCtx)) error {
	return gb.registerRoute(MethodDelete, path, handler)
}

// Connect registers an http relevant method
func (gb *gearboxApp) Connect(path string, handler func(*fasthttp.RequestCtx)) error {
	return gb.registerRoute(MethodConnect, path, handler)
}

// Options registers an http relevant method
func (gb *gearboxApp) Options(path string, handler func(*fasthttp.RequestCtx)) error {
	return gb.registerRoute(MethodOptions, path, handler)
}

// Trace registers an http relevant method
func (gb *gearboxApp) Trace(path string, handler func(*fasthttp.RequestCtx)) error {
	return gb.registerRoute(MethodTrace, path, handler)
}

// Handles all incoming requests and route them to proper handler according to
// method and path
func (gb *gearboxApp) handler(ctx *fasthttp.RequestCtx) {
	if handler := gb.matchRoute(getString(ctx.Request.Header.Method()), getString(ctx.URI().Path())); handler != nil {
		handler(ctx)
		return
	}

	ctx.Error("Not found", fasthttp.StatusNotFound)
}
