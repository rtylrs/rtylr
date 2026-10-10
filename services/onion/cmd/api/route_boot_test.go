package main

import (
	"context"
	"database/sql"
	stdhttp "net/http"
	"strings"
	"testing"

	"github.com/voxire/onion/internal/platform/cdn"
	platformConfig "github.com/voxire/onion/internal/platform/config"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/server/http"
)

// Every module's routes must register onto the shared mux without a
// duplicate or conflicting pattern: either one panics at server start.
func TestAllRoutesMountWithoutConflicts(t *testing.T) {
	ctx := context.Background()
	// An unconnected handle: modules keep it, nothing here queries it.
	db, err := sql.Open("mysql", "test:test@tcp(127.0.0.1:1)/none")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	dbclient.UseForTesting(db)
	previous := platformConfig.GlobalConfig
	platformConfig.GlobalConfig = &platformConfig.AppConfig{}
	t.Cleanup(func() { platformConfig.GlobalConfig = previous })
	modules := initializeModules(ctx)

	mux := stdhttp.NewServeMux()
	// The platform and server routes share the root mux with every module.
	noop := stdhttp.HandlerFunc(func(stdhttp.ResponseWriter, *stdhttp.Request) {})
	for _, pattern := range []string{"GET /metrics", "GET /health", "GET /{$}", "GET /favicon.ico"} {
		mux.Handle(pattern, noop)
	}
	http.RegisterCacheRoutes(ctx, mux)
	http.RegisterStatusRoutes(ctx, mux)
	http.RegisterVersionRoutes(mux)
	cdn.New(nil, nil, "", 0, "", "").RegisterRoutes(mux)
	http.AutoRegisterAPIs(ctx, mux, modules.apiRegistrars()...)
	routes := http.NewLegacyRoutes()
	http.AutoRegisterRoutes(ctx, routes, modules.routeRegistrars()...)
	registerCompatContracts(mux, routes)
	routes.Mount(mux, compatRoutes().Contains)

	// A compat route with no handler in the table would answer 404, and one
	// whose handler runs inside route middlewares (auth guards, rate limits)
	// would lose them: it must stay mounted from the table.
	for route := range compatRoutes() {
		method, pattern, _ := strings.Cut(route, " ")
		if !routes.Has(method, pattern) {
			t.Errorf("compat contract serves %s, but no legacy handler is registered for it", route)
		}
		if routes.HasMiddlewares(method, pattern) {
			t.Errorf("compat contract serves %s, which runs inside route middlewares the contract would skip", route)
		}
	}

	if got := len(routes.Routes()) + len(routes.RawRoutes()); got == 0 {
		t.Fatal("no legacy routes registered")
	}
}

// The standard mux treats a trailing slash as a subtree match and supports
// only {name} wildcards; chi treated both differently. No legacy route may
// depend on either difference.
func TestLegacyRoutePatternsMeanTheSameOnTheStandardMux(t *testing.T) {
	ctx := context.Background()
	db, err := sql.Open("mysql", "test:test@tcp(127.0.0.1:1)/none")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	dbclient.UseForTesting(db)
	previous := platformConfig.GlobalConfig
	platformConfig.GlobalConfig = &platformConfig.AppConfig{}
	t.Cleanup(func() { platformConfig.GlobalConfig = previous })

	routes := http.NewLegacyRoutes()
	http.AutoRegisterRoutes(ctx, routes, initializeModules(ctx).routeRegistrars()...)
	for _, route := range append(routes.Routes(), routes.RawRoutes()...) {
		pattern := route[strings.IndexByte(route, ' ')+1:]
		if len(pattern) > 1 && strings.HasSuffix(pattern, "/") {
			t.Errorf("%s ends in a slash, which the standard mux treats as a subtree", route)
		}
		if strings.Contains(pattern, "*") || strings.Contains(pattern, ":") {
			t.Errorf("%s uses a chi-only pattern feature", route)
		}
	}
}
