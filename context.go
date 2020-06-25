package gearbox

import (
	"fmt"

	"github.com/valyala/fasthttp"
)

// handlerFunc defines the handler used by middleware as return value.
type handlerFunc func(ctx *Context)

// handlersChain defines a handlerFunc array.
type handlersChain []handlerFunc

// Context defines the current context of request and handlers/middlewares to execute
type Context struct {
	RequestCtx *fasthttp.RequestCtx
	Params     map[string]string
	handlers   handlersChain
	index      int
}

// Next function is used to successfully pass from current middleware to next middleware.
// if the middleware thinks it's okay to pass it.
func (ctx *Context) Next(err ...error) {
	if len(err) > 0 {
		defer func() {
			if r := recover(); r != nil {
				err, ok := r.(error)
				if !ok {
					err = fmt.Errorf("%v", r)
				}
				ctx.Next(err)
				return
			}
		}()
		ctx.Next()
	}
	ctx.index++
	ctx.handlers[ctx.index](ctx)
}
