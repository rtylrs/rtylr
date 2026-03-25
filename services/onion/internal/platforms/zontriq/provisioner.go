package zontriq

import (
	"context"
	"fmt"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platforms/zontriq/users/adapter"
)

const defaultStarterPlanID = 1

type Provisioner struct {
	usersAdapter *adapter.UserAdapter
}

func NewProvisioner(usersAdapter *adapter.UserAdapter) *Provisioner {
	return &Provisioner{usersAdapter: usersAdapter}
}

func (p *Provisioner) ProvisionUser(ctx context.Context, userID int64) (*authPkg.PlatformClaims, error) {
	existing, err := p.usersAdapter.GetZontriqUserByUserID(ctx, int(userID))
	if err == nil {
		return &authPkg.PlatformClaims{
			Platform:       "zontriq",
			PlatformUserID: int64(existing.ID),
			PlanID:         int64(existing.PlanID),
		}, nil
	}

	created, err := p.usersAdapter.CreateZontriqUser(ctx, int(userID), defaultStarterPlanID)
	if err != nil {
		return nil, fmt.Errorf("failed to provision zontriq user: %w", err)
	}

	return &authPkg.PlatformClaims{
		Platform:       "zontriq",
		PlatformUserID: int64(created.ID),
		PlanID:         int64(created.PlanID),
	}, nil
}

func (p *Provisioner) ResolvePlatformClaims(ctx context.Context, userID int64) (*authPkg.PlatformClaims, error) {
	existing, err := p.usersAdapter.GetZontriqUserByUserID(ctx, int(userID))
	if err != nil {
		return nil, fmt.Errorf("zontriq user not found for user_id %d: %w", userID, err)
	}

	return &authPkg.PlatformClaims{
		Platform:       "zontriq",
		PlatformUserID: int64(existing.ID),
		PlanID:         int64(existing.PlanID),
	}, nil
}
