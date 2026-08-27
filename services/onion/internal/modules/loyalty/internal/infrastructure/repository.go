package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"strconv"

	"github.com/voxire/onion/internal/modules/loyalty/internal/domain"
	"github.com/voxire/onion/internal/modules/loyalty/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	pool    *sql.DB
	queries *sqlc.Queries
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		pool:    db,
		queries: sqlc.New(db),
	}
}

func (r *Repository) ListTransactions(ctx context.Context, businessID, customerID int64, page, pageSize int) ([]domain.LoyaltyTransaction, int, error) {
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
			BusinessID: businessID,
			CustomerID: customerID,
		})
	} else {
		total, err = r.queries.CountTransactions(ctx, businessID)
	}

	if err != nil {
		return nil, 0, err
	}

	if customerID > 0 {
		rows, qErr := r.queries.ListTransactionsByCustomer(ctx, sqlc.ListTransactionsByCustomerParams{
			BusinessID: businessID,
			CustomerID: customerID,
			Limit:      int32(pageSize),
			Offset:     int32(offset), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
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
				BusinessID:    row.BusinessID,
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
		BusinessID: businessID,
		Limit:      int32(pageSize),
		Offset:     int32(offset), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
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
		BusinessID:    input.BusinessID,
		CustomerID:    input.CustomerID,
		Points:        int32(input.Points), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
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
		BusinessID:    input.BusinessID,
		CustomerID:    input.CustomerID,
		Points:        int32(input.Points), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		Description:   sql.NullString{String: input.Description, Valid: input.Description != ""},
		ReferenceID:   nullInt64FromPtr(input.ReferenceID),
		ReferenceType: sql.NullString{String: input.ReferenceType, Valid: input.ReferenceType != ""},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create redeem transaction", logger.Error(err))
	}

	return err
}

func (r *Repository) GetCustomerPointsBalance(ctx context.Context, businessID, customerID int64) (int64, error) {
	return r.queries.GetCustomerPointsBalance(ctx, sqlc.GetCustomerPointsBalanceParams{
		BusinessID: businessID,
		CustomerID: customerID,
	})
}

// parseLoyaltyDecimal parses a DECIMAL column stored as a string. A corrupt or
// empty value logs at Warn and yields 0 rather than being silently discarded.
func parseLoyaltyDecimal(ctx context.Context, s string) float64 {
	if s == "" {
		return 0
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		logger.Ctx(ctx).Warn("failed to parse loyalty config decimal",
			logger.String("value", s), logger.Error(err))
		return 0
	}
	return f
}

func (r *Repository) GetConfig(ctx context.Context, businessID int64) (domain.LoyaltyConfig, error) {
	row, err := r.queries.GetConfig(ctx, businessID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.LoyaltyConfig{}, domain.ErrLoyaltyConfigNotFound
		}

		logger.Ctx(ctx).Error("Failed to get loyalty config", logger.Error(err))

		return domain.LoyaltyConfig{}, err
	}

	ppd := parseLoyaltyDecimal(ctx, row.PointsPerCurrencyUnit.String)
	pv := parseLoyaltyDecimal(ctx, row.CurrencyPerPoint.String)

	return domain.LoyaltyConfig{
		ID:                  row.ID,
		OwnerID:             row.OwnerID,
		BusinessID:          row.BusinessID,
		PointsPerDollar:     ppd,
		MinRedemptionPoints: int(row.MinRedemptionPoints.Int32),
		PointsValue:         pv,
		IsActive:            row.IsEnabled.Bool,
		UpdatedAt:           row.UpdatedAt.Time,
	}, nil
}

