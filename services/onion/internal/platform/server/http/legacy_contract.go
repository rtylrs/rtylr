package http

import (
	"bytes"
	"context"
	"encoding/json"
	stdErrors "errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"

	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
)

// errInvalidLegacyBody marks a body that is not one JSON value, so callers can
// answer with the same "Invalid request" the old router helpers sent.
var errInvalidLegacyBody = stdErrors.New("invalid request body")

// LegacyRoutes is the route table for handlers written against the
// Register* helpers (a Validate/Serve pair with Go request and response
// structs). Modules register into it instead of a router. Each route is then
// served by a OneKit contract method that dispatches to it by verb and path
// (Serve), or, until its contract exists, mounted directly on the shared
// ServeMux (Mount).
type LegacyRoutes struct {
	mu     sync.Mutex
	routes map[string]*legacyRoute
	raw    map[string]rawRoute
}

type rawRoute struct {
	method  string
	pattern string
	handler http.Handler
}

type legacyRoute struct {
	method      string
	pattern     string
	serve       func(ctx context.Context, r *http.Request, body json.RawMessage) (json.RawMessage, error)
	middlewares []MiddlewareFunc
}

// LegacyRouter is what the Register* helpers register into. Raw handlers
// (webhooks, redirects, downloads, sockets) register with Get/Post/... and are
// mounted on the shared ServeMux as they are.
type LegacyRouter interface {
	handleLegacy(method, pattern string, serve func(context.Context, *http.Request, json.RawMessage) (json.RawMessage, error), middlewares ...MiddlewareFunc)
	Method(method, pattern string, handler http.Handler)
	Get(pattern string, handler http.HandlerFunc)
	Post(pattern string, handler http.HandlerFunc)
	Put(pattern string, handler http.HandlerFunc)
	Patch(pattern string, handler http.HandlerFunc)
	Delete(pattern string, handler http.HandlerFunc)
	// With returns a router whose raw handlers run inside the middlewares.
	With(middlewares ...func(http.Handler) http.Handler) LegacyRouter
}

func NewLegacyRoutes() *LegacyRoutes {
	return &LegacyRoutes{routes: map[string]*legacyRoute{}, raw: map[string]rawRoute{}}
}

func legacyRouteKey(method, pattern string) string {
	return method + " " + normalizeLegacyPattern(pattern)
}

// normalizeLegacyPattern drops path parameter names, so /x/{id} and
// /x/{lead_id} name the same route.
func normalizeLegacyPattern(pattern string) string {
	var out strings.Builder
	depth := 0
	for _, r := range pattern {
		switch {
		case r == '{':
			depth++
			out.WriteString("{}")
		case r == '}':
			depth--
		case depth == 0:
			out.WriteRune(r)
		}
	}
	return out.String()
}

func (l *LegacyRoutes) handleLegacy(method, pattern string, serve func(context.Context, *http.Request, json.RawMessage) (json.RawMessage, error), middlewares ...MiddlewareFunc) {
	l.mu.Lock()
	defer l.mu.Unlock()
	key := legacyRouteKey(method, pattern)
	if _, exists := l.routes[key]; exists {
		panic(fmt.Errorf("duplicate legacy route %s", key))
	}
	if _, exists := l.raw[key]; exists {
		panic(fmt.Errorf("duplicate raw route %s", key))
	}
	l.routes[key] = &legacyRoute{method: method, pattern: pattern, serve: serve, middlewares: middlewares}
}

func (l *LegacyRoutes) Method(method, pattern string, handler http.Handler) {
	l.mu.Lock()
	defer l.mu.Unlock()
	key := legacyRouteKey(method, pattern)
	if _, exists := l.raw[key]; exists {
		panic(fmt.Errorf("duplicate raw route %s", key))
	}
	if _, exists := l.routes[key]; exists {
		panic(fmt.Errorf("duplicate legacy route %s", key))
	}
	l.raw[key] = rawRoute{method: method, pattern: pattern, handler: handler}
}

