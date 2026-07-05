package cacheops

// Module is kept so old wiring does not break, but /invalidate-cache is
// registered by the platform HTTP layer where the raw Internal-Token header can
// be checked before any cache invalidation is allowed.
type Module struct{}

func NewModule() *Module { return &Module{} }
