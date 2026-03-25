package adapter

import (
	"context"

	"github.com/voxire/onion/internal/platforms/zontriq/users/internal/application"
)

type ZontriqUserInfo struct {
	ID     int
	PlanID int
}

type UserAdapter struct {
	service *application.Service
}

func NewUserAdapter(service *application.Service) *UserAdapter {
	return &UserAdapter{service: service}
}

func (a *UserAdapter) GetZontriqUserByUserID(ctx context.Context, userID int) (*ZontriqUserInfo, error) {
	u, err := a.service.GetZontriqUserByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &ZontriqUserInfo{ID: u.ID, PlanID: u.PlanID}, nil
}

func (a *UserAdapter) CreateZontriqUser(ctx context.Context, userID, planID int) (*ZontriqUserInfo, error) {
	u, err := a.service.CreateZontriqUser(ctx, userID, planID)
	if err != nil {
		return nil, err
	}

	return &ZontriqUserInfo{ID: u.ID, PlanID: u.PlanID}, nil
}
