package gearbox

import (
	"log"
	"strings"

	"github.com/valyala/fasthttp"
)

type methodType uint8

// Http methods
const (
	// Get Method
	MethodGet methodType = iota + 1
	// Head Method
	MethodHead
	// Post Method
	MethodPost
	// Put Method
	MethodPut
	// Patch Method
	MethodPatch
	// Delete Method
	MethodDelete
	// Connect Method
	MethodConnect
	// Options Method
	MethodOptions
	// Trace Method
	MethodTrace
)

var httpMethodsList = [...]string{
	fasthttp.MethodGet,
	fasthttp.MethodHead,
	fasthttp.MethodPost,
	fasthttp.MethodPut,
	fasthttp.MethodPatch,
	fasthttp.MethodDelete,
	fasthttp.MethodConnect,
	fasthttp.MethodOptions,
	fasthttp.MethodTrace,
}

type node struct {
	Name     string
	Method   methodType
	MatchAll bool
	Handlers []func(*fasthttp.RequestCtx)
	Children []*node
}

type route struct {
	Method  methodType
	Path    string
	Handler func(*fasthttp.RequestCtx)
}

func setHTTPMethodsMapping() *TST {
	mapping := &TST{}

	mapping.Set(fasthttp.MethodGet, MethodGet)
	mapping.Set(fasthttp.MethodHead, MethodHead)
	mapping.Set(fasthttp.MethodPost, MethodPost)
	mapping.Set(fasthttp.MethodPut, MethodPut)
	mapping.Set(fasthttp.MethodPatch, MethodPatch)
	mapping.Set(fasthttp.MethodDelete, MethodDelete)
	mapping.Set(fasthttp.MethodConnect, MethodConnect)
	mapping.Set(fasthttp.MethodOptions, MethodOptions)
	mapping.Set(fasthttp.MethodTrace, MethodTrace)
	return mapping
}

func getHTTPMethodStr(mType methodType) string {
	index := mType - 1
	if index < 0 || int(index) > len(httpMethodsList) {
		log.Fatalf("Unknown method type %v", mType)
	}
	return httpMethodsList[index]
}

func validateRoutePath(path string) bool {
	length := len(path)
	if length == 0 {
		return false
	}

	if path[0] != '/' {
		return false
	}

	starIndex := strings.Index(path, "*")
	if starIndex > -1 && starIndex < length-1 && path[starIndex-1] == '/' {
		return false
	}

	return true
}

func (gb *gearbox) registerRoute(mType methodType, path string, handler func(*fasthttp.RequestCtx)) {
	if handler == nil {
		log.Fatalf("Route %s with Method %s does not contain any handlers!", path, getHTTPMethodStr(mType))
	}

	if !validateRoutePath(path) {
		log.Fatalf("Route %s is not valid!", path)
	}

	gb.registeredRoutes = append(gb.registeredRoutes, &route{
		Path:    path,
		Method:  mType,
		Handler: handler,
	})
}

func (gb *gearbox) extractKeywords() {
	counter := 0
	for i := range gb.registeredRoutes {
		keywords := strings.Split(gb.registeredRoutes[i].Path, "/")
		for j := range keywords {
			if keywords[j] == "*" || keywords[j] == "" {
				continue
			} else if gb.pathKeywordsMapping.Get(keywords[j]) == nil {
				gb.pathKeywordsMapping.Set(keywords[j], counter)
				counter++
			}
		}
	}
}

func (gb *gearbox) createEmptyNode(name string) *node {
	return &node{
		Name:     name,
		Children: make([]*node, gb.pathKeywordsMapping.Count()),
		Handlers: make([]func(*fasthttp.RequestCtx), MethodTrace),
		MatchAll: false,
	}
}

func (gb *gearbox) constructRoutingTree() {
	gb.routingTree = gb.createEmptyNode("root")
	for _, route := range gb.registeredRoutes {
		currentNode := gb.routingTree
		keywords := strings.Split(route.Path, "/")
		keywordsLen := len(keywords)
		for i := 1; i < keywordsLen; i++ {
			keyword := keywords[i]
			if keyword == "" || keyword == "*" {
				if keyword == "*" {
					currentNode.MatchAll = true
				}
				continue
			}

			keywordID := gb.pathKeywordsMapping.Get(keyword).(int)
			if currentNode.Children[keywordID] == nil {
				currentNode.Children[keywordID] = gb.createEmptyNode(keyword)
			}
			currentNode = currentNode.Children[keywordID]
		}
		currentNode.Handlers[route.Method-1] = route.Handler
	}
}

func (gb *gearbox) matchRoute(mType methodType, path string) func(*fasthttp.RequestCtx) {
	keywords := strings.Split(path, "/")
	keywordsLen := len(keywords)
	currentNode := gb.routingTree
	var lastMatchAll *node
	if currentNode.MatchAll {
		lastMatchAll = currentNode
	}

	for i := 1; i < keywordsLen; i++ {
		keyword := keywords[i]
		if keyword == "" {
			continue
		}

		keywordID, ok := gb.pathKeywordsMapping.Get(keyword).(int)
		if !ok || currentNode.Children[keywordID] == nil {
			if lastMatchAll == nil {
				return nil
			}
			return lastMatchAll.Handlers[mType-1]
		}

		currentNode = currentNode.Children[keywordID]
		if currentNode.MatchAll {
			lastMatchAll = currentNode
		}
	}

	return currentNode.Handlers[mType-1]
}

func (gb *gearbox) handler(ctx *fasthttp.RequestCtx) {
	method := ctx.Request.Header.Method()
	path := ctx.URI().Path()
	methodID, ok := gb.httpMapping.Get(string(method)).(methodType)
	if ok {
		handler := gb.matchRoute(methodID, string(path))
		if handler != nil {
			handler(ctx)
			return
		}
	}
	ctx.Error("Unsupported path", fasthttp.StatusNotFound)
}