func (l *LegacyRoutes) Get(pattern string, handler http.HandlerFunc) {
	l.Method(http.MethodGet, pattern, handler)
}

func (l *LegacyRoutes) Post(pattern string, handler http.HandlerFunc) {
	l.Method(http.MethodPost, pattern, handler)
}

func (l *LegacyRoutes) Put(pattern string, handler http.HandlerFunc) {
	l.Method(http.MethodPut, pattern, handler)
}

func (l *LegacyRoutes) Patch(pattern string, handler http.HandlerFunc) {
	l.Method(http.MethodPatch, pattern, handler)
}

func (l *LegacyRoutes) Delete(pattern string, handler http.HandlerFunc) {
	l.Method(http.MethodDelete, pattern, handler)
}

func (l *LegacyRoutes) With(middlewares ...func(http.Handler) http.Handler) LegacyRouter {
	return &legacyRouterWith{parent: l, middlewares: middlewares}
}

// legacyRouterWith wraps raw handlers in middlewares; JSON handlers keep the
// middlewares they were registered with.
type legacyRouterWith struct {
	parent      LegacyRouter
	middlewares []func(http.Handler) http.Handler
}

func (w *legacyRouterWith) handleLegacy(method, pattern string, serve func(context.Context, *http.Request, json.RawMessage) (json.RawMessage, error), middlewares ...MiddlewareFunc) {
	wrapped := make([]MiddlewareFunc, 0, len(w.middlewares)+len(middlewares))
	for _, mw := range w.middlewares {
		wrapped = append(wrapped, MiddlewareFunc(mw))
	}
	w.parent.handleLegacy(method, pattern, serve, append(wrapped, middlewares...)...)
}

func (w *legacyRouterWith) Method(method, pattern string, handler http.Handler) {
	for i := len(w.middlewares) - 1; i >= 0; i-- {
		handler = w.middlewares[i](handler)
	}
	w.parent.Method(method, pattern, handler)
}

func (w *legacyRouterWith) Get(pattern string, handler http.HandlerFunc) {
	w.Method(http.MethodGet, pattern, handler)
}

func (w *legacyRouterWith) Post(pattern string, handler http.HandlerFunc) {
	w.Method(http.MethodPost, pattern, handler)
}

func (w *legacyRouterWith) Put(pattern string, handler http.HandlerFunc) {
	w.Method(http.MethodPut, pattern, handler)
}

func (w *legacyRouterWith) Patch(pattern string, handler http.HandlerFunc) {
	w.Method(http.MethodPatch, pattern, handler)
}

func (w *legacyRouterWith) Delete(pattern string, handler http.HandlerFunc) {
	w.Method(http.MethodDelete, pattern, handler)
}

func (w *legacyRouterWith) With(middlewares ...func(http.Handler) http.Handler) LegacyRouter {
	return &legacyRouterWith{parent: w, middlewares: middlewares}
}

// RawRoutes lists raw handler routes as "METHOD /pattern", sorted.
func (l *LegacyRoutes) RawRoutes() []string {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make([]string, 0, len(l.raw))
	for _, route := range l.raw {
		out = append(out, route.method+" "+route.pattern)
	}
	sort.Strings(out)
	return out
}

func (l *LegacyRoutes) route(method, pattern string) (*legacyRoute, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	route, ok := l.routes[legacyRouteKey(method, pattern)]
	return route, ok
}

// Routes lists every registered route as "METHOD /pattern", sorted.
func (l *LegacyRoutes) Routes() []string {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make([]string, 0, len(l.routes))
	for _, route := range l.routes {
		out = append(out, route.method+" "+route.pattern)
	}
	sort.Strings(out)
	return out
}