func (r *Repository) UpsertConfig(ctx context.Context, input domain.LoyaltyConfigInput) error {
	err := r.queries.UpsertConfig(ctx, sqlc.UpsertConfigParams{
		OwnerID:               input.OwnerID,
		BusinessID:            input.BusinessID,
		PointsPerCurrencyUnit: sql.NullString{String: strconv.FormatFloat(input.PointsPerDollar, 'f', 2, 64), Valid: true},
		CurrencyPerPoint:      sql.NullString{String: strconv.FormatFloat(input.PointsValue, 'f', 4, 64), Valid: true},
		MinRedemptionPoints:   sql.NullInt32{Int32: int32(input.MinRedemptionPoints), Valid: true}, // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		IsEnabled:             sql.NullBool{Bool: input.IsActive, Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to upsert loyalty config", logger.Error(err))
	}

	return err
}

// ListMembers returns customers with loyalty engagement (points or visits),
// ordered by points desc. Projection over crm_customers — tracked here rather
// than in the customers module so the loyalty domain stays self-contained.
func (r *Repository) ListMembers(ctx context.Context, businessID int64, limit, offset int) ([]domain.LoyaltyMember, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	// cross-schema read from crm_customers — loyalty members projection
	rows, err := r.queries.ListLoyaltyMembers(ctx, sqlc.ListLoyaltyMembersParams{
		BusinessID: businessID,
		Limit:      int32(limit),
		Offset:     int32(offset),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list loyalty members", logger.Error(err))
		return nil, err
	}

	members := make([]domain.LoyaltyMember, 0, len(rows))
	for _, row := range rows {
		m := domain.LoyaltyMember{
			ID:      row.ID,
			OwnerID: row.OwnerID,
			Name:    row.Name,
			Points:  int(row.LoyaltyPoints),
			Tier:    string(row.LoyaltyTier),
			Visits:  int(row.VisitCount),
		}
		if row.CreatedAt.Valid {
			m.JoinedAt = row.CreatedAt.Time
		}
		if row.LastActiveAt.Valid {
			m.LastActiveAt = row.LastActiveAt.Time
		}
		members = append(members, m)
	}
	return members, nil
}

// ListRewards returns the catalog of rewards for an owner.
func (r *Repository) ListRewards(ctx context.Context, businessID int64, activeOnly bool) ([]domain.LoyaltyReward, error) {
	var rows []sqlc.ListLoyaltyRewardsRow
	var err error
	if activeOnly {
		activeRows, activeErr := r.queries.ListActiveLoyaltyRewards(ctx, businessID)
		rows = make([]sqlc.ListLoyaltyRewardsRow, len(activeRows))
		for i, row := range activeRows {
			rows[i] = sqlc.ListLoyaltyRewardsRow(row)
		}
		err = activeErr
	} else {
		rows, err = r.queries.ListLoyaltyRewards(ctx, businessID)
	}
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list loyalty rewards", logger.Error(err))
		return nil, err
	}

	rewards := make([]domain.LoyaltyReward, 0, len(rows))
	for _, row := range rows {
		rw := domain.LoyaltyReward{
			ID:          row.ID,
			OwnerID:     row.OwnerID,
			Name:        row.Name,
			Description: row.Description,
			Category:    row.Category,
			PointsCost:  int(row.PointsCost),
			Icon:        row.Icon,
			Active:      row.Active,
		}
		if row.Stock.Valid {
			v := int(row.Stock.Int32)
			rw.Stock = &v
		}
		if row.CreatedAt.Valid {
			rw.CreatedAt = row.CreatedAt.Time
		}
		if row.UpdatedAt.Valid {
			rw.UpdatedAt = row.UpdatedAt.Time
		}
		rewards = append(rewards, rw)
	}
	return rewards, nil
}

// UpsertReward creates or updates a reward. ID=0 → insert; ID>0 → update.
func (r *Repository) UpsertReward(ctx context.Context, rw domain.LoyaltyReward) (int64, error) {
	businessID := rw.BusinessID
	if businessID == 0 {
		businessID = rw.OwnerID
	}

	var stock sql.NullInt32
	if rw.Stock != nil {
		stock = sql.NullInt32{Int32: int32(*rw.Stock), Valid: true} // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	}
	if rw.ID == 0 {
		res, err := r.queries.CreateReward(ctx, sqlc.CreateRewardParams{
			OwnerID:     rw.OwnerID,
			BusinessID:  businessID,
			Name:        rw.Name,
			Description: sql.NullString{String: rw.Description, Valid: true},
			Category:    sql.NullString{String: rw.Category, Valid: true},
			PointsCost:  int32(rw.PointsCost), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
			Stock:       stock,
			Icon:        sql.NullString{String: rw.Icon, Valid: true},
			Active:      sql.NullBool{Bool: rw.Active, Valid: true},
		})
		if err != nil {
			logger.Ctx(ctx).Error("Failed to insert loyalty reward", logger.Error(err))
			return 0, err
		}
		id, idErr := res.LastInsertId()
		if idErr != nil {
			return 0, idErr
		}
		return id, nil
	}
	err := r.queries.UpdateReward(ctx, sqlc.UpdateRewardParams{
		Name:        rw.Name,
		Description: sql.NullString{String: rw.Description, Valid: true},
		Category:    sql.NullString{String: rw.Category, Valid: true},
		PointsCost:  int32(rw.PointsCost), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		Stock:       stock,
		Icon:        sql.NullString{String: rw.Icon, Valid: true},
		Active:      sql.NullBool{Bool: rw.Active, Valid: true},
		ID:          rw.ID,
		BusinessID:  businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update loyalty reward", logger.Error(err))
		return 0, err
	}
	return rw.ID, nil
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
