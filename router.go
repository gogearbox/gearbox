package gearbox

import (
	"log"
	"strings"
	"sync"

	"github.com/valyala/fasthttp"
)

var (
	defaultContentType = []byte("text/plain; charset=utf-8")
)

type router struct {
	trees    map[string]*node
	cache    map[string]*matchResult
	cacheLen int
	mutex    sync.RWMutex
	notFound handlersChain
	settings *Settings
	pool     sync.Pool
}

type matchResult struct {
	handlers handlersChain
	params   map[string]string
}

func (r *router) acquireCtx(fctx *fasthttp.RequestCtx) *context {
	ctx := r.pool.Get().(*context)

	ctx.index = 0
	ctx.paramValues = make(map[string]string)
	ctx.requestCtx = fctx
	return ctx
}

func (r *router) releaseCtx(ctx *context) {
	ctx.handlers = nil
	ctx.paramValues = nil
	ctx.requestCtx = nil
	r.pool.Put(ctx)
}

// handle registers handlers for provided method and path to be used
// in routing incoming requests
func (r *router) handle(method, path string, handlers handlersChain) {
	if len(path) == 0 {
		panic("path is empty")
	} else if len(method) == 0 {
		panic("method is empty")
	} else if path[0] != '/' {
		panic("path must begin with '/' in path '" + path + "'")
	} else if len(handlers) == 0 {
		panic("no handlers provided with path '" + path + "'")
	}

	if r.trees == nil {
		r.trees = make(map[string]*node)
	}

	root := r.trees[method]
	if root == nil {
		root = createRootNode()
		r.trees[method] = root
	}

	root.addRoute(path, handlers)
}

// allowed checks if provided path can be routed in another method(s)
func (r *router) allowed(reqMethod, path string, ctx *context) string {
	var allow string
	pathLen := len(path)
	if (pathLen == 1 && path[0] == '*') || (pathLen > 1 && path[1] == '*') {
		for method := range r.trees {
			if method == MethodOptions {
				continue
			}

			if len(allow) != 0 {
				allow += ", " + method
			} else {
				allow = method
			}
		}
		return allow
	}

	for method, tree := range r.trees {
		if method == reqMethod || method == MethodOptions {
			continue
		}

		handlers := tree.matchRoute(path, ctx)
		if handlers != nil {
			if len(allow) != 0 {
				allow += ", " + method
			} else {
				allow = method
			}
		}
	}

	if len(allow) > 0 {
		allow += ", " + MethodOptions
	}
	return allow
}

// Handler handles all incoming requests
func (r *router) Handler(fctx *fasthttp.RequestCtx) {
	context := r.acquireCtx(fctx)
	defer r.releaseCtx(context)

	if r.settings.AutoRecover {
		defer func(fctx *fasthttp.RequestCtx) {
			if rcv := recover(); rcv != nil {
				log.Printf("recovered from error: %v", rcv)
				fctx.Error(fasthttp.StatusMessage(fasthttp.StatusInternalServerError),
					fasthttp.StatusInternalServerError)
			}
		}(fctx)
	}

	path := GetString(fctx.URI().PathOriginal())

	if r.settings.CaseInSensitive {
		path = strings.ToLower(path)
	}

	method := GetString(fctx.Method())

	var cacheKey string
	if !r.settings.DisableCaching {
		cacheKey = path + method
		r.mutex.RLock()
		cacheResult, ok := r.cache[cacheKey]

		if ok {
			context.handlers = cacheResult.handlers
			context.paramValues = cacheResult.params
			r.mutex.RUnlock()
			context.handlers[0](context)
			return
		}
		r.mutex.RUnlock()
	}

	if root := r.trees[method]; root != nil {
		if handlers := root.matchRoute(path, context); handlers != nil {
			context.handlers = handlers
			context.handlers[0](context)

			if !r.settings.DisableCaching {
				r.mutex.Lock()

				if r.cacheLen == r.settings.CacheSize {
					r.cache = make(map[string]*matchResult)
					r.cacheLen = 0
				}
				r.cache[cacheKey] = &matchResult{
					handlers: handlers,
					params:   context.paramValues,
				}
				r.cacheLen++
				r.mutex.Unlock()
			}
			return
		}
	}

	if method == MethodOptions && r.settings.HandleOPTIONS {
		if allow := r.allowed(method, path, context); len(allow) > 0 {
			fctx.Response.Header.Set("Allow", allow)
			return
		}
	} else if r.settings.HandleMethodNotAllowed {
		if allow := r.allowed(method, path, context); len(allow) > 0 {
			fctx.Response.Header.Set("Allow", allow)
			fctx.SetStatusCode(fasthttp.StatusMethodNotAllowed)
			fctx.SetContentTypeBytes(defaultContentType)
			fctx.SetBodyString(fasthttp.StatusMessage(fasthttp.StatusMethodNotAllowed))
			return
		}
	}

	// Custom Not Found (404) handlers
	if r.notFound != nil {
		r.notFound[0](context)
		return
	}

	// Default Not Found response
	fctx.Error(fasthttp.StatusMessage(fasthttp.StatusNotFound),
		fasthttp.StatusNotFound)
}

func (r *router) SetNotFound(handlers handlersChain) {
	r.notFound = append(r.notFound, handlers...)
}
