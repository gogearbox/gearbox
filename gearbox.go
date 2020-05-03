// Package gearbox is a web framework with a focus on high performance and memory optimization
package gearbox

import (
	"fmt"
	"net"

	"github.com/valyala/fasthttp"
)

type gearbox struct {
	httpServer          *fasthttp.Server
	routingTree         *node
	pathKeywordsMapping *TST
	httpMapping         *TST
	registeredRoutes    []*route
}

// New creates a new instance.
func New() *gearbox {
	gb := new(gearbox)
	gb.httpMapping = setHTTPMethodsMapping()
	gb.pathKeywordsMapping = &TST{}
	gb.registeredRoutes = make([]*route, 0)
	return gb
}

// Start handling requests.
func (gb *gearbox) Start(address string) error {
	gb.extractKeywords()
	gb.constructRoutingTree()
	gb.httpServer = &fasthttp.Server{
		Handler: gb.handler,
	}

	ln, err := net.Listen("tcp4", address)
	if err != nil {
		return err
	}

	return gb.httpServer.Serve(ln)
}

// Stop serving
func (gb *gearbox) Stop() error {
	if gb.httpServer == nil {
		return fmt.Errorf("Service is not running")
	}
	return gb.httpServer.Shutdown()
}

// Get registers an http relevant method
func (gb *gearbox) Get(path string, handler func(*fasthttp.RequestCtx)) {
	gb.registerRoute(MethodGet, path, handler)
}

// Head registers an http relevant method
func (gb *gearbox) Head(path string, handler func(*fasthttp.RequestCtx)) {
	gb.registerRoute(MethodHead, path, handler)
}

// Post registers an http relevant method
func (gb *gearbox) Post(path string, handler func(*fasthttp.RequestCtx)) {
	gb.registerRoute(MethodPost, path, handler)
}

// Put registers an http relevant method
func (gb *gearbox) Put(path string, handler func(*fasthttp.RequestCtx)) {
	gb.registerRoute(MethodPut, path, handler)
}

// Patch registers an http relevant method
func (gb *gearbox) Patch(path string, handler func(*fasthttp.RequestCtx)) {
	gb.registerRoute(MethodPatch, path, handler)
}

// Delete registers an http relevant method
func (gb *gearbox) Delete(path string, handler func(*fasthttp.RequestCtx)) {
	gb.registerRoute(MethodDelete, path, handler)
}

// Connect registers an http relevant method
func (gb *gearbox) Connect(path string, handler func(*fasthttp.RequestCtx)) {
	gb.registerRoute(MethodConnect, path, handler)
}

// Options registers an http relevant method
func (gb *gearbox) Options(path string, handler func(*fasthttp.RequestCtx)) {
	gb.registerRoute(MethodOptions, path, handler)
}

// Trace registers an http relevant method
func (gb *gearbox) Trace(path string, handler func(*fasthttp.RequestCtx)) {
	gb.registerRoute(MethodTrace, path, handler)
}
