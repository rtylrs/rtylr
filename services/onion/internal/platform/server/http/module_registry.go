package http

import (
	"context"
	"reflect"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/platform/logger"
)

type ModuleRegistry struct {
	modules interface{}
}

func NewModuleRegistry(modules interface{}) *ModuleRegistry {
	return &ModuleRegistry{
		modules: modules,
	}
}

func AutoRegisterRoutes(ctx context.Context, modules interface{}, r *chi.Mux, dependencies ...interface{}) {
	registry := NewModuleRegistry(modules)
	registry.RegisterRoutes(ctx, r, dependencies...)
}

func AutoRegisterAnonymousRoutes(ctx context.Context, modules interface{}, accessPolicy *RouteAccessPolicy) {
	registry := NewModuleRegistry(modules)
	registry.RegisterAnonymousRoutes(ctx, accessPolicy)
}

func AutoConfigureAccessPolicy(ctx context.Context, modules interface{}, accessPolicy *RouteAccessPolicy) {
	registry := NewModuleRegistry(modules)
	registry.ConfigureAccessPolicy(ctx, accessPolicy)
}

func (mr *ModuleRegistry) RegisterRoutes(ctx context.Context, r *chi.Mux, dependencies ...interface{}) {
	mr.registerRoutesRecursive(ctx, reflect.ValueOf(mr.modules).Elem(), r, dependencies...)
}

func (mr *ModuleRegistry) registerRoutesRecursive(ctx context.Context, val reflect.Value, r *chi.Mux, dependencies ...interface{}) {
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldName := typ.Field(i).Name

		if !isNilableAndNonNil(field) {
			continue
		}

		registerMethod := field.MethodByName("RegisterRoutes")
		if registerMethod.IsValid() {
			logger.Ctx(ctx).Debug("Registering routes", logger.String("module", fieldName))

			args := make([]reflect.Value, 0, 2+len(dependencies))
			args = append(args,
				reflect.ValueOf(ctx),
				reflect.ValueOf(r),
			)

			for _, dep := range dependencies {
				args = append(args, reflect.ValueOf(dep))
			}

			registerMethod.Call(args)

			continue
		}

		if isPlatformStruct(field) {
			logger.Ctx(ctx).Debug("Registering platform routes", logger.String("platform", fieldName))

			mr.registerRoutesRecursive(ctx, field.Elem(), r, dependencies...)
		}
	}
}

func (mr *ModuleRegistry) RegisterAnonymousRoutes(ctx context.Context, accessPolicy *RouteAccessPolicy) {
	mr.registerAnonymousRoutesRecursive(ctx, reflect.ValueOf(mr.modules).Elem(), accessPolicy)
}

func (mr *ModuleRegistry) registerAnonymousRoutesRecursive(ctx context.Context, val reflect.Value, accessPolicy *RouteAccessPolicy) {
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldName := typ.Field(i).Name

		if !isNilableAndNonNil(field) {
			continue
		}

		getAnonymousRoutesMethod := field.MethodByName("GetAnonymousRoutes")
		if getAnonymousRoutesMethod.IsValid() {
			logger.Ctx(ctx).Debug("Registering anonymous routes", logger.String("module", fieldName))

			mr.collectAnonymousRoutes(ctx, fieldName, getAnonymousRoutesMethod, accessPolicy)

			continue
		}

		if isPlatformStruct(field) {
			logger.Ctx(ctx).Debug("Registering platform anonymous routes", logger.String("platform", fieldName))

			mr.registerAnonymousRoutesRecursive(ctx, field.Elem(), accessPolicy)
		}
	}
}

func (mr *ModuleRegistry) collectAnonymousRoutes(ctx context.Context, fieldName string, method reflect.Value, accessPolicy *RouteAccessPolicy) {
	result := method.Call([]reflect.Value{})
	if len(result) == 0 || result[0].IsNil() {
		return
	}

	routes, ok := result[0].Interface().([]string)
	if !ok {
		logger.Ctx(ctx).Warn("GetAnonymousRoutes returned unexpected type",
			logger.String("module", fieldName))

		return
	}

	accessPolicy.RegisterAnonymousRoutes(routes...)
}

func (mr *ModuleRegistry) ConfigureAccessPolicy(ctx context.Context, accessPolicy *RouteAccessPolicy) {
	mr.configureAccessPolicyRecursive(ctx, reflect.ValueOf(mr.modules).Elem(), accessPolicy)
}

func (mr *ModuleRegistry) configureAccessPolicyRecursive(ctx context.Context, val reflect.Value, accessPolicy *RouteAccessPolicy) {
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldName := typ.Field(i).Name

		if !isNilableAndNonNil(field) {
			continue
		}

		configureMethod := field.MethodByName("ConfigureAccessPolicy")
		if configureMethod.IsValid() {
			logger.Ctx(ctx).Debug("Configuring access policy", logger.String("module", fieldName))

			configureMethod.Call([]reflect.Value{
				reflect.ValueOf(ctx),
				reflect.ValueOf(accessPolicy),
			})

			continue
		}

		if isPlatformStruct(field) {
			logger.Ctx(ctx).Debug("Configuring platform access policy", logger.String("platform", fieldName))

			mr.configureAccessPolicyRecursive(ctx, field.Elem(), accessPolicy)
		}
	}
}
func isNilableAndNonNil(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Pointer, reflect.Interface, reflect.Map, reflect.Slice, reflect.Chan, reflect.Func:
		return !v.IsNil()
	default:
		return false
	}
}

func isPlatformStruct(v reflect.Value) bool {
	if v.Kind() != reflect.Pointer || v.Elem().Kind() != reflect.Struct {
		return false
	}

	// A platform is a struct that doesn't have RegisterRoutes or GetAnonymousRoutes itself
	// but contains fields that do (i.e. it's a container of modules)
	hasRegister := v.MethodByName("RegisterRoutes").IsValid()
	hasAnonymous := v.MethodByName("GetAnonymousRoutes").IsValid()

	return !hasRegister && !hasAnonymous
}
