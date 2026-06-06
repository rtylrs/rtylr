package http

import (
	"strings"
	"sync"
)

type RouteAccessPolicy struct {
	mu              sync.RWMutex
	anonymousRoutes map[string]struct{}
	agentRoutes     map[string]map[string]struct{}
}

func NewRouteAccessPolicy() *RouteAccessPolicy {
	return &RouteAccessPolicy{
		anonymousRoutes: make(map[string]struct{}),
		agentRoutes:     make(map[string]map[string]struct{}),
	}
}

func (p *RouteAccessPolicy) RegisterAnonymousRoutes(routes ...string) {
	if p == nil {
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	for _, route := range routes {
		if route == "" {
			continue
		}

		p.anonymousRoutes[route] = struct{}{}
	}
}

func (p *RouteAccessPolicy) RegisterAgentRoutes(userAgent string, routes ...string) {
	if p == nil || userAgent == "" {
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if _, ok := p.agentRoutes[userAgent]; !ok {
		p.agentRoutes[userAgent] = make(map[string]struct{})
	}

	for _, route := range routes {
		if route == "" {
			continue
		}

		p.agentRoutes[userAgent][route] = struct{}{}
	}
}

func (p *RouteAccessPolicy) IsAllowedAnonymously(path string) bool {
	if p == nil {
		return false
	}

	p.mu.RLock()
	defer p.mu.RUnlock()

	return hasMatchingRoute(p.anonymousRoutes, path)
}

func (p *RouteAccessPolicy) IsAllowedForAgent(userAgent, path string) bool {
	if p == nil || userAgent == "" {
		return false
	}

	p.mu.RLock()
	defer p.mu.RUnlock()

	routes, ok := p.agentRoutes[userAgent]
	if !ok {
		return false
	}

	return hasMatchingRoute(routes, path)
}

func hasMatchingRoute(routes map[string]struct{}, path string) bool {
	if _, ok := routes[path]; ok {
		return true
	}

	for route := range routes {
		if routePatternMatchesPath(route, path) {
			return true
		}
	}
	return false
}

func routePatternMatchesPath(route, path string) bool {
	if !strings.Contains(route, "{") {
		return false
	}

	routeParts := splitRoutePath(route)
	pathParts := splitRoutePath(path)
	if len(routeParts) != len(pathParts) {
		return false
	}

	for i, routePart := range routeParts {
		if strings.HasPrefix(routePart, "{") && strings.HasSuffix(routePart, "}") {
			if pathParts[i] == "" {
				return false
			}
			continue
		}
		if routePart != pathParts[i] {
			return false
		}
	}
	return true
}

func splitRoutePath(path string) []string {
	return strings.Split(strings.Trim(path, "/"), "/")
}
