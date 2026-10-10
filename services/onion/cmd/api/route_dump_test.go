package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	platformConfig "github.com/voxire/onion/internal/platform/config"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/server/http"
)

// ROUTE_DUMP=path writes every legacy route ("json"/"raw", method, pattern)
// for migration tooling.
func TestDumpLegacyRoutes(t *testing.T) {
	out := os.Getenv("ROUTE_DUMP")
	if out == "" {
		t.Skip("set ROUTE_DUMP to write the legacy route table")
	}
	ctx := context.Background()
	db, _ := sql.Open("mysql", "test:test@tcp(127.0.0.1:1)/none")
	dbclient.UseForTesting(db)
	platformConfig.GlobalConfig = &platformConfig.AppConfig{}
	routes := http.NewLegacyRoutes()
	http.AutoRegisterRoutes(ctx, routes, initializeModules(ctx).routeRegistrars()...)
	f, err := os.Create(out)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	for _, r := range routes.Routes() {
		fmt.Fprintln(f, "json\t"+r)
	}
	for _, r := range routes.RawRoutes() {
		fmt.Fprintln(f, "raw\t"+r)
	}
}
