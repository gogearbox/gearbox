package gearbox

import (
	"fmt"
	"strings"

	jsoniter "github.com/json-iterator/go"
	"github.com/valyala/fasthttp"
)

// MIME types
const (
	MIMEApplicationJSON = "application/json"
)

// Context interface
type Context interface {
	Next()
	Context() *fasthttp.RequestCtx
	Param(key string) string
	Query(key string) string
	SendBytes(value []byte) Context
	SendString(value string) Context
	SendJSON(in interface{}) error
	Status(status int) Context
	Set(key string, value string)
	Get(key string) string
	SetLocal(key string, value interface{})
	GetLocal(key string) interface{}
	Body() string
	ParseBody(out interface{}) error
}

// handlerFunc defines the handler used by middleware as return value.
type handlerFunc func(ctx Context)

// handlersChain defines a handlerFunc array.
type handlersChain []handlerFunc

// Context defines the current context of request and handlers/middlewares to execute
type context struct {
	requestCtx  *fasthttp.RequestCtx
	paramValues map[string]string
	handlers    handlersChain
	index       int
}

// Next function is used to successfully pass from current middleware to next middleware.
// if the middleware thinks it's okay to pass it
func (ctx *context) Next() {
	ctx.index++
	if ctx.index < len(ctx.handlers) {
		ctx.handlers[ctx.index](ctx)
	}
}

// Param returns value of path parameter specified by key
func (ctx *context) Param(key string) string {
	return ctx.paramValues[key]
}

// Context returns Fasthttp context
func (ctx *context) Context() *fasthttp.RequestCtx {
	return ctx.requestCtx
}

// SendBytes sets body of response for []byte type
func (ctx *context) SendBytes(value []byte) Context {
	ctx.requestCtx.Response.SetBodyRaw(value)
	return ctx
}

// SendString sets body of response for string type
func (ctx *context) SendString(value string) Context {
	ctx.requestCtx.SetBodyString(value)
	return ctx
}

// SendJSON converts any interface to json, sets it to the body of response
// and sets content type header to application/json.
func (ctx *context) SendJSON(in interface{}) error {
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	raw, err := json.Marshal(in)
	// Check for errors
	if err != nil {
		return err
	}

	// Set http headers
	ctx.requestCtx.Response.Header.SetContentType(MIMEApplicationJSON)
	ctx.requestCtx.Response.SetBodyRaw(raw)

	return nil
}

// Status sets the HTTP status code
func (ctx *context) Status(status int) Context {
	ctx.requestCtx.Response.SetStatusCode(status)
	return ctx
}

// Get returns the HTTP request header specified by field key
func (ctx *context) Get(key string) string {
	return GetString(ctx.requestCtx.Request.Header.Peek(key))
}

// Set sets the response's HTTP header field key to the specified key, value
func (ctx *context) Set(key, value string) {
	ctx.requestCtx.Response.Header.Set(key, value)
}

// Query returns the query string parameter in the request url
func (ctx *context) Query(key string) string {
	return GetString(ctx.requestCtx.QueryArgs().Peek(key))
}

// Body returns the raw body submitted in a POST request
func (ctx *context) Body() string {
	return GetString(ctx.requestCtx.Request.Body())
}

// SetLocal stores value with key within request scope and it is accessible through
// handlers of that request
func (ctx *context) SetLocal(key string, value interface{}) {
	ctx.requestCtx.SetUserValue(key, value)
}

// GetLocal gets value by key which are stored by SetLocal within request scope
func (ctx *context) GetLocal(key string) interface{} {
	return ctx.requestCtx.UserValue(key)
}

// ParseBody parses request body into provided struct
// Supports decoding theses types: application/json
func (ctx *context) ParseBody(out interface{}) error {
	contentType := GetString(ctx.requestCtx.Request.Header.ContentType())
	if strings.HasPrefix(contentType, MIMEApplicationJSON) {
		json := jsoniter.ConfigCompatibleWithStandardLibrary
		return json.Unmarshal(ctx.requestCtx.Request.Body(), out)
	}

	return fmt.Errorf("content type '%s' is not supported, "+
		"please open a request to support it "+
		"(https://github.com/gogearbox/gearbox/issues/new?template=feature_request.md)",
		contentType)
}
