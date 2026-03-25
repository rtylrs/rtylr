package voxire

import (
	"context"

	"github.com/voxire/onion/internal/platforms/voxire/careers"
)

type Platform struct {
	Careers *careers.Module
}

func NewPlatform(ctx context.Context) *Platform {
	return &Platform{
		Careers: careers.NewModule(ctx),
	}
}
