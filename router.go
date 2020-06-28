package gearbox

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"
)

type routeNode struct {
	Name      string
	Endpoints map[string][]*endpoint
	Children  map[string]*routeNode
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
	Name       string
	Value      string
	Type       paramType
	IsOptional bool
}

type endpoint struct {
	Params   []*param
	Handlers handlersChain
}

type routerFallback struct {
	Handlers handlersChain
}

type matchParamsResult struct {
	Matched  bool
	Handlers handlersChain
	Params   map[string]string
}

// validateRoutePath makes sure that path complies with path's rules
func validateRoutePath(path string) error {
	// Check length of the path
	length := len(path)
	if length == 0 {
		return fmt.Errorf("length is zero")
	}

	// Make sure path starts with /
	if path[0] != '/' {
		return fmt.Errorf("path must start with /")
	}

	params := make(map[string]bool)
	parts := strings.Split(trimPath(path), "/")
	partsLen := len(parts)
	for i := 0; i < partsLen; i++ {
		if parts[i] == "" {
			continue
		}
		if p := parseParameter(parts[i]); p != nil {
			if p.Type == ptMatchAll && i != partsLen-1 {
				return fmt.Errorf("* must be in the end of path")
			} else if p.IsOptional && i != partsLen-1 {
				return fmt.Errorf("only last parameter can be optional")
			} else if p.Type == ptParam || p.Type == ptRegexp {
				if _, ok := params[p.Name]; ok {
					return fmt.Errorf("parameter is duplicated")
				}
				params[p.Name] = true
			}
		}
	}

	return nil
}

