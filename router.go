package gearbox

import (
	"bytes"
	"fmt"
	"regexp"
	"sort"
	"sync"
)

type routeNode struct {
	Name      []byte
	Endpoints tst
	Children  tst
}

type paramType uint8

// Supported parameter types
const (
	ptNoParam  paramType = iota // No parameter (most strict)
	ptRegexp                    // Regex parameter
	ptParam                     // Normal parameter
	ptMatchAll                  // Match all parameter (least strict)
)

type param struct {
	Name       []byte
	Value      string
	Type       paramType
	IsOptional bool
}

type endpoint struct {
	Params   []*param
	Handlers handlersChain
}

type route struct {
	Method   []byte
	Path     []byte
	Handlers handlersChain
}

type routerFallback struct {
	Handlers handlersChain
}

type matchParamsResult struct {
	Matched  bool
	Handlers handlersChain
	Params   tst
}

// validateRoutePath makes sure that path complies with path's rules
func validateRoutePath(path []byte) error {
	// Check length of the path
	length := len(path)
	if length == 0 {
		return fmt.Errorf("length is zero")
	}

	// Make sure path starts with /
	if path[0] != '/' {
		return fmt.Errorf("path must start with /")
	}

	params := newTST()
	parts := bytes.Split(trimPath(path), []byte("/"))
	partsLen := len(parts)
	for i := 0; i < partsLen; i++ {
		if len(parts[i]) == 0 {
			continue
		}
		if p := parseParameter(parts[i]); p != nil {
			if p.Type == ptMatchAll && i != partsLen-1 {
				return fmt.Errorf("* must be in the end of path")
			} else if p.IsOptional && i != partsLen-1 {
				return fmt.Errorf("only last parameter can be optional")
			} else if p.Type == ptParam || p.Type == ptRegexp {
				if pName := params.Get(p.Name); pName != nil {
					return fmt.Errorf("parameter is duplicated")
				}
				params.Set(p.Name, true)
			}
		}
	}

	return nil
}

// registerRoute registers handler with method and path
func (gb *gearbox) registerRoute(method, path []byte, handlers handlersChain) error {
	// Handler is not provided
	if handlers == nil {
		return fmt.Errorf("route %s with method %s does not contain any handlers", path, method)
	}

	// Check if path is valid or not
	if err := validateRoutePath(path); err != nil {
		return fmt.Errorf("route %s is not valid! - %s", path, err.Error())
	}

	if !gb.settings.CaseSensitive {
		path = bytes.ToLower(path)
	}

	// Add route to registered routes
	gb.registeredRoutes = append(gb.registeredRoutes, &route{
		Path:     path,
		Method:   method,
		Handlers: handlers,
	})
	return nil
}

// registerFallback registers a single handler that will match only if all other routes fail to match
func (gb *gearbox) registerFallback(handlers handlersChain) error {
	// Handler is not provided
	if handlers == nil {
		return fmt.Errorf("fallback does not contain a handler")
	}

	gb.registeredFallback = &routerFallback{Handlers: handlers}
	return nil
}

// createEmptyRouteNode creates a new route node with name
func createEmptyRouteNode(name []byte) *routeNode {
	return &routeNode{
		Name:      name,
		Children:  newTST(),
		Endpoints: newTST(),
	}
}

// parseParameter parses path part into param struct, or returns nil if it's
// not a parameter
func parseParameter(pathPart []byte) *param {
	pathPartLen := len(pathPart)
	if pathPartLen == 0 {
		return nil
	}

	// match all
	if pathPart[0] == '*' {
		return &param{
			Name: []byte("*"),
			Type: ptMatchAll,
		}
	}

	isOptional := pathPart[pathPartLen-1] == '?'
	if isOptional {
		pathPart = pathPart[0 : pathPartLen-1]
	}

	params := bytes.Split(pathPart, []byte(":"))
	paramsLen := len(params)

	if paramsLen == 2 && len(params[0]) == 0 { // Just a parameter
		return &param{
			Name:       params[1],
			Type:       ptParam,
			IsOptional: isOptional,
		}
	} else if paramsLen == 3 && len(params[0]) == 0 { // Regex parameter
		return &param{
			Name:       params[1],
			Value:      string(params[2]),
			Type:       ptRegexp,
			IsOptional: isOptional,
		}
	}

	return nil
}

