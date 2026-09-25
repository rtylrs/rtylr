package jobs

import (
	"context"
	"reflect"

	"github.com/voxire/onion/internal/platform/logger"
)

type Registrar interface {
	RegisterJobs(context.Context, *Registry)
}

func AutoRegister(ctx context.Context, modules any) {
	AutoRegisterInto(ctx, modules, defaultRegistry)
}

func AutoRegisterInto(ctx context.Context, modules any, registry *Registry) {
	if modules == nil || registry == nil {
		return
	}
	registerRecursive(ctx, reflect.ValueOf(modules), registry)
}

func registerRecursive(ctx context.Context, val reflect.Value, registry *Registry) {
	if !val.IsValid() {
		return
	}
	if val.Kind() == reflect.Interface {
		val = val.Elem()
	}
	if val.Kind() != reflect.Pointer || val.IsNil() {
		return
	}

	if registrar, ok := val.Interface().(Registrar); ok {
		registrar.RegisterJobs(ctx, registry)
		return
	}

	elem := val.Elem()
	if elem.Kind() != reflect.Struct {
		return
	}

	typ := elem.Type()
	for i := 0; i < elem.NumField(); i++ {
		field := elem.Field(i)
		if !field.CanInterface() || !isNilableAndNonNil(field) {
			continue
		}

		before := len(registry.Names())
		registerRecursive(ctx, field, registry)
		after := len(registry.Names())
		if after > before {
			logger.Ctx(ctx).Debug("registered module jobs",
				logger.String("module", typ.Field(i).Name),
				logger.Int("count", after-before),
			)
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
