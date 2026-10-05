package main

import (
	"reflect"
	"testing"

	"github.com/voxire/onion/internal/platform/server/http"
)

func TestModuleRegistrarsMatchDeclaredCapabilities(t *testing.T) {
	modules := &apiModules{}
	moduleType := reflect.TypeOf(*modules)
	registries := []struct {
		name       string
		contract   reflect.Type
		registered []reflect.Type
	}{
		{
			name:       "oneKit APIs",
			contract:   reflect.TypeOf((*http.APIRegistrar)(nil)).Elem(),
			registered: registrarTypes(modules.apiRegistrars()),
		},
		{
			name:       "chi routes",
			contract:   reflect.TypeOf((*http.RouteRegistrar)(nil)).Elem(),
			registered: registrarTypes(modules.routeRegistrars()),
		},
		{
			name:       "anonymous routes",
			contract:   reflect.TypeOf((*http.AnonymousRouteProvider)(nil)).Elem(),
			registered: registrarTypes(modules.anonymousRouteProviders()),
		},
	}
	for _, registry := range registries {
		want := make(map[reflect.Type]bool)
		for i := 0; i < moduleType.NumField(); i++ {
			module := moduleType.Field(i).Type
			if module.Implements(registry.contract) {
				want[module] = true
			}
		}
		got := make(map[reflect.Type]bool)
		for _, module := range registry.registered {
			got[module] = true
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("%s registry types differ from apiModules capability implementations: got %d, want %d", registry.name, len(got), len(want))
		}
	}
}

func registrarTypes[T any](registrars []T) []reflect.Type {
	types := make([]reflect.Type, 0, len(registrars))
	for _, registrar := range registrars {
		types = append(types, reflect.TypeOf(registrar))
	}
	return types
}
