package main

import (
	"context"
	"database/sql"
	stdhttp "net/http"
	"strings"
	"testing"

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
	http.AutoRegisterAPIs(ctx, mux, modules.apiRegistrars()...)
	routes := http.NewLegacyRoutes()
	http.AutoRegisterRoutes(ctx, routes, modules.routeRegistrars()...)
	routes.Mount(mux, nil)

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
