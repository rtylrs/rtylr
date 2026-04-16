package domain

import (
	"context"
	"time"
)

type DashboardWidget struct {
	ID         int64
	OwnerID    int64
	RoleID     int64
	WidgetType string
	PositionX  int32
	PositionY  int32
	Width      int32
	Height     int32
	Config     []byte
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type Repository interface {
	GetWidgets(ctx context.Context, ownerID int64) ([]*DashboardWidget, error)
	UpsertWidgets(ctx context.Context, ownerID int64, widgets []*DashboardWidget) error
}
