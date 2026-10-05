package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"

	sq "github.com/Masterminds/squirrel"

	"github.com/voxire/onion/internal/modules/crm/campaigns"
	"github.com/voxire/onion/internal/modules/crm/customers/internal/domain"
	"github.com/voxire/onion/internal/modules/crm/customers/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/modules/crm/segments"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/searchutil"
)

type Repository struct {
	reportingTimezone func(context.Context, int64) (string, error)
	queries           *sqlc.Queries
	pool              *sql.DB
	campaigns         *campaigns.Module
	segments          *segments.Module
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

func (r *Repository) WithReportingTimezone(read func(context.Context, int64) (string, error)) *Repository {
	r.reportingTimezone = read
	return r
}
