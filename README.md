<p align="center">
	<a href="https://gogearbox.com">
    	<img src="https://raw.githubusercontent.com/gogearbox/gearbox/master/assets/gearbox-512.png"/>
	</a>
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


**gearbox** :gear: is a web framework for building micro services written in Go with a focus on high performance. It's built on [fasthttp](https://github.com/valyala/fasthttp) which is **up to 10x faster** than net/http


### gearbox seeks to be
+ Secure :closed_lock_with_key:
+ Fast :rocket:
+ Easy to use :eyeglasses:
+ Lightweight


### Supported Go versions & installation

:gear: gearbox requires version `1.14` or higher of Go ([Download Go](https://golang.org/dl/))

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
	gb.Get("/hello", func(ctx gearbox.Context) {
		ctx.SendString("Hello World!")
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
	gb.Get("/users/:user", func(ctx gearbox.Context) {
		ctx.SendString(ctx.Param("user"))
	})

	// Start service
	gb.Start(":3000")
}
```

#### Middlewares
```go
package main

import (
	"log"

	"github.com/gogearbox/gearbox"
)

func main() {
	// Setup gearbox
	gb := gearbox.New()

	// create a logger middleware
	logMiddleware := func(ctx gearbox.Context) {
		log.Printf("log message!")

		// Next is what allows the request to continue to the next
		// middleware/handler
		ctx.Next()
	}

	// create an unauthorized middleware
	unAuthorizedMiddleware := func(ctx gearbox.Context) {
		ctx.Status(gearbox.StatusUnauthorized)
			.SendString("You are unauthorized to access this page!")
	}

	// Register the log middleware for all requests
	gb.Use(logMiddleware)

	// Define your handlers
	gb.Get("/hello", func(ctx gearbox.Context) {
		ctx.SendString("Hello World!")
	})

	// Register the routes to be used when grouping routes
	routes := []*gearbox.Route{
		gb.Get("/id", func(ctx gearbox.Context) {
			ctx.SendString("User X")
		}),
		gb.Delete("/id", func(ctx gearbox.Context) {
			ctx.SendString("Deleted")
		}),
	}

	// Group account routes
	accountRoutes := gb.Group("/account", routes)

	// Group account routes to be under api
	gb.Group("/api", accountRoutes)

	// Define a route with unAuthorizedMiddleware as the middleware
	// you can define as many middlewares as you want and have
	// the handler as the last argument
	gb.Get("/protected", unAuthorizedMiddleware, func(ctx gearbox.Context) {
		ctx.SendString("You accessed a protected page")
	})

	// Start service
	gb.Start(":3000")
}
```

#### Static Files

```go
package main

import (
	"github.com/gogearbox/gearbox"
)

func main() {
	// Setup gearbox
	gb := gearbox.New()

	// Serve files in assets directory for prefix static
	// for example /static/gearbox.png, etc.
	gb.Static("/static", "./assets")

	// Start service
	gb.Start(":3000")
}
```

### Benchmarks

- **CPU** 3.1 GHz Intel Xeon® Platinum 8175M (24 physical cores, 48 logical cores)
- **MEMORY** 192GB
- **GO** go 1.14.6 linux/amd64
- **OS** Linux

<p align="center">
	<img src="https://raw.githubusercontent.com/gogearbox/gearbox/master/assets/benchmark-pipeline.png" width="85%"/>
</p>

For more results, check [Our Docs](https://gogearbox.com/docs/benchmarks)

### Contribute & Support
+ Add a [GitHub Star](https://github.com/gogearbox/gearbox/stargazers)
+ [Suggest new features, ideas and optimizations](https://github.com/gogearbox/gearbox/issues)
+ [Report issues](https://github.com/gogearbox/gearbox/issues)
+ Donating a [cup of coffee](https://buymeacoff.ee/gearbox)


Check [Our Docs](https://gogearbox.com/docs) for more information about **gearbox** and how to **contribute**

### Sponsors
Organizations that are helping to manage, promote, and support **Gearbox** :gear: 

| [<img src="https://raw.githubusercontent.com/gogearbox/gearbox/master/assets/trella-sponsor.png"/>](https://www.trella.app) 	|
|:-:	|
| [trella](https://www.trella.app): *A B2B technology platform and trucking <br/>marketplace that connects shippers with carriers* |


### Who uses Gearbox
**Gearbox** :gear: is being used by multiple organizations including but not limited to 

[<img src="https://raw.githubusercontent.com/gogearbox/gearbox/master/assets/erply-user.png"/>](https://erply.com)	
[<img src="https://raw.githubusercontent.com/gogearbox/gearbox/master/assets/trella-sponsor.png"/>](https://www.trella.app)


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
- [json-iterator](https://github.com/json-iterator/go/blob/master/LICENSE)
