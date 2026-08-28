package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"math"
	"strconv"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"golang.org/x/sync/errgroup"

	"github.com/voxire/onion/internal/modules/campaigns"
	"github.com/voxire/onion/internal/modules/customers/internal/domain"
	"github.com/voxire/onion/internal/modules/customers/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/modules/segments"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/searchutil"
	"github.com/voxire/onion/internal/platform/segmentrules"
)

type Repository struct {
	queries   *sqlc.Queries
	pool      *sql.DB
	campaigns *campaigns.Module
	segments  *segments.Module
}

func NewRepository(db *sql.DB) *Repository {
	pool := db

	return &Repository{
		queries:   sqlc.New(pool),
		pool:      pool,
		campaigns: campaigns.NewModule(context.Background(), pool),
		segments:  segments.NewModule(context.Background(), pool),
	}
}

// ─── Customers ───────────────────────────────────────────────────────────────

func (r *Repository) ListCustomers(ctx context.Context, filter domain.CustomerListFilter) ([]domain.Customer, int, error) {
	businessID := filter.BusinessID
	if businessID == 0 {
		businessID = filter.OwnerID
	}
	filter.BusinessID = businessID

	if filter.Page < 1 {
		filter.Page = 1
	}

	if filter.PageSize < 1 || filter.PageSize > 100 {
		filter.PageSize = 20
	}

	offset := (filter.Page - 1) * filter.PageSize

	// Dynamic filtering requires raw SQL for search/tier/tag
	if filter.Search != "" || filter.Tier != "" || filter.Tag != "" {
		return r.listCustomersFiltered(ctx, filter, offset)
	}

	total, err := r.queries.CountCustomers(ctx, filter.BusinessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to count customers", logger.Error(err))

		return nil, 0, err
	}

	rows, err := r.queries.ListCustomers(ctx, sqlc.ListCustomersParams{
		BusinessID: filter.BusinessID,
		Limit:      int32(filter.PageSize), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		Offset:     int32(offset),          // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list customers", logger.Error(err))

		return nil, 0, err
	}

	customers := make([]domain.Customer, 0, len(rows))

	for _, row := range rows {
		customers = append(customers, customerFromRow(row))
	}

	return customers, int(total), nil
}

// dynamic-query: squirrel
func (r *Repository) listCustomersFiltered(ctx context.Context, filter domain.CustomerListFilter, offset int) ([]domain.Customer, int, error) {
	conds := sq.And{sq.Eq{"business_id": filter.BusinessID}}

	if search := strings.TrimSpace(filter.Search); search != "" {
		booleanQuery := searchutil.BooleanPrefixQuery(search)
		if booleanQuery != "" {
			conds = append(conds, sq.Expr(
				"(email = ? OR phone = ? OR MATCH(first_name, last_name, email, phone) AGAINST (? IN BOOLEAN MODE))",
				search, search, booleanQuery,
			))
		} else {
			conds = append(conds, sq.Or{sq.Eq{"email": search}, sq.Eq{"phone": search}})
		}
	}

	if filter.Tier != "" {
		conds = append(conds, sq.Eq{"loyalty_tier": filter.Tier})
	}

	if filter.Tag != "" {
		tagJSON, err := json.Marshal(filter.Tag)
		if err != nil {
			return nil, 0, err
		}
		conds = append(conds, sq.Expr("JSON_CONTAINS(tags, ?)", string(tagJSON)))
	}

	countSQL, countArgs, err := sq.Select("COUNT(*)").
		From("crm_customers").
		Where(conds).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("Failed to build customer count query", logger.Error(err))

		return nil, 0, err
	}

	var total int

	if err := r.pool.QueryRowContext(ctx, countSQL, countArgs...).Scan(&total); err != nil {
		logger.Ctx(ctx).Error("Failed to count customers", logger.Error(err))

		return nil, 0, err
	}

	listSQL, listArgs, err := sq.Select(
		"id", "owner_id", "business_id", "first_name", "last_name", "email", "phone", "date_of_birth", "gender",
		"address", "city", "country", "loyalty_points", "total_spent", "visit_count", "last_visit_at",
		"loyalty_tier", "tags", "notes", "consent", "is_active", "created_at", "updated_at",
	).
		From("crm_customers").
		Where(conds).
		OrderBy("created_at DESC").
		Limit(uint64(filter.PageSize)). // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		Offset(uint64(offset)).         // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("Failed to build customer list query", logger.Error(err))

		return nil, 0, err
	}

	rows, err := r.pool.QueryContext(ctx, listSQL, listArgs...)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list customers", logger.Error(err))

		return nil, 0, err
	}
	defer rows.Close()

	customers := make([]domain.Customer, 0, filter.PageSize)

	for rows.Next() {
		c, scanErr := scanCustomerRaw(rows)
		if scanErr != nil {
			logger.Ctx(ctx).Error("Failed to scan customer", logger.Error(scanErr))

			return nil, 0, scanErr
		}

		customers = append(customers, c)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return customers, total, nil
}

func (r *Repository) GetCustomerByID(ctx context.Context, businessID, customerID int64) (domain.Customer, error) {
	row, err := r.queries.GetCustomerByID(ctx, sqlc.GetCustomerByIDParams{
		ID:         customerID,
		BusinessID: businessID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Customer{}, domain.ErrCustomerNotFound
		}

		logger.Ctx(ctx).Error("Failed to get customer", logger.Error(err))

		return domain.Customer{}, err
	}

	return customerFromRow(row), nil
}

