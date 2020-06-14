<p align="center">
    <img src="https://raw.githubusercontent.com/gogearbox/gearbox/master/assets/gearbox-512.png"/>
    <br />
    <a href="https://godoc.org/github.com/gogearbox/gearbox">
      <img src="https://godoc.org/github.com/gogearbox/gearbox?status.png" />
    </a>
    <img src="https://github.com/gogearbox/gearbox/workflows/Test%20&%20Build/badge.svg?branch=master" />
    <a href="https://codecov.io/gh/gogearbox/gearbox">
      <img src="https://codecov.io/gh/gogearbox/gearbox/branch/master/graph/badge.svg" />
    </a>
    <a href="https://goreportcard.com/report/github.com/gogearbox/gearbox">
      <img src="https://goreportcard.com/badge/github.com/gogearbox/gearbox" />
    </a>
	<a href="https://discord.com/invite/CT8my4R">
      <img src="https://img.shields.io/discord/716724372642988064?label=Discord&logo=discord">
  	</a>
    <a href="https://deepsource.io/gh/gogearbox/gearbox/?ref=repository-badge" target="_blank">
      <img alt="DeepSource" title="DeepSource" src="https://static.deepsource.io/deepsource-badge-light-mini.svg">
    </a>
</p>


**gearbox** :gear: is a web framework for building micro services written in Go with a focus on high performance and memory optimization. 

Currently, **gearbox** :gear: is **under development (not production ready)** and built on [fasthttp](https://github.com/valyala/fasthttp) which is **up to 10x faster** than net/http

In **gearbox**, we care about peformance and memory which will be used by each method while building things up and how we can improve that. It also takes more time to **research** about each component that will be used and **compare** it with different implementations of other open source web frameworks. It may end up writing our **own components** in an optimized way to achieve our goals

### gearbox seeks to be
+ Secure :closed_lock_with_key:
+ Fast :rocket:
+ Simple :eyeglasses:
+ Easy to use
+ Lightweight


### Supported Go versions & installation

:gear: gearbox requires version `1.11` or higher of Go ([Download Go](https://golang.org/dl/))

Just use [go get](https://golang.org/cmd/go/#hdr-Add_dependencies_to_current_module_and_install_them) to download and install gearbox

```bash
go get -u github.com/gogearbox/gearbox
```

### Examples

```go
package main

import (
	"github.com/gogearbox/gearbox"
)

func main() {
	// Setup gearbox
	gb := gearbox.New()

	// Define your handlers
	gb.Get("/hello", func(ctx *gearbox.Context) {
		ctx.RequestCtx.Response.SetBodyString("Hello World!")
	})

	// Start service
	gb.Start(":3000")
}
```

#### Parameters
```go
package main

import (
	"github.com/gogearbox/gearbox"
)

func main() {
	// Setup gearbox
	gb := gearbox.New()

	// Handler with parameter
	gb.Get("/users/:user", func(ctx *gearbox.Context) {
		fmt.Printf("%s\n", ctx.Params.GetString("user"))
	})

	// Handler with optional parameter
	gb.Get("/search/:pattern?", func(ctx *gearbox.Context) {
		fmt.Printf("%s\n", ctx.Params.GetString("pattern"))
	})

	// Handler with regex parameter
	gb.Get("/book/:name:([a-z]+[0-3])", func(ctx *gearbox.Context) {
		fmt.Printf("%s\n", ctx.Params.GetString("name"))
	})

	// Start service
	gb.Start(":3000")
}
```

#### Middlewares
```go
package main

import (
	"github.com/gogearbox/gearbox"
	"log"
)

func main() {
	// Setup gearbox
	gb := gearbox.New()

	// create a logger middleware
	logMiddleware := func(ctx *gearbox.Context) {
		log.Printf(ctx.RequestCtx.String())
		ctx.Next() // Next is what allows the request to continue to the next middleware/handler
	}

	// create an unauthorized middleware
	unAuthorizedMiddleware := func(ctx *gearbox.Context) {
		ctx.RequestCtx.SetStatusCode(401) // unauthorized status code
		ctx.RequestCtx.Response.SetBodyString("You are unauthorized to access this page!")
	}

	// Register the log middleware for all requests
	gb.Use(logMiddleware)

	// Define your handlers
	gb.Get("/hello", func(ctx *gearbox.Context) {
		ctx.RequestCtx.Response.SetBodyString("Hello World!")
	})
    
    // Register the routes to be used when grouping routes
    routes := []*gearbox.Route{gb.Get("/id", func(ctx *gearbox.Context) {
        ctx.RequestCtx.Response.SetBodyString("User X")
    }), gb.Delete("/id", func(ctx *gearbox.Context) {
        ctx.RequestCtx.Response.SetBodyString("Deleted")
    })}
    
    // Group account routes
    accountRoutes := gb.Group("/account", routes)
    
    // Group account routes to be under api
    gb.Group("/api", accountRoutes)

	// Define a route with unAuthorizedMiddleware as the middleware
	// you can define as many middlewares as you want and have the handler as the last argument
	gb.Get("/protected", unAuthorizedMiddleware, func(ctx *gearbox.Context) {
		ctx.RequestCtx.Response.SetBodyString("You accessed a protected page")
	})

	// Start service
	gb.Start(":3000")
}
```

### Contribute & Support
+ Add a [GitHub Star](https://github.com/gogearbox/gearbox/stargazers)
+ [Suggest new features, ideas and optimizations](https://github.com/gogearbox/gearbox/issues)
+ [Report issues](https://github.com/gogearbox/gearbox/issues)

Check [Our Docs](https://gogearbox.com/docs) for more information about **gearbox** and how to **contribute**

### Contributors

<a href="https://github.com/gogearbox/gearbox/graphs/contributors">
  <img src="https://contributors-img.firebaseapp.com/image?repo=gogearbox/gearbox" />
</a>

### Get in touch!

Feel free to chat with us on [Discord](https://discord.com/invite/CT8my4R), or email us at [gearbox@googlegroups.com](gearbox@googlegroups.com)  if you have questions, or suggestions

### License

gearbox is licensed under [MIT License](LICENSE)

Logo is created by [Mahmoud Sayed](https://www.facebook.com/mahmoudsayedae) and distributed under [Creative Commons License](https://creativecommons.org/licenses/by-sa/4.0/)

#### Third-party library licenses
- [FastHTTP](https://github.com/valyala/fasthttp/blob/master/LICENSE)