// registerRoute registers handler with method and path
func (gb *gearbox) registerRoute(method, path string, handlers handlersChain) *Route {

	if gb.settings.CaseSensitive {
		path = strings.ToLower(path)
	}

	route := &Route{
		Path:     path,
		Method:   method,
		Handlers: handlers,
	}

	// Add route to registered routes
	gb.registeredRoutes = append(gb.registeredRoutes, route)
	return route
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
func createEmptyRouteNode(name string) *routeNode {
	return &routeNode{
		Name:      name,
		Children:  make(map[string]*routeNode),
		Endpoints: make(map[string][]*endpoint),
	}
}

// parseParameter parses path part into param struct, or returns nil if it's
// not a parameter
func parseParameter(pathPart string) *param {
	pathPartLen := len(pathPart)
	if pathPartLen == 0 {
		return nil
	}

	// match all
	if pathPart[0] == '*' {
		return &param{
			Name: "*",
			Type: ptMatchAll,
		}
	}

	isOptional := pathPart[pathPartLen-1] == '?'
	if isOptional {
		pathPart = pathPart[0 : pathPartLen-1]
	}

	params := strings.Split(pathPart, ":")
	paramsLen := len(params)

	if paramsLen == 2 && params[0] == "" { // Just a parameter
		return &param{
			Name:       params[1],
			Type:       ptParam,
			IsOptional: isOptional,
		}
	} else if paramsLen == 3 && params[0] == "" { // Regex parameter
		return &param{
			Name:       params[1],
			Value:      params[2],
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
func trimPath(path string) string {
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
	gb.routingTreeRoot = createEmptyRouteNode("root")

	for _, route := range gb.registeredRoutes {
		currentNode := gb.routingTreeRoot

		// Handler is not provided
		if route.Handlers == nil {
			return fmt.Errorf("route %s with method %s does not contain any handlers", route.Path, route.Method)
		}

		// Check if path is valid or not
		if err := validateRoutePath(route.Path); err != nil {
			return fmt.Errorf("route %s is not valid! - %s", route.Path, err.Error())
		}

		params := make([]*param, 0)

		// Split path into slices of parts
		parts := strings.Split(route.Path, "/")

		partsLen := len(parts)
		for i := 1; i < partsLen; i++ {
			part := parts[i]

			// Do not create node if part is empty
			if part == "" {
				continue
			}

			// Parse part as a parameter if it is
			if param := parseParameter(part); param != nil {
				params = append(params, param)
				continue
			}

			// Try to get a child of current node with part, otherwise
			//creates a new node and make it current node
			partNode, ok := currentNode.Children[part]
			if !ok {
				partNode = createEmptyRouteNode(part)
				currentNode.Children[part] = partNode
			}
			currentNode = partNode
		}

		currentEndpoint := &endpoint{
			Handlers: route.Handlers,
			Params:   params,
		}

		// Make sure that current node does not have a handler for route's method
		var endpoints []*endpoint
		if result, ok := currentNode.Endpoints[route.Method]; ok {
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
		currentNode.Endpoints[route.Method] = endpoints
	}
	return nil
}

// matchRoute matches provided method and path with handler if it's existing
func (gb *gearbox) matchRoute(method, path string) (handlersChain, map[string]string) {
	if handlers, params := gb.matchRouteAgainstRegistered(method, path); handlers != nil {
		return handlers, params
	}

	if gb.registeredFallback != nil && gb.registeredFallback.Handlers != nil {
		return gb.registeredFallback.Handlers, make(map[string]string)
	}

	return nil, nil
}

func matchEndpointParams(ep *endpoint, paths []string, pathIndex int) (map[string]string, bool) {
	endpointParams := ep.Params
	endpointParamsLen := len(endpointParams)
	pathsLen := len(paths)
	paramDic := make(map[string]string, endpointParamsLen)

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

		if paths[pathIndex] == "" {
			pathIndex++
			continue
		}

		if endpointParams[paramIdx].Type == ptParam {
			paramDic[endpointParams[paramIdx].Name] = paths[pathIndex]
		} else if endpointParams[paramIdx].Type == ptRegexp {
			if match, _ := regexp.MatchString(endpointParams[paramIdx].Value, paths[pathIndex]); match {
				paramDic[endpointParams[paramIdx].Name] = paths[pathIndex]
			} else if !endpointParams[paramIdx].IsOptional {
				return nil, false
			}
		}

		paramIdx++
		pathIndex++
	}

	for pathIndex < pathsLen && paths[pathIndex] == "" {
		pathIndex++
	}

	// There is more parts, so no match
	if pathsLen-pathIndex > 0 {
		return nil, false
	}

	return paramDic, true
}

func matchNodeEndpoints(node *routeNode, method string, paths []string,
	pathIndex int, result *matchParamsResult, wg *sync.WaitGroup) {
	if endpoints, ok := node.Endpoints[method]; ok {
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

func (gb *gearbox) matchRouteAgainstRegistered(method, path string) (handlersChain, map[string]string) {
	// Start with root node
	currentNode := gb.routingTreeRoot

	// Return if root is empty, or path is not valid
	if currentNode == nil || path == "" || path[0] != '/' || len(path) > defaultMaxRequestUrlLength {
		return nil, nil
	}

	if gb.settings.CaseSensitive {
		path = strings.ToLower(path)
	}

	trimmedPath := trimPath(path)

	// Try to get from cache if it's enabled
	cacheKey := ""
	if !gb.settings.DisableCaching {
		cacheKey = method + trimmedPath
		if cacheResult, ok := gb.cache.Get(cacheKey).(*matchParamsResult); ok {
			return cacheResult.Handlers, cacheResult.Params
		}
	}

	paths := strings.Split(trimmedPath, "/")

	var wg sync.WaitGroup
	lastMatchedNodes := []*matchParamsResult{{}}
	lastMatchedNodesIndex := 1
	wg.Add(1)
	go matchNodeEndpoints(currentNode, method, paths, 0, lastMatchedNodes[0], &wg)

	for i := range paths {
		if paths[i] == "" {
			continue
		}

		// Try to match part with a child of current node
		pathNode, ok := currentNode.Children[paths[i]]
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
			if !gb.settings.DisableCaching {
				go func(key string, matchResult *matchParamsResult) {
					gb.cache.Set(key, matchResult)
				}(string(cacheKey), lastMatchedNodes[i])
			}

			return lastMatchedNodes[i].Handlers, lastMatchedNodes[i].Params
		}
	}

	return nil, nil
}