func (r *Repository) CreateCustomer(ctx context.Context, input domain.CustomerInput) (int64, error) {
	tagsJSON, err := json.Marshal(input.Tags)
	if err != nil {
		tagsJSON = []byte("[]")
	}

	result, err := r.queries.CreateCustomer(ctx, sqlc.CreateCustomerParams{
		OwnerID:     input.OwnerID,
		BusinessID:  input.BusinessID,
		FirstName:   input.FirstName,
		LastName:    sql.NullString{String: input.LastName, Valid: input.LastName != ""},
		Email:       sql.NullString{String: input.Email, Valid: input.Email != ""},
		Phone:       sql.NullString{String: input.Phone, Valid: input.Phone != ""},
		DateOfBirth: nullableTime(input.DateOfBirth),
		Gender:      sqlc.NullCrmCustomersGender{CrmCustomersGender: sqlc.CrmCustomersGender(input.Gender), Valid: input.Gender != ""},
		Address:     sql.NullString{String: input.Address, Valid: input.Address != ""},
		City:        sql.NullString{String: input.City, Valid: input.City != ""},
		Country:     sql.NullString{String: input.Country, Valid: input.Country != ""},
		Tags:        tagsJSON,
		Notes:       sql.NullString{String: input.Notes, Valid: input.Notes != ""},
		Consent:     input.Consent,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create customer", logger.Error(err))

		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *Repository) UpdateCustomer(ctx context.Context, input domain.CustomerUpdateInput) error {
	businessID := input.BusinessID
	if businessID == 0 {
		businessID = input.OwnerID
	}
	input.BusinessID = businessID
	return updateCustomer(ctx, r.queries, input)
}

// UpdateCustomerWithLoyalty keeps the customer update and tier change in one
// transaction. The points row is locked before it is preserved, preventing a
// concurrent earn/redeem from being overwritten by a stale read.
func (r *Repository) UpdateCustomerWithLoyalty(ctx context.Context, input domain.CustomerUpdateInput) error {
	businessID := input.BusinessID
	if businessID == 0 {
		businessID = input.OwnerID
	}
	input.BusinessID = businessID

	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	qtx := r.queries.WithTx(tx)
	if err := updateCustomer(ctx, qtx, input); err != nil {
		return err
	}

	var loyaltyPoints sql.NullInt64
	if err := tx.QueryRowContext(ctx,
		"SELECT loyalty_points FROM crm_customers WHERE id = ? AND business_id = ? FOR UPDATE",
		input.ID, businessID,
	).Scan(&loyaltyPoints); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrCustomerNotFound
		}
		return err
	}
	points := int64(0)
	if loyaltyPoints.Valid {
		points = loyaltyPoints.Int64
	}
	if err := qtx.UpdateCustomerLoyalty(ctx, sqlc.UpdateCustomerLoyaltyParams{
		LoyaltyPoints: sql.NullInt32{Int32: int32(points), Valid: true}, // #nosec G115 -- Reviewed width conversion uses validated database field width.
		LoyaltyTier: sqlc.NullCrmCustomersLoyaltyTier{
			CrmCustomersLoyaltyTier: sqlc.CrmCustomersLoyaltyTier(input.LoyaltyTier),
			Valid:                   true,
		},
		ID:         input.ID,
		BusinessID: businessID,
	}); err != nil {
		return err
	}
	return tx.Commit()
}