// Serve runs the route registered for method and pattern on behalf of a
// contract method and returns its JSON response unchanged. contractPattern is
// the contract's route; its path parameters are made available to the
// handler under the names the handler registered with.
func (l *LegacyRoutes) Serve(ctx context.Context, r *http.Request, method, contractPattern string, body json.RawMessage) (json.RawMessage, error) {
	route, ok := l.route(method, contractPattern)
	if !ok {
		return nil, onekiterrors.New(http.StatusNotFound, "Not found")
	}
	if len(route.middlewares) > 0 {
		// Route middlewares (auth guards, rate limits) are HTTP handlers and
		// cannot run inside a contract method. Such a route must stay mounted
		// from the table; refusing here keeps a guard from being bypassed.
		logger.Ctx(ctx).Error("Contract dispatch refused for a route with middlewares", logger.String("route", method+" "+contractPattern))
		return nil, onekiterrors.New(http.StatusInternalServerError, "Internal server error")
	}
	ctx = withLegacyPathValues(ctx, r, route.pattern, contractPattern)
	// Contracts cannot declare a DELETE body, but legacy DELETEs accept one
	// ({"id": N}); read it here when the client sent it.
	if body == nil && r != nil && r.Method == http.MethodDelete && r.Body != nil && r.ContentLength != 0 {
		raw, err := io.ReadAll(io.LimitReader(r.Body, DefaultMaxRequestBodyBytes+1))
		if err != nil {
			return nil, onekiterrors.New(http.StatusBadRequest, "Invalid request")
		}
		if int64(len(raw)) > DefaultMaxRequestBodyBytes {
			return nil, onekiterrors.New(http.StatusRequestEntityTooLarge, "Request body too large")
		}
		body = raw
	}
	out, err := route.serve(ctx, r, body)
	if stdErrors.Is(err, errInvalidLegacyBody) {
		return nil, onekiterrors.New(http.StatusBadRequest, "Invalid request")
	}
	return out, err
}

// Mount serves every route the skip function does not claim straight from the
// shared ServeMux, writing responses exactly as the old router helpers did.
func (l *LegacyRoutes) Mount(mux *http.ServeMux, skip func(route string) bool) {
	for _, name := range l.Routes() {
		if skip != nil && skip(name) {
			continue
		}
		method, pattern, _ := strings.Cut(name, " ")
		route, _ := l.route(method, pattern)
		var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			body, err := readLegacyBody(w, r)
			if err != nil {
				if isRequestBodyTooLarge(err) {
					writeRequestEntityTooLarge(ctx, w)
				} else {
					writeBadRequest(ctx, w)
				}
				return
			}
			out, err := route.serve(withLegacyPathValues(ctx, r, route.pattern, route.pattern), r, body)
			if err != nil {
				if stdErrors.Is(err, errInvalidLegacyBody) {
					writeBadRequest(ctx, w)
					return
				}
				WriteOneKitError(ctx, w, err)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write(append(out, '\n')); err != nil {
				logger.Ctx(ctx).Error("Failed to encode response", logger.Error(err))
			}
		})
		for i := len(route.middlewares) - 1; i >= 0; i-- {
			handler = route.middlewares[i](handler)
		}
		mux.Handle(method+" "+pattern, handler)
	}
	l.mu.Lock()
	raw := make([]rawRoute, 0, len(l.raw))
	for _, route := range l.raw {
		raw = append(raw, route)
	}
	l.mu.Unlock()
	for _, route := range raw {
		if skip != nil && skip(route.method+" "+route.pattern) {
			continue
		}
		mux.Handle(route.method+" "+route.pattern, route.handler)
	}
}

