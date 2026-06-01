package application

import (
	"context"
	"database/sql"
	"strconv"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/pos_kds/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_kds/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
	"github.com/voxire/onion/internal/platform/wsbus"
)

// KDS active orders are kitchen tickets — extremely time-sensitive. The TTL
// is intentionally short and any item-status mutation also invalidates the
// scope explicitly.
const kdsCacheTTL = 10 * time.Second

const kdsScope = "pos_kds"

type repository interface {
	ListActiveOrders(ctx context.Context, ownerID int64) ([]domain.KDSOrder, error)
	UpdateItemStatus(ctx context.Context, input domain.UpdateItemStatusInput) error
	UpdateItemStatusByItemID(ctx context.Context, ownerID, itemID int64, status string) error
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{repo: infrastructure.NewRepository(db)}
}

func invalidateKDSScope(ctx context.Context, ownerID int64) {
	if cache := cacher.GetRedisCache(); cache != nil {
		cache.DeleteByPattern(ctx, cacheKeys.ScopedPrefix(kdsScope, ownerID))
	}
}

func (s *Service) ListActiveOrders(ctx context.Context, ownerID int64) ([]domain.KDSOrder, error) {
	return cacher.Of[[]domain.KDSOrder]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(kdsScope, ownerID, "active"),
		func(ctx context.Context) ([]domain.KDSOrder, error) {
			return s.repo.ListActiveOrders(ctx, ownerID)
		},
		cacher.WithExpiry(ctx, kdsCacheTTL),
	)
}

// ListActiveOrdersByStation returns active KDS orders routed to the given station.
// When station is empty, returns all active orders. Matching is a case-insensitive
// substring match against the order kitchen_notes field until a dedicated station
// column is added to the schema.
func (s *Service) ListActiveOrdersByStation(ctx context.Context, ownerID int64, station string) ([]domain.KDSOrder, error) {
	orders, err := s.ListActiveOrders(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	if station == "" {
		return orders, nil
	}
	needle := strings.ToLower(strings.TrimSpace(station))
	filtered := make([]domain.KDSOrder, 0, len(orders))
	for _, o := range orders {
		if strings.Contains(strings.ToLower(o.KitchenNotes), needle) {
			filtered = append(filtered, o)
		}
	}
	return filtered, nil
}

func (s *Service) UpdateItemStatus(ctx context.Context, input domain.UpdateItemStatusInput) error {
	if err := s.repo.UpdateItemStatus(ctx, input); err != nil {
		return err
	}
	invalidateKDSScope(ctx, input.OwnerID)
	wsbus.Publish(wsbus.Event{
		Topic:    "owner:" + strconv.FormatInt(input.OwnerID, 10),
		Type:     "pos:kds:item:updated",
		EntityID: input.ItemID,
	})
	return nil
}

func (s *Service) UpdateItemStatusByItemID(ctx context.Context, ownerID, itemID int64, status string) error {
	if err := s.repo.UpdateItemStatusByItemID(ctx, ownerID, itemID, status); err != nil {
		return err
	}
	invalidateKDSScope(ctx, ownerID)
	wsbus.Publish(wsbus.Event{
		Topic:    "owner:" + strconv.FormatInt(ownerID, 10),
		Type:     "pos:kds:item:updated",
		EntityID: itemID,
	})
	return nil
}
