package http

import (
	"context"
	"net/http"
	"reflect"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/platform/logger"
)

// MountOneKit attaches the shared oneKit ServeMux at the generated contract
// paths, such as /hub/user/v1/users/me. Chi's specific routes continue to take
// precedence over this wildcard, allowing generated and platform routes to
// coexist without a separate deployment prefix.
func MountOneKit(router chi.Router, mux *http.ServeMux) {
	router.Handle("/*", mux)
}

// AutoRegisterAPIs walks the modules struct and calls RegisterAPI(mux) on every
// module that exposes it. RegisterAPI is where a module registers its generated
// oneKit service servers on the shared mux.
func AutoRegisterAPIs(ctx context.Context, modules interface{}, mux *http.ServeMux) {
	registry := NewModuleRegistry(modules)
	registry.RegisterAPIs(ctx, mux)
}

func (mr *ModuleRegistry) RegisterAPIs(ctx context.Context, mux *http.ServeMux) {
	mr.registerAPIsRecursive(ctx, reflect.ValueOf(mr.modules).Elem(), mux)
}

func (mr *ModuleRegistry) registerAPIsRecursive(ctx context.Context, val reflect.Value, mux *http.ServeMux) {
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldName := typ.Field(i).Name

		if !isNilableAndNonNil(field) {
			continue
		}

		registerAPIMethod := field.MethodByName("RegisterAPI")
		if registerAPIMethod.IsValid() {
			logger.Ctx(ctx).Debug("Registering oneKit services", logger.String("module", fieldName))

			registerAPIMethod.Call([]reflect.Value{reflect.ValueOf(mux)})

			continue
		}

		if isPlatformStruct(field) {
			mr.registerAPIsRecursive(ctx, field.Elem(), mux)
		}
	}
}