// getLeastStrictParamType returns least strict parameter type from list of
// parameters
func getLeastStrictParamType(params []*param) paramType {
	pLen := len(params)
	if pLen == 0 {
		return ptNoParam
	}

	pType := params[0].Type
	for i := 1; i < pLen; i++ {
		if params[i].Type > pType {
			pType = params[i].Type
		}
	}
	return pType
}

func isValidEndpoint(endpoints []*endpoint, newEndpoint *endpoint) bool {
	for i := range endpoints {
		if len(endpoints[i].Params) == len(newEndpoint.Params) {
			isValid := false
			for j := range endpoints[i].Params {
				if endpoints[i].Params[j].Type != newEndpoint.Params[j].Type {
					isValid = true
				}
			}
			return isValid
		}
	}
	return true
}

// trimPath trims left and right slashes in path
func trimPath(path []byte) []byte {
	pathLastIndex := len(path) - 1

	for path[pathLastIndex] == '/' && pathLastIndex > 0 {
		pathLastIndex--
	}

	pathFirstIndex := 1
	if path[0] != '/' {
		pathFirstIndex = 0
	}

	return path[pathFirstIndex : pathLastIndex+1]
}

// constructRoutingTree constructs routing tree according to provided routes
func (gb *gearbox) constructRoutingTree() error {
	// Firstly, create root node
	gb.routingTreeRoot = createEmptyRouteNode([]byte("root"))

	for _, route := range gb.registeredRoutes {
		currentNode := gb.routingTreeRoot
		params := make([]*param, 0)

		// Split path into slices of parts
		parts := bytes.Split(route.Path, []byte("/"))

		partsLen := len(parts)
		for i := 1; i < partsLen; i++ {
			part := parts[i]

			// Do not create node if part is empty
			if len(part) == 0 {
				continue
			}

			// Parse part as a parameter if it is
			if param := parseParameter(part); param != nil {
				params = append(params, param)
				continue
			}

			// Try to get a child of current node with part, otherwise
			//creates a new node and make it current node
			partNode, ok := currentNode.Children.Get(part).(*routeNode)
			if !ok {
				partNode = createEmptyRouteNode(part)
				currentNode.Children.Set(part, partNode)
			}
			currentNode = partNode
		}

		currentEndpoint := &endpoint{
			Handlers: route.Handlers,
			Params:   params,
		}

		// Make sure that current node does not have a handler for route's method
		var endpoints []*endpoint
		if result, ok := currentNode.Endpoints.Get(route.Method).([]*endpoint); ok {
			if ok := isValidEndpoint(result, currentEndpoint); !ok {
				return fmt.Errorf("there already registered method %s for %s", route.Method, route.Path)
			}

			endpoints = append(result, currentEndpoint)
			sort.Slice(endpoints, func(i, j int) bool {
				iLen := len(endpoints[i].Params)
				jLen := len(endpoints[j].Params)
				if iLen == jLen {
					iParamType := getLeastStrictParamType(endpoints[i].Params)
					jParamType := getLeastStrictParamType(endpoints[j].Params)
					return iParamType < jParamType
				}

				return iLen > jLen
			})
		} else {
			endpoints = []*endpoint{currentEndpoint}
		}

		// Save handler to route's method for current node
		currentNode.Endpoints.Set(route.Method, endpoints)
	}
	return nil
}

// matchRoute matches provided method and path with handler if it's existing
func (gb *gearbox) matchRoute(method, path []byte) (handlersChain, tst) {
	if handlers, params := gb.matchRouteAgainstRegistered(method, path); handlers != nil {
		return handlers, params
	}

	if gb.registeredFallback != nil && gb.registeredFallback.Handlers != nil {
		tst := newTST()
		return gb.registeredFallback.Handlers, tst
	}

	return nil, nil
}