// readLegacyBody reads the body the way the old helpers decided whether to
// decode one: always for POST/PUT/PATCH, and for DELETE only when sent.
func readLegacyBody(w http.ResponseWriter, r *http.Request) (json.RawMessage, error) {
	switch r.Method {
	case http.MethodPost, http.MethodPut, http.MethodPatch:
	case http.MethodDelete:
		if r.Body == nil || r.ContentLength == 0 {
			return nil, nil
		}
	default:
		return nil, nil
	}
	if r.Body == nil {
		return nil, nil
	}
	reader := r.Body
	if !requestBodyLimitExempt(r.URL.Path) {
		reader = http.MaxBytesReader(w, r.Body, DefaultMaxRequestBodyBytes)
	}
	return io.ReadAll(reader)
}

func (r legacyRegistration[Req, Res]) serve(ctx context.Context, req *http.Request, body json.RawMessage) (json.RawMessage, error) {
	return ServeLegacy(ctx, req, body, r.handler, r.validate)
}

type legacyRegistration[Req, Res any] struct {
	handler  HandlerFunc[Req, Res]
	validate bool
}

func registerLegacy[Req, Res any](router LegacyRouter, method string, paths []string, handler HandlerFunc[Req, Res], validate bool, middlewares ...MiddlewareFunc) {
	registration := legacyRegistration[Req, Res]{handler: handler, validate: validate}
	for _, path := range paths {
		router.handleLegacy(method, path, registration.serve, middlewares...)
	}
}

type legacyRequestKey struct{}

type legacyPathValuesKey struct{}

// withLegacyPathValues maps the handler's path parameter names onto the
// values of the route actually matched, by position.
func withLegacyPathValues(ctx context.Context, r *http.Request, legacyPattern, matchedPattern string) context.Context {
	legacyNames, matchedNames := pathParamNames(legacyPattern), pathParamNames(matchedPattern)
	if len(legacyNames) == 0 || len(legacyNames) != len(matchedNames) || r == nil {
		return ctx
	}
	values := make(map[string]string, len(legacyNames))
	for i, name := range legacyNames {
		values[name] = r.PathValue(matchedNames[i])
	}
	return context.WithValue(ctx, legacyPathValuesKey{}, values)
}

func pathParamNames(pattern string) []string {
	var names []string
	for {
		start := strings.IndexByte(pattern, '{')
		if start < 0 {
			return names
		}
		end := strings.IndexByte(pattern[start:], '}')
		if end < 0 {
			return names
		}
		names = append(names, strings.TrimSuffix(pattern[start+1:start+end], "..."))
		pattern = pattern[start+end+1:]
	}
}

// ServeLegacy runs a Validate/Serve handler and returns its response JSON
// unchanged, so moving a route onto a contract keeps its wire format exactly.
//
// The request binds the way the old router helpers bound it: GET reads query
// parameters, POST/PUT/PATCH read the JSON body, and DELETE reads the body
// when one was sent and query parameters otherwise.
func ServeLegacy[Req, Res any](ctx context.Context, r *http.Request, body json.RawMessage, handler HandlerFunc[Req, Res], validate bool) (json.RawMessage, error) {
	if r == nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "request is unavailable")
	}
	ctx = context.WithValue(ctx, legacyRequestKey{}, r)

	var req Req
	useBody := false
	switch r.Method {
	case http.MethodPost, http.MethodPut, http.MethodPatch:
		useBody = true
	case http.MethodDelete:
		useBody = len(bytes.TrimSpace(body)) > 0
	}
	if useBody {
		if err := decodeLegacyBody(body, &req); err != nil {
			return nil, errInvalidLegacyBody
		}
	} else if err := parseQueryParams(r, &req); err != nil {
		return nil, onekiterrors.New(http.StatusBadRequest, err.Error())
	}

	if validate {
		if err := validateRequest(ctx, handler, &req); err != nil {
			return nil, err
		}
	}
	res, err := handler.Serve(ctx, &req)
	if err != nil {
		return nil, err
	}
	if responseProtoValidationEnabled() {
		if err := validateProtoMessages(res); err != nil {
			logger.Ctx(ctx).Error("Response proto validation failed", logger.Error(err))
			return nil, onekiterrors.New(http.StatusInternalServerError, "Response validation failed")
		}
	}
	return json.Marshal(res)
}

