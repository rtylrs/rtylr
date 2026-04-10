package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"strconv"

	"github.com/voxire/onion/internal/modules/loyalty/internal/domain"
	"github.com/voxire/onion/internal/modules/loyalty/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	queries *sqlc.Queries
}

func NewRepository(_ context.Context) *Repository {
	return &Repository{
		queries: sqlc.New(dbclient.Pool()),
	}
}

func (r *Repository) ListTransactions(ctx context.Context, ownerID int64, customerID int64, page, pageSize int) ([]domain.LoyaltyTransaction, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var total int64
	var err error

	if customerID > 0 {
		total, err = r.queries.CountTransactionsByCustomer(ctx, sqlc.CountTransactionsByCustomerParams{
			OwnerID:    ownerID,
			CustomerID: customerID,
		})
	} else {
		total, err = r.queries.CountTransactions(ctx, ownerID)
	}

	if err != nil {
		return nil, 0, err
	}

	if customerID > 0 {
		rows, qErr := r.queries.ListTransactionsByCustomer(ctx, sqlc.ListTransactionsByCustomerParams{
			OwnerID:    ownerID,
			CustomerID: customerID,
			Limit:      int32(pageSize),
			Offset:     int32(offset),
		})
		if qErr != nil {
			logger.Ctx(ctx).Error("Failed to list loyalty transactions", logger.Error(qErr))

			return nil, 0, qErr
		}

		txs := make([]domain.LoyaltyTransaction, 0, len(rows))
		for _, row := range rows {
			txs = append(txs, domain.LoyaltyTransaction{
				ID:            row.ID,
				OwnerID:       row.OwnerID,
				CustomerID:    row.CustomerID,
				Points:        int(row.Points),
				Type:          string(row.Type),
				Description:   row.Description.String,
				ReferenceID:   nullInt64Ptr(row.ReferenceID),
				ReferenceType: row.ReferenceType.String,
				CreatedAt:     row.CreatedAt.Time,
			})
		}

		return txs, int(total), nil
	}

	rows, err := r.queries.ListTransactions(ctx, sqlc.ListTransactionsParams{
		OwnerID: ownerID,
		Limit:   int32(pageSize),
		Offset:  int32(offset),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list loyalty transactions", logger.Error(err))

		return nil, 0, err
	}

	txs := make([]domain.LoyaltyTransaction, 0, len(rows))
	for _, row := range rows {
		txs = append(txs, domain.LoyaltyTransaction{
			ID:            row.ID,
			OwnerID:       row.OwnerID,
			CustomerID:    row.CustomerID,
			Points:        int(row.Points),
			Type:          string(row.Type),
			Description:   row.Description.String,
			ReferenceID:   nullInt64Ptr(row.ReferenceID),
			ReferenceType: row.ReferenceType.String,
			CreatedAt:     row.CreatedAt.Time,
		})
	}

	return txs, int(total), nil
}

func (r *Repository) CreateTransaction(ctx context.Context, input domain.AwardInput, txType string) error {
	err := r.queries.CreateTransaction(ctx, sqlc.CreateTransactionParams{
		OwnerID:       input.OwnerID,
		CustomerID:    input.CustomerID,
		Points:        int32(input.Points),
		Type:          sqlc.CrmLoyaltyTransactionsType(txType),
		Description:   sql.NullString{String: input.Description, Valid: input.Description != ""},
		ReferenceID:   nullInt64FromPtr(input.ReferenceID),
		ReferenceType: sql.NullString{String: input.ReferenceType, Valid: input.ReferenceType != ""},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create loyalty transaction", logger.Error(err))
	}

	return err
}

func (r *Repository) CreateRedeemTransaction(ctx context.Context, input domain.RedeemInput) error {
	err := r.queries.CreateRedeemTransaction(ctx, sqlc.CreateRedeemTransactionParams{
		OwnerID:       input.OwnerID,
		CustomerID:    input.CustomerID,
		Points:        int32(input.Points),
		Description:   sql.NullString{String: input.Description, Valid: input.Description != ""},
		ReferenceID:   nullInt64FromPtr(input.ReferenceID),
		ReferenceType: sql.NullString{String: input.ReferenceType, Valid: input.ReferenceType != ""},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create redeem transaction", logger.Error(err))
	}

	return err
}

func (r *Repository) GetConfig(ctx context.Context, ownerID int64) (domain.LoyaltyConfig, error) {
	row, err := r.queries.GetConfig(ctx, ownerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.LoyaltyConfig{}, domain.ErrLoyaltyConfigNotFound
		}

		logger.Ctx(ctx).Error("Failed to get loyalty config", logger.Error(err))

		return domain.LoyaltyConfig{}, err
	}

	ppd, _ := strconv.ParseFloat(row.PointsPerCurrencyUnit.String, 64)
	pv, _ := strconv.ParseFloat(row.CurrencyPerPoint.String, 64)

	return domain.LoyaltyConfig{
		ID:              row.ID,
		OwnerID:         row.OwnerID,
		PointsPerDollar: ppd,
		PointsValue:     pv,
		IsActive:        row.IsEnabled.Bool,
		UpdatedAt:       row.UpdatedAt.Time,
	}, nil
}

func (r *Repository) UpsertConfig(ctx context.Context, input domain.LoyaltyConfigInput) error {
	err := r.queries.UpsertConfig(ctx, sqlc.UpsertConfigParams{
		OwnerID:               input.OwnerID,
		PointsPerCurrencyUnit: sql.NullString{String: strconv.FormatFloat(input.PointsPerDollar, 'f', 2, 64), Valid: true},
		CurrencyPerPoint:      sql.NullString{String: strconv.FormatFloat(input.PointsValue, 'f', 4, 64), Valid: true},
		IsEnabled:             sql.NullBool{Bool: input.IsActive, Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to upsert loyalty config", logger.Error(err))
	}

	return err
}

func nullInt64Ptr(v sql.NullInt64) *int64 {
	if v.Valid {
		return &v.Int64
	}

	return nil
}

func nullInt64FromPtr(p *int64) sql.NullInt64 {
	if p != nil {
		return sql.NullInt64{Int64: *p, Valid: true}
	}

	return sql.NullInt64{}
}