func matchEndpointParams(ep *endpoint, paths [][]byte, pathIndex int) (tst, bool) {
	paramDic := newTST()
	endpointParams := ep.Params
	endpointParamsLen := len(endpointParams)
	pathsLen := len(paths)

	//matched := false
	paramIdx := 0
	for paramIdx < endpointParamsLen {
		if endpointParams[paramIdx].Type == ptMatchAll {
			// Last parameter, so we can return
			return paramDic, true
		}

		// path has ended and there is more parameters to match
		if pathIndex >= pathsLen {
			// If it's optional means this is the last parameter.
			if endpointParams[paramIdx].IsOptional {
				return paramDic, true
			}

			return nil, false
		}

		//
		if len(paths[pathIndex]) == 0 {
			pathIndex++
			continue
		}

		if endpointParams[paramIdx].Type == ptParam {
			paramDic.Set(endpointParams[paramIdx].Name, paths[pathIndex])
		} else if endpointParams[paramIdx].Type == ptRegexp {
			if match, _ := regexp.Match(endpointParams[paramIdx].Value, paths[pathIndex]); match {
				paramDic.Set(endpointParams[paramIdx].Name, paths[pathIndex])
			} else if !endpointParams[paramIdx].IsOptional {
				return nil, false
			}
		}

		paramIdx++
		pathIndex++
	}

	// There is more parts, so no match
	if pathsLen-pathIndex > 0 {
		return nil, false
	}

	return paramDic, true
}

func matchNodeEndpoints(node *routeNode, method []byte, paths [][]byte,
	pathIndex int, result *matchParamsResult, wg *sync.WaitGroup) {
	if endpoints, ok := node.Endpoints.Get(method).([]*endpoint); ok {
		for j := range endpoints {
			if params, matched := matchEndpointParams(endpoints[j], paths, pathIndex); matched {
				result.Matched = true
				result.Params = params
				result.Handlers = endpoints[j].Handlers
				wg.Done()
				return
			}
		}
	}

	result.Matched = false
	wg.Done()
}

func (gb *gearbox) matchRouteAgainstRegistered(method, path []byte) (handlersChain, tst) {
	// Start with root node
	currentNode := gb.routingTreeRoot

	// Return if root is empty, or path is not valid
	if currentNode == nil || len(path) == 0 || path[0] != '/' {
		return nil, nil
	}

	if !gb.settings.CaseSensitive {
		path = bytes.ToLower(path)
	}

	trimmedPath := trimPath(path)

	// Try to get from cache
	cacheKey := append(method, trimmedPath...)
	if cacheResult, ok := gb.cache.Get(cacheKey).(*matchParamsResult); ok {
		return cacheResult.Handlers, cacheResult.Params
	}

	paths := bytes.Split(trimmedPath, []byte("/"))

	var wg sync.WaitGroup
	lastMatchedNodes := []*matchParamsResult{{}}
	lastMatchedNodesIndex := 1
	wg.Add(1)
	go matchNodeEndpoints(currentNode, method, paths, 0, lastMatchedNodes[0], &wg)

	for i := range paths {
		if len(paths[i]) == 0 {
			continue
		}

		// Try to match part with a child of current node
		pathNode, ok := currentNode.Children.Get(paths[i]).(*routeNode)
		if !ok {
			break
		}

		// set matched node as current node
		currentNode = pathNode

		v := matchParamsResult{}
		lastMatchedNodes = append(lastMatchedNodes, &v)
		wg.Add(1)
		go matchNodeEndpoints(currentNode, method, paths, i+1, &v, &wg)
		lastMatchedNodesIndex++
	}

	wg.Wait()

	// Return longest prefix match
	for i := lastMatchedNodesIndex - 1; i >= 0; i-- {
		if lastMatchedNodes[i].Matched {
			go func(key []byte, matchResult *matchParamsResult) {
				gb.cache.Set(key, matchResult)
			}(append(make([]byte, 0, len(cacheKey)), cacheKey...), lastMatchedNodes[i])

			return lastMatchedNodes[i].Handlers, lastMatchedNodes[i].Params
		}
	}

	return nil, nil
}