// decodeLegacyBody reads at most one JSON value; an empty body is accepted so
// no-payload endpoints keep working.
func decodeLegacyBody(body json.RawMessage, dst any) error {
	trimmed := bytes.TrimSpace(body)
	if len(trimmed) == 0 {
		return nil
	}
	decoder := json.NewDecoder(bytes.NewReader(trimmed))
	if err := decoder.Decode(dst); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !stdErrors.Is(err, io.EOF) {
		return errMultipleJSONBodyValues
	}
	return nil
}

// PathParam returns a path value such as {id} for a handler in the legacy
// route table, replacing chi's URL parameter lookup.
func PathParam(ctx context.Context, name string) string {
	if values, ok := ctx.Value(legacyPathValuesKey{}).(map[string]string); ok {
		if value, found := values[name]; found {
			return value
		}
	}
	if r, ok := ctx.Value(legacyRequestKey{}).(*http.Request); ok && r != nil {
		return r.PathValue(name)
	}
	return ""
}

// LegacyRequest returns the HTTP request a legacy handler is serving, for the
// few handlers that read headers or the raw URL.
func LegacyRequest(ctx context.Context) (*http.Request, bool) {
	r, ok := ctx.Value(legacyRequestKey{}).(*http.Request)
	return r, ok && r != nil
}

// WithPathParams sets the path values a legacy handler reads through
// PathParam. Tests use it to call a handler's Serve directly.
func WithPathParams(ctx context.Context, values map[string]string) context.Context {
	return context.WithValue(ctx, legacyPathValuesKey{}, values)
}

// Handler returns a ServeMux serving every registered route, for tests and
// for runs that do not register contracts.
func (l *LegacyRoutes) Handler() http.Handler {
	mux := http.NewServeMux()
	l.Mount(mux, nil)
	return mux
}

// ServeHTTP serves one request against every registered route. It builds a
// fresh mux per call, so it is meant for tests; servers Mount the table once.
func (l *LegacyRoutes) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	l.Handler().ServeHTTP(w, r)
}

// HasMiddlewares reports whether a registered JSON route runs inside route
// middlewares, which only the table's own mount applies.
func (l *LegacyRoutes) HasMiddlewares(method, pattern string) bool {
	route, ok := l.route(method, pattern)
	return ok && len(route.middlewares) > 0
}

// Has reports whether a route is registered for method and pattern (path
// parameter names are ignored).
func (l *LegacyRoutes) Has(method, pattern string) bool {
	_, ok := l.route(method, pattern)
	return ok
}

// RouteSet is a set of "METHOD /pattern" routes, matched ignoring path
// parameter names.
type RouteSet map[string]struct{}

func NewRouteSet(routes ...[]string) RouteSet {
	set := RouteSet{}
	for _, group := range routes {
		for _, route := range group {
			method, pattern, _ := strings.Cut(route, " ")
			set[legacyRouteKey(method, pattern)] = struct{}{}
		}
	}
	return set
}

func (s RouteSet) Contains(route string) bool {
	method, pattern, _ := strings.Cut(route, " ")
	_, ok := s[legacyRouteKey(method, pattern)]
	return ok
}

// WriteContractError reports a generated contract server's error the way the
// legacy helpers did. Server failures go through WriteOneKitError, which logs
// unhandled errors; client errors (bad parameters, failed validation, a
// handler's 4xx) keep the status and public message the server chose.
func WriteContractError(w http.ResponseWriter, r *http.Request, status int, message string, violations []string, cause error) {
	if cause != nil && status >= http.StatusInternalServerError {
		WriteOneKitError(r.Context(), w, cause)
		return
	}
	body := map[string]any{"message": message}
	if len(violations) > 0 {
		body["violations"] = violations
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		logger.Ctx(r.Context()).Error("Failed to write error response", logger.Error(err))
	}
}
