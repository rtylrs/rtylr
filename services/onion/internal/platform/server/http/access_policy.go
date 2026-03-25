package http

import "sync"

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

	_, ok := p.anonymousRoutes[path]

	return ok
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

	_, ok = routes[path]

	return ok
}
