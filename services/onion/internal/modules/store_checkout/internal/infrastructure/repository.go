package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"strconv"

	"github.com/voxire/onion/internal/modules/store_checkout/internal/domain"
	"github.com/voxire/onion/internal/modules/store_checkout/internal/infrastructure/persistence/sqlc"
	customersync "github.com/voxire/onion/internal/platform/customer_sync"
	"github.com/voxire/onion/internal/platform/logger"
)

// ensure domain is used.
var _ = domain.Checkout{}

type Repository struct {
	queries *sqlc.Queries
	pool    *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	pool := db
	return &Repository{queries: sqlc.New(pool), pool: pool}
}

func (r *Repository) Create(ctx context.Context, businessID int64, ownerID int64, input domain.CheckoutCreateInput) (int64, error) {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to begin checkout transaction", logger.Error(err))
		return 0, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err = customersync.UpsertCRMCustomer(ctx, tx, businessID, customersync.Contact{
		Email: input.CustomerEmail,
	}); err != nil {
		logger.Ctx(ctx).Error("Failed to sync checkout customer", logger.Error(err))
		return 0, err
	}

	result, err := r.queries.WithTx(tx).CreateCheckout(ctx, sqlc.CreateCheckoutParams{
		OwnerID:       ownerID,
		BusinessID:    businessID,
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
	id, err := result.LastInsertId()
	if err != nil {
		logger.Ctx(ctx).Error("Failed to read checkout insert id", logger.Error(err))
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		logger.Ctx(ctx).Error("Failed to commit checkout transaction", logger.Error(err))
		return 0, err
	}
	return id, nil
}

func (r *Repository) Complete(ctx context.Context, businessID, id int64, ownerID int64) error {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to begin checkout completion transaction", logger.Error(err))
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var customerEmail sql.NullString
	var totalString string
	checkout, err := r.queries.WithTx(tx).GetCheckoutForComplete(ctx, sqlc.GetCheckoutForCompleteParams{
		BusinessID: businessID,
		ID:         id,
	})
	if errors.Is(err, sql.ErrNoRows) {
		err = nil
	} else if err != nil {
		logger.Ctx(ctx).Error("Failed to load checkout customer", logger.Error(err))
		return err
	} else {
		customerEmail = checkout.CustomerEmail
		totalString = checkout.Total
		total, parseErr := strconv.ParseFloat(totalString, 64)
		if parseErr != nil {
			logger.Ctx(ctx).Warn("failed to parse checkout total for CRM sync",
				logger.String("total", totalString),
				logger.Error(parseErr),
			)
		} else if _, err = customersync.UpsertCRMCustomer(ctx, tx, businessID, customersync.Contact{
			Email:      customerEmail.String,
			TotalSpent: total,
			CountVisit: true,
		}); err != nil {
			logger.Ctx(ctx).Error("Failed to sync completed checkout customer", logger.Error(err))
			return err
		}
	}

	err = r.queries.WithTx(tx).CompleteCheckout(ctx, sqlc.CompleteCheckoutParams{BusinessID: businessID, ID: id})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to complete checkout", logger.Error(err))
		return err
	}
	if err = tx.Commit(); err != nil {
		logger.Ctx(ctx).Error("Failed to commit checkout completion transaction", logger.Error(err))
		return err
	}
	return nil
}

func nullJSON(s string) []byte {
	if s == "" {
		return nil
	}
	return []byte(s)
}
