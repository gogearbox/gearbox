package gearbox

import "github.com/valyala/fasthttp"

// HandlerFunc defines the handler used by middleware as return value.
type HandlerFunc func(ctx *Context)

// HandlersChain defines a HandlerFunc array.
type HandlersChain []HandlerFunc

// Context defines the current context of request and handlers/middlewares to execute
type Context struct {
	*fasthttp.RequestCtx
	handlers HandlersChain
	index    int
}


// Next function is used to successfully pass from current middleware to next middleware.
// if the middleware thinks it's okay to pass it.
func (ctx *Context) Next() {
	ctx.index++
	ctx.handlers[ctx.index](ctx)
}
