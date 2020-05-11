<p align="center">
    <img src="https://github.com/abahmed/gearbox/blob/master/assets/gearbox-512.png"/>
    <br />
    <a href="https://godoc.org/github.com/abahmed/gearbox">
      <img src="https://godoc.org/github.com/abahmed/gearbox?status.png" />
    </a>
    <img src="https://github.com/abahmed/gearbox/workflows/Test%20&%20Build/badge.svg?branch=master" />
    <a href="https://codecov.io/gh/abahmed/gearbox">
      <img src="https://codecov.io/gh/abahmed/gearbox/branch/master/graph/badge.svg" />
    </a>
    <a href="https://goreportcard.com/report/github.com/abahmed/gearbox">
      <img src="https://goreportcard.com/badge/github.com/abahmed/gearbox" />
    </a>
    <a href="https://gitter.im/abahmed/gearbox?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge">
      <img src="https://badges.gitter.im/abahmed/gearbox.svg"/>
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
go get -u github.com/abahmed/gearbox
```

### Example

```go
package main

import (
  "github.com/abahmed/gearbox"
  "github.com/valyala/fasthttp"
)

func main() {
  // Setup gearbox
  gearbox := gearbox.New()
  
  // Define your handlers
  gearbox.Get("/hello", func(ctx *fasthttp.RequestCtx) {
	ctx.Response.SetBodyString("Hello World!")
  })

  // Start service
  gearbox.Start(":3000")
}
```

### Contribute & Support
+ Add a [GitHub Star](https://github.com/abahmed/gearbox/stargazers)
+ [Suggest new features, ideas and optimizations](https://github.com/abahmed/gearbox/issues)
+ [Report issues](https://github.com/abahmed/gearbox/issues)

Check [Our Wiki](https://github.com/abahmed/gearbox/wiki) for more information about **gearbox** and how to **contribute**

### Contributors

<a href="https://github.com/abahmed/gearbox/graphs/contributors">
  <img src="https://contributors-img.firebaseapp.com/image?repo=abahmed/gearbox" />
</a>

### Get in touch!

Feel free to Join us on [Gitter](https://gitter.im/abahmed/gearbox), or email us at [gearbox@googlegroups.com](gearbox@googlegroups.com)  if you have questions, or suggestions

### License

gearbox is licensed under [MIT License](LICENSE)

Logo is created by [Mahmoud Sayed](https://www.facebook.com/mahmoudsayedae) and distributed under [Creative Commons License](https://creativecommons.org/licenses/by-sa/4.0/)

#### Third-party library licenses
- [FastHTTP](https://github.com/valyala/fasthttp/blob/master/LICENSE)