func updateCustomer(ctx context.Context, q *sqlc.Queries, input domain.CustomerUpdateInput) error {
	tagsJSON, err := json.Marshal(input.Tags)
	if err != nil {
		tagsJSON = []byte("[]")
	}

	isActive := true
	if input.IsActive != nil {
		isActive = *input.IsActive
	}

	result, err := q.UpdateCustomer(ctx, sqlc.UpdateCustomerParams{
		FirstName:   input.FirstName,
		LastName:    sql.NullString{String: input.LastName, Valid: input.LastName != ""},
		Email:       sql.NullString{String: input.Email, Valid: input.Email != ""},
		Phone:       sql.NullString{String: input.Phone, Valid: input.Phone != ""},
		DateOfBirth: nullableTime(input.DateOfBirth),
		Gender:      sqlc.NullCrmCustomersGender{CrmCustomersGender: sqlc.CrmCustomersGender(input.Gender), Valid: input.Gender != ""},
		Address:     sql.NullString{String: input.Address, Valid: input.Address != ""},
		City:        sql.NullString{String: input.City, Valid: input.City != ""},
		Country:     sql.NullString{String: input.Country, Valid: input.Country != ""},
		Tags:        tagsJSON,
		Notes:       sql.NullString{String: input.Notes, Valid: input.Notes != ""},
		Consent:     input.Consent,
		IsActive:    sql.NullBool{Bool: isActive, Valid: true},
		ID:          input.ID,
		BusinessID:  input.BusinessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update customer", logger.Error(err))

		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return domain.ErrCustomerNotFound
	}

	return nil
}

func (r *Repository) DeleteCustomer(ctx context.Context, businessID, customerID int64) error {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	result, err := r.queries.WithTx(tx).DeleteCustomer(ctx, sqlc.DeleteCustomerParams{
		ID:         customerID,
		BusinessID: businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete customer", logger.Error(err))

		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return domain.ErrCustomerNotFound
	}

	// The CRM schema has no foreign keys: rows in these tables would keep
	// referencing the deleted customer forever (campaign/feedback stats join
	// on them). Tables with NOT NULL customer_id are purged; nullable
	// references are detached.
	cleanup := []string{
		"DELETE FROM crm_campaign_sends WHERE business_id = ? AND customer_id = ?",
		"DELETE FROM crm_feedback WHERE business_id = ? AND customer_id = ?",
		"DELETE FROM crm_loyalty_transactions WHERE business_id = ? AND customer_id = ?",
		"DELETE FROM crm_customer_timeline WHERE business_id = ? AND customer_id = ?",
		"UPDATE crm_contacts SET customer_id = NULL WHERE business_id = ? AND customer_id = ?",
	}
	for _, stmt := range cleanup {
		if _, err := tx.ExecContext(ctx, stmt, businessID, customerID); err != nil {
			logger.Ctx(ctx).Error("Failed to clean up customer dependencies", logger.Error(err))

			return err
		}
	}

	return tx.Commit()
}

func (r *Repository) UpdateCustomerLoyalty(ctx context.Context, businessID, customerID int64, newPoints int, newTier string) error {
	err := r.queries.UpdateCustomerLoyalty(ctx, sqlc.UpdateCustomerLoyaltyParams{
		LoyaltyPoints: sql.NullInt32{Int32: int32(newPoints), Valid: true}, // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		LoyaltyTier: sqlc.NullCrmCustomersLoyaltyTier{
			CrmCustomersLoyaltyTier: sqlc.CrmCustomersLoyaltyTier(newTier),
			Valid:                   newTier != "",
		},
		ID:         customerID,
		BusinessID: businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update customer loyalty", logger.Error(err))
	}
	return err
}

func (r *Repository) RecordVisit(ctx context.Context, businessID, customerID int64, amount float64) error {
	err := r.queries.RecordVisit(ctx, sqlc.RecordVisitParams{
		TotalSpent: sql.NullString{String: strconv.FormatFloat(amount, 'f', 2, 64), Valid: true},
		ID:         customerID,
		BusinessID: businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to record visit", logger.Error(err))
	}

	return err
}

// ─── Loyalty Transactions ─────────────────────────────────────────────────────

func (r *Repository) CreateLoyaltyTransaction(ctx context.Context, tx domain.LoyaltyTransaction) error {
	err := r.queries.CreateLoyaltyTransaction(ctx, sqlc.CreateLoyaltyTransactionParams{
		OwnerID:       tx.OwnerID,
		BusinessID:    tx.BusinessID,
		CustomerID:    tx.CustomerID,
		Type:          sqlc.CrmLoyaltyTransactionsType(tx.Type),
		Points:        int32(tx.Points),        // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		BalanceBefore: int32(tx.BalanceBefore), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		BalanceAfter:  int32(tx.BalanceAfter),  // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		Description:   sql.NullString{String: tx.Description, Valid: tx.Description != ""},
		ReferenceID:   nullInt64FromPtr(tx.ReferenceID),
		ReferenceType: sql.NullString{String: tx.ReferenceType, Valid: tx.ReferenceType != ""},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create loyalty transaction", logger.Error(err))
	}

	return err
}

// ApplyLoyaltyDelta serializes the balance update and audit insert in one
// transaction. The customer row lock prevents concurrent awards/redeems from
// calculating from the same stale balance. onCommitted runs only after the
// transaction commits, so callers can invalidate caches without racing a
// rollback.
func (r *Repository) ApplyLoyaltyDelta(ctx context.Context, input domain.LoyaltyTransaction, delta int, tierFor func(int) string, onCommitted func()) error {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var balanceBefore int
	err = tx.QueryRowContext(ctx,
		"SELECT loyalty_points FROM crm_customers WHERE id = ? AND business_id = ? FOR UPDATE",
		input.CustomerID, input.BusinessID,
	).Scan(&balanceBefore)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrCustomerNotFound
	}
	if err != nil {
		return err
	}
	balanceAfter := balanceBefore + delta
	if balanceAfter < 0 {
		return domain.ErrInsufficientPoints
	}
	if _, err = tx.ExecContext(ctx,
		"UPDATE crm_customers SET loyalty_points = ?, loyalty_tier = ? WHERE id = ? AND business_id = ?",
		balanceAfter, tierFor(balanceAfter), input.CustomerID, input.BusinessID,
	); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `
		INSERT INTO crm_loyalty_transactions
		  (owner_id, business_id, customer_id, type, points, balance_before, balance_after,
		   description, reference_id, reference_type)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		input.OwnerID, input.BusinessID, input.CustomerID, input.Type, input.Points,
		balanceBefore, balanceAfter, input.Description, input.ReferenceID, input.ReferenceType,
	); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	if onCommitted != nil {
		onCommitted()
	}
	return nil
}

func (r *Repository) ListLoyaltyTransactions(ctx context.Context, businessID, customerID int64, page, pageSize int) ([]domain.LoyaltyTransaction, int, error) {
	if page < 1 {
		page = 1
	}

	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	var err error
	var total int64
	if customerID == 0 {
		total, err = r.queries.CountAllLoyaltyTransactions(ctx, businessID)
	} else {
		total, err = r.queries.CountLoyaltyTransactions(ctx, sqlc.CountLoyaltyTransactionsParams{
			BusinessID: businessID,
			CustomerID: customerID,
		})
	}
	if err != nil {
		return nil, 0, err
	}

	var rows []sqlc.CrmLoyaltyTransaction
	if customerID == 0 {
		rows, err = r.queries.ListAllLoyaltyTransactions(ctx, sqlc.ListAllLoyaltyTransactionsParams{
			BusinessID: businessID,
			Limit:      int32(pageSize),
			Offset:     int32(offset), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		})
	} else {
		rows, err = r.queries.ListLoyaltyTransactions(ctx, sqlc.ListLoyaltyTransactionsParams{
			BusinessID: businessID,
			CustomerID: customerID,
			Limit:      int32(pageSize),
			Offset:     int32(offset), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		})
	}
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list loyalty transactions", logger.Error(err))

		return nil, 0, err
	}

	txs := make([]domain.LoyaltyTransaction, 0, len(rows))

	for _, row := range rows {
		tx := domain.LoyaltyTransaction{
			ID:            row.ID,
			OwnerID:       row.OwnerID,
			BusinessID:    row.BusinessID,
			CustomerID:    row.CustomerID,
			Type:          string(row.Type),
			Points:        int(row.Points),
			BalanceBefore: int(row.BalanceBefore),
			BalanceAfter:  int(row.BalanceAfter),
			Description:   row.Description.String,
			ReferenceID:   nullInt64Ptr(row.ReferenceID),
			ReferenceType: row.ReferenceType.String,
			CreatedAt:     row.CreatedAt.Time,
		}
		txs = append(txs, tx)
	}

	return txs, int(total), nil
}

// ─── Loyalty Config ───────────────────────────────────────────────────────────

func (r *Repository) GetLoyaltyConfig(ctx context.Context, businessID int64) (domain.LoyaltyConfig, error) {
	row, err := r.queries.GetLoyaltyConfig(ctx, businessID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.LoyaltyConfig{}, nil
		}

		logger.Ctx(ctx).Error("Failed to get loyalty config", logger.Error(err))

		return domain.LoyaltyConfig{}, err
	}

	ppcu := parseDecimal(row.PointsPerCurrencyUnit.String)
	cpp := parseDecimal(row.CurrencyPerPoint.String)
	bdp := parseDecimal(row.BronzeDiscountPct.String)
	sdp := parseDecimal(row.SilverDiscountPct.String)
	gdp := parseDecimal(row.GoldDiscountPct.String)
	pdp := parseDecimal(row.PlatinumDiscountPct.String)

	return domain.LoyaltyConfig{
		ID:                    row.ID,
		OwnerID:               row.OwnerID,
		BusinessID:            row.BusinessID,
		IsEnabled:             row.IsEnabled.Bool,
		PointsPerCurrencyUnit: ppcu,
		CurrencyPerPoint:      cpp,
		BronzeMin:             int(row.BronzeMin.Int32),
		SilverMin:             int(row.SilverMin.Int32),
		GoldMin:               int(row.GoldMin.Int32),
		PlatinumMin:           int(row.PlatinumMin.Int32),
		BronzeDiscountPct:     bdp,
		SilverDiscountPct:     sdp,
		GoldDiscountPct:       gdp,
		PlatinumDiscountPct:   pdp,
		PointsExpiryDays:      int(row.PointsExpiryDays.Int32),
		UpdatedAt:             row.UpdatedAt.Time,
	}, nil
}

func (r *Repository) UpsertLoyaltyConfig(ctx context.Context, input domain.LoyaltyConfigInput) error {
	err := r.queries.UpsertLoyaltyConfig(ctx, sqlc.UpsertLoyaltyConfigParams{
		OwnerID:               input.OwnerID,
		BusinessID:            input.BusinessID,
		IsEnabled:             sql.NullBool{Bool: input.IsEnabled, Valid: true},
		PointsPerCurrencyUnit: sql.NullString{String: strconv.FormatFloat(input.PointsPerCurrencyUnit, 'f', 2, 64), Valid: true},
		CurrencyPerPoint:      sql.NullString{String: strconv.FormatFloat(input.CurrencyPerPoint, 'f', 4, 64), Valid: true},
		BronzeMin:             sql.NullInt32{Int32: int32(input.BronzeMin), Valid: true},   // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		SilverMin:             sql.NullInt32{Int32: int32(input.SilverMin), Valid: true},   // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		GoldMin:               sql.NullInt32{Int32: int32(input.GoldMin), Valid: true},     // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		PlatinumMin:           sql.NullInt32{Int32: int32(input.PlatinumMin), Valid: true}, // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		BronzeDiscountPct:     sql.NullString{String: strconv.FormatFloat(input.BronzeDiscountPct, 'f', 2, 64), Valid: true},
		SilverDiscountPct:     sql.NullString{String: strconv.FormatFloat(input.SilverDiscountPct, 'f', 2, 64), Valid: true},
		GoldDiscountPct:       sql.NullString{String: strconv.FormatFloat(input.GoldDiscountPct, 'f', 2, 64), Valid: true},
		PlatinumDiscountPct:   sql.NullString{String: strconv.FormatFloat(input.PlatinumDiscountPct, 'f', 2, 64), Valid: true},
		PointsExpiryDays:      sql.NullInt32{Int32: int32(input.PointsExpiryDays), Valid: true}, // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to upsert loyalty config", logger.Error(err))
	}

	return err
}

// ─── Segments ─────────────────────────────────────────────────────────────────
//
// Segment persistence belongs to the segments module. These compatibility
// methods keep the existing customer routes stable while delegating to its
// public facade rather than maintaining another SQLC query surface.

func (r *Repository) ListSegments(ctx context.Context, businessID int64) ([]domain.Segment, error) {
	rows, err := r.segments.List(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list segments", logger.Error(err))

		return nil, err
	}

	segments := make([]domain.Segment, 0, len(rows))

	for _, row := range rows {
		segments = append(segments, segmentFromCanonical(row))
	}

	return segments, nil
}

func (r *Repository) GetSegmentByID(ctx context.Context, businessID, segmentID int64) (domain.Segment, error) {
	row, err := r.segments.Get(ctx, businessID, segmentID)
	if err != nil {
		if errors.Is(err, segments.ErrNotFound) {
			return domain.Segment{}, domain.ErrSegmentNotFound
		}

		return domain.Segment{}, err
	}

	return segmentFromCanonical(row), nil
}

func (r *Repository) CreateSegment(ctx context.Context, input domain.SegmentInput) (int64, error) {
	id, err := r.segments.Create(ctx, segments.SegmentInput{
		OwnerID: input.OwnerID, BusinessID: input.BusinessID, Name: input.Name, Description: input.Description, Rules: input.Rules,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create segment", logger.Error(err))

		return 0, err
	}

	return id, nil
}

func (r *Repository) UpdateSegment(ctx context.Context, input domain.SegmentUpdateInput) error {
	if input.BusinessID == 0 {
		input.BusinessID = input.OwnerID
	}

	err := r.segments.Update(ctx, segments.SegmentUpdateInput{
		ID: input.ID, OwnerID: input.OwnerID, BusinessID: input.BusinessID, Name: input.Name, Description: input.Description, Rules: input.Rules,
	})
	if err != nil {
		if errors.Is(err, segments.ErrNotFound) {
			return domain.ErrSegmentNotFound
		}
		logger.Ctx(ctx).Error("Failed to update segment", logger.Error(err))

		return err
	}

	return nil
}

func (r *Repository) DeleteSegment(ctx context.Context, businessID, segmentID int64) error {
	err := r.segments.Delete(ctx, businessID, segmentID)
	if err != nil {
		if errors.Is(err, segments.ErrNotFound) {
			return domain.ErrSegmentNotFound
		}
		logger.Ctx(ctx).Error("Failed to delete segment", logger.Error(err))

		return err
	}

	return nil
}

// ─── Campaigns ────────────────────────────────────────────────────────────────
//
// Campaign persistence belongs to the campaigns module. These compatibility
// methods retain the customer API's DTOs while delegating to the canonical
// campaign facade.

func (r *Repository) ListCampaigns(ctx context.Context, businessID int64, status string) ([]domain.Campaign, error) {
	rows, err := r.campaigns.List(ctx, businessID, status)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list campaigns", logger.Error(err))

		return nil, err
	}

	campaigns := make([]domain.Campaign, 0, len(rows))

	for _, row := range rows {
		campaigns = append(campaigns, campaignFromCanonical(row))
	}

	return campaigns, nil
}

func (r *Repository) GetCampaignByID(ctx context.Context, businessID, campaignID int64) (domain.Campaign, error) {
	row, err := r.campaigns.Get(ctx, businessID, campaignID)
	if err != nil {
		if errors.Is(err, campaigns.ErrNotFound) {
			return domain.Campaign{}, domain.ErrCampaignNotFound
		}

		return domain.Campaign{}, err
	}

	return campaignFromCanonical(row), nil
}

func (r *Repository) CreateCampaign(ctx context.Context, input domain.CampaignInput) (int64, error) {
	id, err := r.campaigns.Create(ctx, campaigns.CampaignInput{
		OwnerID: input.OwnerID, BusinessID: input.BusinessID, Name: input.Name, Description: input.Description, Channel: input.Channel,
		SegmentID: input.SegmentID, Subject: input.Subject, Content: input.Content, ScheduledAt: input.ScheduledAt,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create campaign", logger.Error(err))

		return 0, err
	}

	return id, nil
}

func (r *Repository) UpdateCampaign(ctx context.Context, input domain.CampaignUpdateInput) error {
	if input.BusinessID == 0 {
		input.BusinessID = input.OwnerID
	}

	err := r.campaigns.Update(ctx, campaigns.CampaignUpdateInput{
		ID: input.ID, OwnerID: input.OwnerID, BusinessID: input.BusinessID, Name: input.Name, Description: input.Description,
		Channel: input.Channel, SegmentID: input.SegmentID, Subject: input.Subject, Content: input.Content, ScheduledAt: input.ScheduledAt,
	})
	if err != nil {
		if errors.Is(err, campaigns.ErrNotFound) {
			return domain.ErrCampaignNotFound
		}
		logger.Ctx(ctx).Error("Failed to update campaign", logger.Error(err))

		return err
	}

	return nil
}

func (r *Repository) UpdateCampaignStatus(ctx context.Context, businessID, campaignID int64, status string) error {
	err := r.campaigns.UpdateStatus(ctx, businessID, campaignID, status)
	if err != nil {
		if errors.Is(err, campaigns.ErrNotFound) {
			return domain.ErrCampaignNotFound
		}
		logger.Ctx(ctx).Error("Failed to update campaign status", logger.Error(err))

		return err
	}

	return nil
}

// ─── Dashboard ────────────────────────────────────────────────────────────────

func (r *Repository) GetDashboardStats(ctx context.Context, businessID int64) (domain.DashboardStats, error) {
	var (
		totalCustomers int64
		newThisMonth   int64
		activeMembers  int64
		topRows        []sqlc.ListTopCustomersRow
	)

	timezone, err := r.queries.GetBusinessTimezone(ctx, businessID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Ctx(ctx).Error("Failed to get business timezone", logger.Error(err))
		return domain.DashboardStats{}, err
	}
	location := time.UTC
	if timezone != "" {
		loaded, loadErr := time.LoadLocation(timezone)
		if loadErr != nil {
			logger.Ctx(ctx).Warn("Invalid business timezone; falling back to UTC", logger.Error(loadErr))
		} else {
			location = loaded
		}
	}

	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		var err error
		totalCustomers, err = r.queries.CountActiveCustomers(groupCtx, businessID)
		if err != nil {
			logger.Ctx(groupCtx).Error("Failed to get total customers", logger.Error(err))
		}
		return err
	})
	group.Go(func() error {
		var err error
		monthStart, monthEnd := customerMonthBounds(time.Now(), location)
		newThisMonth, err = r.queries.CountNewCustomersThisMonth(groupCtx, sqlc.CountNewCustomersThisMonthParams{
			BusinessID:     businessID,
			CreatedAtStart: sql.NullTime{Time: monthStart, Valid: true},
			CreatedAtEnd:   sql.NullTime{Time: monthEnd, Valid: true},
		})
		if err != nil {
			logger.Ctx(groupCtx).Error("Failed to get new customers this month", logger.Error(err))
		}
		return err
	})
	group.Go(func() error {
		var err error
		activeMembers, err = r.queries.CountActiveLoyaltyMembers(groupCtx, businessID)
		if err != nil {
			logger.Ctx(groupCtx).Error("Failed to get active loyalty members", logger.Error(err))
		}
		return err
	})
	group.Go(func() error {
		var err error
		topRows, err = r.queries.ListTopCustomers(groupCtx, businessID)
		if err != nil {
			logger.Ctx(groupCtx).Error("Failed to get top customers", logger.Error(err))
		}
		return err
	})
	if err := group.Wait(); err != nil {
		return domain.DashboardStats{}, err
	}

	stats := domain.DashboardStats{
		TotalCustomers:       int(totalCustomers),
		NewThisMonth:         int(newThisMonth),
		ActiveLoyaltyMembers: int(activeMembers),
		TopCustomers:         make([]domain.TopCustomer, 0, len(topRows)),
	}
	for _, row := range topRows {
		spent := parseDecimal(row.TotalSpent.String)
		stats.TopCustomers = append(stats.TopCustomers, domain.TopCustomer{
			ID:            row.ID,
			FirstName:     row.FirstName,
			LastName:      row.LastName.String,
			Email:         row.Email.String,
			TotalSpent:    spent,
			LoyaltyPoints: int(row.LoyaltyPoints.Int32),
			LoyaltyTier:   string(row.LoyaltyTier.CrmCustomersLoyaltyTier),
		})
	}

	return stats, nil
}

func customerMonthBounds(now time.Time, location *time.Location) (time.Time, time.Time) {
	localNow := now.In(location)
	start := time.Date(localNow.Year(), localNow.Month(), 1, 0, 0, 0, 0, location)
	return start.UTC(), start.AddDate(0, 1, 0).UTC()
}

// ─── Timeline ─────────────────────────────────────────────────────────────────

func (r *Repository) GetCustomerTimeline(ctx context.Context, businessID, customerID int64) ([]domain.TimelineEntry, error) {
	rows, err := r.queries.GetCustomerTimeline(ctx, sqlc.GetCustomerTimelineParams{
		BusinessID: businessID,
		CustomerID: customerID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get customer timeline", logger.Error(err))

		return nil, err
	}

	entries := make([]domain.TimelineEntry, 0, len(rows))

	for _, row := range rows {
		e := domain.TimelineEntry{
			ID:          row.ID,
			OwnerID:     row.OwnerID,
			BusinessID:  row.BusinessID,
			CustomerID:  row.CustomerID,
			Type:        row.Type,
			Description: row.Description.String,
			CreatedAt:   row.CreatedAt.Time,
		}

		if row.Metadata != nil {
			e.Metadata = []byte(row.Metadata)
		} else {
			e.Metadata = []byte("{}")
		}

		entries = append(entries, e)
	}

	return entries, nil
}

func (r *Repository) ListCustomerTimelinePage(ctx context.Context, businessID, customerID int64, cursorCreatedAt *time.Time, cursorID int64, pageSize int) ([]domain.TimelineEntry, error) {
	rows, err := r.queries.ListCustomerTimelinePage(ctx, sqlc.ListCustomerTimelinePageParams{
		BusinessID:      businessID,
		CustomerID:      customerID,
		CursorCreatedAt: nullableTime(cursorCreatedAt),
		CursorID:        cursorID,
		Limit:           int32(pageSize), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list customer timeline page", logger.Error(err))
		return nil, err
	}

	entries := make([]domain.TimelineEntry, 0, len(rows))
	for _, row := range rows {
		entry := domain.TimelineEntry{
			ID:          row.ID,
			OwnerID:     row.OwnerID,
			BusinessID:  row.BusinessID,
			CustomerID:  row.CustomerID,
			Type:        row.Type,
			Description: row.Description.String,
			CreatedAt:   row.CreatedAt.Time,
		}
		if row.Metadata != nil {
			entry.Metadata = []byte(row.Metadata)
		} else {
			entry.Metadata = []byte("{}")
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func (r *Repository) InsertTimelineEntry(ctx context.Context, entry domain.TimelineEntry) error {
	err := r.queries.InsertTimelineEntry(ctx, sqlc.InsertTimelineEntryParams{
		OwnerID:     entry.OwnerID,
		BusinessID:  entry.BusinessID,
		CustomerID:  entry.CustomerID,
		Type:        entry.Type,
		Description: sql.NullString{String: entry.Description, Valid: entry.Description != ""},
		Metadata:    json.RawMessage(entry.Metadata),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to insert timeline entry", logger.Error(err))
	}

	return err
}

// ─── Import ──────────────────────────────────────────────────────────────────

func (r *Repository) BulkCreateCustomers(ctx context.Context, businessID, ownerID int64, items []domain.CustomerImportItem, dedupe bool) (int, int, error) {
	imported := 0
	failed := 0

	for _, item := range items {
		if dedupe && item.Email != "" {
			var existing int
			if err := r.pool.QueryRowContext(ctx,
				"SELECT COUNT(*) FROM crm_customers WHERE business_id = ? AND email = ?",
				businessID, item.Email,
			).Scan(&existing); err != nil {
				failed++
				continue
			}
			if existing > 0 {
				continue
			}
		}
		var dob *time.Time

		if item.DateOfBirth != "" {
			t, err := time.Parse("2006-01-02", item.DateOfBirth)
			if err == nil {
				dob = &t
			}
		}

		tagsJSON, err := json.Marshal(item.Tags)
		if err != nil {
			tagsJSON = []byte("[]")
		}

		_, err = r.queries.CreateCustomer(ctx, sqlc.CreateCustomerParams{
			OwnerID:     ownerID,
			BusinessID:  businessID,
			FirstName:   item.FirstName,
			LastName:    sql.NullString{String: item.LastName, Valid: item.LastName != ""},
			Email:       sql.NullString{String: item.Email, Valid: item.Email != ""},
			Phone:       sql.NullString{String: item.Phone, Valid: item.Phone != ""},
			DateOfBirth: nullableTime(dob),
			Gender:      sqlc.NullCrmCustomersGender{CrmCustomersGender: sqlc.CrmCustomersGender(item.Gender), Valid: item.Gender != ""},
			Address:     sql.NullString{String: item.Address, Valid: item.Address != ""},
			City:        sql.NullString{String: item.City, Valid: item.City != ""},
			Country:     sql.NullString{String: item.Country, Valid: item.Country != ""},
			Tags:        tagsJSON,
			Notes:       sql.NullString{String: item.Notes, Valid: item.Notes != ""},
		})
		if err != nil {
			logger.Ctx(ctx).Error("Failed to import customer", logger.Error(err))
			failed++
		} else {
			imported++
		}
	}

	return imported, failed, nil
}

// ─── Export (GDPR) ──────────────────────────────────────────────────────────

func (r *Repository) ExportCustomerData(ctx context.Context, businessID, customerID int64) (domain.CustomerExport, error) {
	customer, err := r.GetCustomerByID(ctx, businessID, customerID)
	if err != nil {
		return domain.CustomerExport{}, err
	}

	txs, _, err := r.ListLoyaltyTransactions(ctx, businessID, customerID, 1, 100)
	if err != nil {
		return domain.CustomerExport{}, err
	}

	timeline, err := r.GetCustomerTimeline(ctx, businessID, customerID)
	if err != nil {
		return domain.CustomerExport{}, err
	}

	return domain.CustomerExport{
		Customer:     customer,
		Transactions: txs,
		Timeline:     timeline,
	}, nil
}

// ─── Merge ──────────────────────────────────────────────────────────────────

func (r *Repository) MergeCustomers(ctx context.Context, input domain.MergeCustomersInput) error {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	queries := r.queries.WithTx(tx)

	// Lock both rows (in deterministic id order to avoid deadlocks between
	// concurrent merges sharing a customer) so that an award/redeem/visit
	// committed between the stat reads below and UpdateCustomerMergeStats
	// cannot be silently lost.
	lockOrder := []int64{input.PrimaryID, input.SecondaryID}
	if lockOrder[1] < lockOrder[0] {
		lockOrder[0], lockOrder[1] = lockOrder[1], lockOrder[0]
	}
	for _, id := range lockOrder {
		if _, err := tx.ExecContext(ctx,
			"SELECT id FROM crm_customers WHERE id = ? AND business_id = ? FOR UPDATE",
			id, input.BusinessID,
		); err != nil {
			logger.Ctx(ctx).Error("Failed to lock customer during merge", logger.Error(err))

			return err
		}
	}

	secondaryRow, err := queries.GetCustomerByID(ctx, sqlc.GetCustomerByIDParams{ID: input.SecondaryID, BusinessID: input.BusinessID})
	if err != nil {
		return err
	}
	primaryRow, err := queries.GetCustomerByID(ctx, sqlc.GetCustomerByIDParams{ID: input.PrimaryID, BusinessID: input.BusinessID})
	if err != nil {
		return err
	}
	secondary := customerFromRow(secondaryRow)
	primary := customerFromRow(primaryRow)

	newPoints := primary.LoyaltyPoints + secondary.LoyaltyPoints
	newSpent := primary.TotalSpent + secondary.TotalSpent
	newVisits := primary.VisitCount + secondary.VisitCount

	// The columns are INT: clamp before the int32 cast instead of wrapping.
	if newPoints > math.MaxInt32 {
		newPoints = math.MaxInt32
	}
	if newVisits > math.MaxInt32 {
		newVisits = math.MaxInt32
	}

	err = queries.UpdateCustomerMergeStats(ctx, sqlc.UpdateCustomerMergeStatsParams{
		LoyaltyPoints: sql.NullInt32{Int32: int32(newPoints), Valid: true},
		TotalSpent:    sql.NullString{String: strconv.FormatFloat(newSpent, 'f', 2, 64), Valid: true},
		VisitCount:    sql.NullInt32{Int32: int32(newVisits), Valid: true},
		ID:            input.PrimaryID,
		BusinessID:    input.BusinessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update primary customer during merge", logger.Error(err))

		return err
	}

	err = queries.MigrateLoyaltyTransactions(ctx, sqlc.MigrateLoyaltyTransactionsParams{
		CustomerID:   input.PrimaryID,
		CustomerID_2: input.SecondaryID,
		BusinessID:   input.BusinessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to migrate loyalty transactions during merge", logger.Error(err))

		return err
	}

	err = queries.MigrateTimeline(ctx, sqlc.MigrateTimelineParams{
		CustomerID:   input.PrimaryID,
		CustomerID_2: input.SecondaryID,
		BusinessID:   input.BusinessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to migrate timeline during merge", logger.Error(err))

		return err
	}

	// The CRM schema has no foreign keys: without migrating these, the
	// survivor loses the secondary's campaign history and feedback, and
	// lead contacts keep pointing at a deleted row.
	for _, stmt := range []string{
		"UPDATE crm_campaign_sends SET customer_id = ? WHERE customer_id = ? AND business_id = ?",
		"UPDATE crm_feedback SET customer_id = ? WHERE customer_id = ? AND business_id = ?",
		"UPDATE crm_contacts SET customer_id = ? WHERE customer_id = ? AND business_id = ?",
	} {
		if _, err := tx.ExecContext(ctx, stmt, input.PrimaryID, input.SecondaryID, input.BusinessID); err != nil {
			logger.Ctx(ctx).Error("Failed to migrate customer references during merge", logger.Error(err))

			return err
		}
	}

	_, err = queries.DeleteCustomer(ctx, sqlc.DeleteCustomerParams{
		ID:         input.SecondaryID,
		BusinessID: input.BusinessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete secondary customer during merge", logger.Error(err))

		return err
	}

	return tx.Commit()
}

// ─── Loyalty Tiers ──────────────────────────────────────────────────────────

func (r *Repository) ListLoyaltyTiers(ctx context.Context, businessID int64) ([]domain.LoyaltyTier, error) {
	rows, err := r.queries.ListLoyaltyTiers(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list loyalty tiers", logger.Error(err))

		return nil, err
	}

	tiers := make([]domain.LoyaltyTier, 0, len(rows))

	for _, row := range rows {
		dp := parseDecimal(row.DiscountPct.String)

		t := domain.LoyaltyTier{
			ID:          row.ID,
			OwnerID:     row.OwnerID,
			BusinessID:  row.BusinessID,
			Name:        row.Name,
			MinPoints:   int(row.MinPoints.Int32),
			DiscountPct: dp,
			SortOrder:   int(row.SortOrder.Int32),
			CreatedAt:   row.CreatedAt.Time,
			UpdatedAt:   row.UpdatedAt.Time,
		}

		if row.Perks != nil {
			t.Perks = []byte(row.Perks)
		} else {
			t.Perks = []byte("[]")
		}

		tiers = append(tiers, t)
	}

	return tiers, nil
}

func (r *Repository) CreateLoyaltyTier(ctx context.Context, input domain.LoyaltyTierInput) (int64, error) {
	result, err := r.queries.CreateLoyaltyTier(ctx, sqlc.CreateLoyaltyTierParams{
		OwnerID:     input.OwnerID,
		BusinessID:  input.BusinessID,
		Name:        input.Name,
		MinPoints:   sql.NullInt32{Int32: int32(input.MinPoints), Valid: true}, // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		DiscountPct: sql.NullString{String: strconv.FormatFloat(input.DiscountPct, 'f', 2, 64), Valid: true},
		Perks:       json.RawMessage(input.Perks),
		SortOrder:   sql.NullInt32{Int32: int32(input.SortOrder), Valid: true}, // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create loyalty tier", logger.Error(err))

		return 0, err
	}

	return result.LastInsertId()
}

func (r *Repository) UpdateLoyaltyTier(ctx context.Context, input domain.LoyaltyTierUpdateInput) error {
	if input.BusinessID == 0 {
		input.BusinessID = input.OwnerID
	}

	result, err := r.queries.UpdateLoyaltyTier(ctx, sqlc.UpdateLoyaltyTierParams{
		Name:        input.Name,
		MinPoints:   sql.NullInt32{Int32: int32(input.MinPoints), Valid: true}, // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		DiscountPct: sql.NullString{String: strconv.FormatFloat(input.DiscountPct, 'f', 2, 64), Valid: true},
		Perks:       json.RawMessage(input.Perks),
		SortOrder:   sql.NullInt32{Int32: int32(input.SortOrder), Valid: true}, // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		ID:          input.ID,
		BusinessID:  input.BusinessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update loyalty tier", logger.Error(err))

		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return domain.ErrTierNotFound
	}

	return nil
}

func (r *Repository) DeleteLoyaltyTier(ctx context.Context, businessID, tierID int64) error {
	tier, err := r.queries.GetLoyaltyTierByID(ctx, sqlc.GetLoyaltyTierByIDParams{
		ID:         tierID,
		BusinessID: businessID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrTierNotFound
		}

		logger.Ctx(ctx).Error("Failed to look up loyalty tier", logger.Error(err), logger.Any("business_id", businessID), logger.Any("tier_id", tierID))

		return err
	}

	// cross-schema read from crm_customers — active members on this tier
	count, err := r.queries.CountActiveMembersOnTier(ctx, sqlc.CountActiveMembersOnTierParams{
		BusinessID:  businessID,
		LoyaltyTier: sqlc.NullCrmCustomersLoyaltyTier{CrmCustomersLoyaltyTier: sqlc.CrmCustomersLoyaltyTier(strings.ToLower(strings.TrimSpace(tier.Name))), Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to count active members on tier", logger.Error(err), logger.Any("business_id", businessID), logger.Any("tier_id", tierID))

		return err
	}

	if count > 0 {
		return domain.ErrTierInUse
	}

	result, err := r.queries.DeleteLoyaltyTier(ctx, sqlc.DeleteLoyaltyTierParams{
		ID:         tierID,
		BusinessID: businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete loyalty tier", logger.Error(err), logger.Any("business_id", businessID), logger.Any("tier_id", tierID))

		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return domain.ErrTierNotFound
	}

	return nil
}

// ─── Campaign Stats ─────────────────────────────────────────────────────────

func (r *Repository) GetCampaignStats(ctx context.Context, businessID, campaignID int64) (domain.CampaignStats, error) {
	campaign, err := r.GetCampaignByID(ctx, businessID, campaignID)
	if err != nil {
		return domain.CampaignStats{}, err
	}

	var stats domain.CampaignStats
	stats.CampaignID = campaign.ID
	stats.AudienceCount = campaign.AudienceCount
	stats.SentCount = campaign.SentCount
	stats.DeliveredCount = campaign.DeliveredCount
	stats.OpenedCount = campaign.OpenedCount
	stats.ClickedCount = campaign.ClickedCount

	sendStats, err := r.queries.GetCampaignSendStats(ctx, sqlc.GetCampaignSendStatsParams{
		CampaignID: campaignID,
		BusinessID: businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get campaign send stats", logger.Error(err))
	} else {
		if v, ok := toInt(sendStats.BouncedCount); ok {
			stats.BouncedCount = v
		}
		if v, ok := toInt(sendStats.FailedCount); ok {
			stats.FailedCount = v
		}
	}

	return stats, nil
}

// ─── Segment Customers ─────────────────────────────────────────────────────

func (r *Repository) GetSegmentCustomers(ctx context.Context, businessID, segmentID int64, page, pageSize int) ([]domain.Customer, int, error) {
	if page < 1 {
		page = 1
	}

	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	segment, err := r.GetSegmentByID(ctx, businessID, segmentID)
	if err != nil {
		return nil, 0, err
	}

	conds := sq.And{
		sq.Eq{"business_id": businessID},
		sq.Eq{"is_active": true},
	}
	// Real rule-based membership: without this, every segment returned all
	// active customers. Segments with no (valid) rules keep the legacy
	// "all active customers" behavior.
	if frag, args := segmentrules.Filter(segment.Rules); frag != "" {
		conds = append(conds, sq.Expr(frag, args...))
	}

	countSQL, countArgs, err := sq.Select("COUNT(*)").
		From("crm_customers").
		Where(conds).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return nil, 0, err
	}

	var total int
	if err := r.pool.QueryRowContext(ctx, countSQL, countArgs...).Scan(&total); err != nil {
		logger.Ctx(ctx).Error("Failed to count segment customers", logger.Error(err))

		return nil, 0, err
	}

	listSQL, listArgs, err := sq.Select(
		"id", "owner_id", "business_id", "first_name", "last_name", "email", "phone", "date_of_birth", "gender",
		"address", "city", "country", "loyalty_points", "total_spent", "visit_count", "last_visit_at",
		"loyalty_tier", "tags", "notes", "consent", "is_active", "created_at", "updated_at",
	).
		From("crm_customers").
		Where(conds).
		OrderBy("created_at DESC").
		Limit(uint64(pageSize)).
		Offset(uint64(offset)). // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.pool.QueryContext(ctx, listSQL, listArgs...)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get segment customers", logger.Error(err))

		return nil, 0, err
	}
	defer rows.Close()

	customers := make([]domain.Customer, 0, pageSize)

	for rows.Next() {
		c, scanErr := scanCustomerRaw(rows)
		if scanErr != nil {
			return nil, 0, scanErr
		}
		customers = append(customers, c)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return customers, total, nil
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func customerFromRow(row sqlc.CrmCustomer) domain.Customer {
	c := domain.Customer{
		ID:            row.ID,
		OwnerID:       row.OwnerID,
		BusinessID:    row.BusinessID,
		FirstName:     row.FirstName,
		LastName:      row.LastName.String,
		Email:         row.Email.String,
		Phone:         row.Phone.String,
		Gender:        string(row.Gender.CrmCustomersGender),
		Address:       row.Address.String,
		City:          row.City.String,
		Country:       row.Country.String,
		LoyaltyPoints: int(row.LoyaltyPoints.Int32),
		VisitCount:    int(row.VisitCount.Int32),
		LoyaltyTier:   string(row.LoyaltyTier.CrmCustomersLoyaltyTier),
		Notes:         row.Notes.String,
		IsActive:      row.IsActive.Bool,
		CreatedAt:     row.CreatedAt.Time,
		UpdatedAt:     row.UpdatedAt.Time,
	}

	if row.Consent != nil {
		c.Consent = append([]byte(nil), row.Consent...)
	}

	spent := parseDecimal(row.TotalSpent.String)
	c.TotalSpent = spent

	if row.DateOfBirth.Valid {
		t := row.DateOfBirth.Time
		c.DateOfBirth = &t
	}

	if row.LastVisitAt.Valid {
		t := row.LastVisitAt.Time
		c.LastVisitAt = &t
	}

	if row.Tags != nil {
		if err := json.Unmarshal(row.Tags, &c.Tags); err != nil {
			logger.Global().Warn("failed to parse customer tags JSON from database",
				logger.Any("customer_id", row.ID), logger.Error(err))
		}
	}

	if c.Tags == nil {
		c.Tags = []string{}
	}

	return c
}

func segmentFromCanonical(row segments.Segment) domain.Segment {
	return domain.Segment{
		ID: row.ID, OwnerID: row.OwnerID, BusinessID: row.BusinessID, Name: row.Name, Description: row.Description,
		Rules: row.Rules, CustomerCount: row.CustomerCount, LastCalculatedAt: row.LastCalculatedAt,
		CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
}

func campaignFromCanonical(row campaigns.Campaign) domain.Campaign {
	return domain.Campaign{
		ID: row.ID, OwnerID: row.OwnerID, BusinessID: row.BusinessID, Name: row.Name, Description: row.Description,
		Channel: row.Channel, Status: row.Status, SegmentID: row.SegmentID, Subject: row.Subject, Content: row.Content,
		ScheduledAt: row.ScheduledAt, SentAt: row.SentAt, AudienceCount: row.AudienceCount, SentCount: row.SentCount,
		DeliveredCount: row.DeliveredCount, OpenedCount: row.OpenedCount, ClickedCount: row.ClickedCount,
		CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
}

func scanCustomerRaw(rows *sql.Rows) (domain.Customer, error) {
	var c domain.Customer
	var (
		lastName    sql.NullString
		email       sql.NullString
		phone       sql.NullString
		dob         sql.NullTime
		gender      sql.NullString
		address     sql.NullString
		city        sql.NullString
		country     sql.NullString
		lastVisitAt sql.NullTime
		tagsJSON    sql.NullString
		notes       sql.NullString
		consent     []byte
	)

	if err := rows.Scan(
		&c.ID, &c.OwnerID, &c.BusinessID, &c.FirstName, &lastName, &email, &phone, &dob, &gender,
		&address, &city, &country, &c.LoyaltyPoints, &c.TotalSpent, &c.VisitCount, &lastVisitAt,
		&c.LoyaltyTier, &tagsJSON, &notes, &consent, &c.IsActive, &c.CreatedAt, &c.UpdatedAt,
	); err != nil {
		return domain.Customer{}, err
	}

	c.LastName = lastName.String
	c.Email = email.String
	c.Phone = phone.String

	if dob.Valid {
		t := dob.Time
		c.DateOfBirth = &t
	}

	c.Gender = gender.String
	c.Address = address.String
	c.City = city.String
	c.Country = country.String

	if lastVisitAt.Valid {
		t := lastVisitAt.Time
		c.LastVisitAt = &t
	}

	c.Notes = notes.String
	if consent != nil {
		c.Consent = append([]byte(nil), consent...)
	}

	if tagsJSON.Valid && tagsJSON.String != "" {
		if err := json.Unmarshal([]byte(tagsJSON.String), &c.Tags); err != nil {
			logger.Global().Warn("failed to parse customer tags JSON from database",
				logger.Any("customer_id", c.ID), logger.Error(err))
		}
	}

	if c.Tags == nil {
		c.Tags = []string{}
	}

	return c, nil
}

func nullableTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{}
	}

	return sql.NullTime{Time: *t, Valid: true}
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

func toInt(v interface{}) (int, bool) {
	switch val := v.(type) {
	case int64:
		return int(val), true
	case int32:
		return int(val), true
	case int:
		return val, true
	case []byte:
		n, err := strconv.Atoi(string(val))
		if err == nil {
			return n, true
		}
	}

	return 0, false
}

// parseDecimal converts a DB decimal string to float64, logging corrupt
// values instead of silently coercing them to 0.
func parseDecimal(s string) float64 {
	if s == "" {
		return 0
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		logger.Global().Warn("failed to parse customer decimal from database",
			logger.String("value", s), logger.Error(err))
		return 0
	}
	return f
}
