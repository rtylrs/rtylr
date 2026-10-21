package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/pos/memberlocations/internal/domain"
)

type repository interface {
	ListForUser(context.Context, int64, int64) ([]domain.MemberLocation, error)
	GetMyLocations(context.Context, int64, int64) ([]domain.MyLocation, error)
	GetBusinessLocations(context.Context, int64) ([]domain.MyLocation, error)
	GetLocationRoster(context.Context, int64, int64) ([]domain.RosterEntry, error)
	IsAssignedToLocation(context.Context, int64, int64, int64) (bool, error)
	UserBelongsToBusiness(context.Context, int64, int64) (bool, error)
	LocationBelongsToBusiness(context.Context, int64, int64) (bool, error)
	Assign(context.Context, int64, int64, int64, bool, bool, bool) (int64, error)
	Remove(context.Context, int64, int64) error
}
