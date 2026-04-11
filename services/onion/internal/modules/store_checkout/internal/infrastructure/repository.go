package infrastructure

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/voxire/onion/internal/modules/store_checkout/internal/domain"
	"github.com/voxire/onion/internal/modules/store_checkout/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

// ensure domain is used
var _ = domain.Checkout{}

type Repository struct {
	queries *sqlc.Queries
	pool    *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	pool := dbclient.Pool()
	return &Repository{queries: sqlc.New(pool), pool: pool}
}

func (r *Repository) Create(ctx context.Context, ownerID int64, input domain.CheckoutCreateInput) (int64, error) {
	result, err := r.queries.CreateCheckout(ctx, sqlc.CreateCheckoutParams{
		OwnerID:       ownerID,
		SessionID:     sql.NullString{String: input.SessionID, Valid: input.SessionID != ""},
		CustomerEmail: sql.NullString{String: input.CustomerEmail, Valid: input.CustomerEmail != ""},
		Items:         nullJSON(input.Items),
		Subtotal:      strconv.FormatFloat(input.Subtotal, 'f', 2, 64),
		Tax:           strconv.FormatFloat(input.Tax, 'f', 2, 64),
		Total:         strconv.FormatFloat(input.Total, 'f', 2, 64),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create checkout", logger.Error(err))
		return 0, err
	}
	id, _ := result.LastInsertId()
	return id, nil
}

func (r *Repository) Complete(ctx context.Context, ownerID, id int64) error {
	err := r.queries.CompleteCheckout(ctx, sqlc.CompleteCheckoutParams{OwnerID: ownerID, ID: id})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to complete checkout", logger.Error(err))
		return err
	}
	return nil
}

func nullJSON(s string) []byte {
	if s == "" { return nil }
	return []byte(s)
}
