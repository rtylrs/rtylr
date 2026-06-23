package runmode

import (
	"context"

	"github.com/go-chi/chi/v5"
)

type DatabaseTag string

const (
	DatabaseMain    DatabaseTag = "main"
	DatabaseLogging DatabaseTag = "logging"
	DatabaseReplica DatabaseTag = "replica"
)

type GRPCConfig struct {
	Enabled bool
	Port    string
}

// Definition captures the boot contract for an Onion runmode.
type Definition interface {
	Name() string
	DatabaseTags() []DatabaseTag
	HasWebServer() bool
	GRPC() GRPCConfig
	Setup(context.Context) error
	Start(context.Context) error
	RegisterRoutes(context.Context, *chi.Mux)
}

// Spec is a lightweight implementation of Definition for simple runmodes.
type Spec struct {
	RunModeName string
	DBTags      []DatabaseTag
	WebServer   bool
	GRPCConfig  GRPCConfig
	SetupFunc   func(context.Context) error
	StartFunc   func(context.Context) error
	RoutesFunc  func(context.Context, *chi.Mux)
}

func (s Spec) Name() string { return s.RunModeName }

func (s Spec) DatabaseTags() []DatabaseTag {
	if len(s.DBTags) == 0 {
		return nil
	}
	out := make([]DatabaseTag, len(s.DBTags))
	copy(out, s.DBTags)
	return out
}

func (s Spec) HasWebServer() bool { return s.WebServer }

func (s Spec) GRPC() GRPCConfig { return s.GRPCConfig }

func (s Spec) Setup(ctx context.Context) error {
	if s.SetupFunc == nil {
		return nil
	}
	return s.SetupFunc(ctx)
}

func (s Spec) Start(ctx context.Context) error {
	if s.StartFunc == nil {
		return nil
	}
	return s.StartFunc(ctx)
}

func (s Spec) RegisterRoutes(ctx context.Context, r *chi.Mux) {
	if s.RoutesFunc != nil {
		s.RoutesFunc(ctx, r)
	}
}

// HTTPAdapter lets a Definition satisfy the existing HTTP server interface.
type HTTPAdapter struct {
	Definition Definition
}

func (a HTTPAdapter) Setup(ctx context.Context, r *chi.Mux) {
	if a.Definition != nil {
		a.Definition.RegisterRoutes(ctx, r)
	}
}
