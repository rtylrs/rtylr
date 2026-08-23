package workspaces

import (
	"net/http"
	"testing"
)

// TestRegisterAPIDoesNotPanic is the regression test for the boot panic. The
// generated Register*ServiceServer helpers take `any` and assert the server
// interface at runtime, so an API type that has fallen behind its contract
// compiles cleanly and only fails here, when the module registers it. The
// panic names the first service that fails and hides the rest, so this ran
// green for two separate rounds of missing handlers.
//
// A nil service is enough: registration inspects the interface and wires
// routes, it never calls a method. That also lets this run without a database.
func TestRegisterAPIDoesNotPanic(t *testing.T) {
	m := &Module{}
	mux := http.NewServeMux()

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("RegisterAPI panicked: %v", r)
		}
	}()

	m.RegisterAPI(mux)
}
