package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	sq "github.com/Masterminds/squirrel"

	"github.com/voxire/onion/internal/modules/customers/internal/domain"
	"github.com/voxire/onion/internal/modules/customers/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	queries *sqlc.Queries
	pool    *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	pool := db

	return &Repository{
		queries: sqlc.New(pool),
		pool:    pool,
	}
}

// ─── Customers ───────────────────────────────────────────────────────────────

func (r *Repository) ListCustomers(ctx context.Context, filter domain.CustomerListFilter) ([]domain.Customer, int, error) {
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

	total, err := r.queries.CountCustomers(ctx, filter.OwnerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to count customers", logger.Error(err))

		return nil, 0, err
	}

	rows, err := r.queries.ListCustomers(ctx, sqlc.ListCustomersParams{
		OwnerID: filter.OwnerID,
		Limit:   int32(filter.PageSize),
		Offset:  int32(offset),
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

func (r *Repository) listCustomersFiltered(ctx context.Context, filter domain.CustomerListFilter, offset int) ([]domain.Customer, int, error) {
	conds := sq.And{sq.Eq{"owner_id": filter.OwnerID}}

	if filter.Search != "" {
		like := "%" + filter.Search + "%"
		conds = append(conds, sq.Expr(
			"(first_name LIKE ? OR last_name LIKE ? OR email LIKE ? OR phone LIKE ?)",
			like, like, like, like,
		))
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
		"id", "owner_id", "first_name", "last_name", "email", "phone", "date_of_birth", "gender",
		"address", "city", "country", "loyalty_points", "total_spent", "visit_count", "last_visit_at",
		"loyalty_tier", "tags", "notes", "is_active", "created_at", "updated_at",
	).
		From("crm_customers").
		Where(conds).
		OrderBy("created_at DESC").
		Limit(uint64(filter.PageSize)).
		Offset(uint64(offset)).
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

func (r *Repository) GetCustomerByID(ctx context.Context, ownerID, customerID int64) (domain.Customer, error) {
	row, err := r.queries.GetCustomerByID(ctx, sqlc.GetCustomerByIDParams{
		ID:      customerID,
		OwnerID: ownerID,
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
	tagsJSON, err := json.Marshal(input.Tags)
	if err != nil {
		tagsJSON = []byte("[]")
	}

	isActive := true
	if input.IsActive != nil {
		isActive = *input.IsActive
	}

	result, err := r.queries.UpdateCustomer(ctx, sqlc.UpdateCustomerParams{
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
		IsActive:    sql.NullBool{Bool: isActive, Valid: true},
		ID:          input.ID,
		OwnerID:     input.OwnerID,
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

func (r *Repository) DeleteCustomer(ctx context.Context, ownerID, customerID int64) error {
	result, err := r.queries.DeleteCustomer(ctx, sqlc.DeleteCustomerParams{
		ID:      customerID,
		OwnerID: ownerID,
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

	return nil
}

func (r *Repository) UpdateCustomerLoyalty(ctx context.Context, customerID int64, newPoints int, newTier string) error {
	err := r.queries.UpdateCustomerLoyalty(ctx, sqlc.UpdateCustomerLoyaltyParams{
		LoyaltyPoints: sql.NullInt32{Int32: int32(newPoints), Valid: true},
		LoyaltyTier:   sqlc.NullCrmCustomersLoyaltyTier{CrmCustomersLoyaltyTier: sqlc.CrmCustomersLoyaltyTier(newTier), Valid: true},
		ID:            customerID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update customer loyalty", logger.Error(err))
	}

	return err
}

func (r *Repository) RecordVisit(ctx context.Context, ownerID, customerID int64, amount float64) error {
	err := r.queries.RecordVisit(ctx, sqlc.RecordVisitParams{
		TotalSpent: sql.NullString{String: strconv.FormatFloat(amount, 'f', 2, 64), Valid: true},
		ID:         customerID,
		OwnerID:    ownerID,
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
		CustomerID:    tx.CustomerID,
		Type:          sqlc.CrmLoyaltyTransactionsType(tx.Type),
		Points:        int32(tx.Points),
		BalanceBefore: int32(tx.BalanceBefore),
		BalanceAfter:  int32(tx.BalanceAfter),
		Description:   sql.NullString{String: tx.Description, Valid: tx.Description != ""},
		ReferenceID:   nullInt64FromPtr(tx.ReferenceID),
		ReferenceType: sql.NullString{String: tx.ReferenceType, Valid: tx.ReferenceType != ""},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create loyalty transaction", logger.Error(err))
	}

	return err
}

func (r *Repository) ListLoyaltyTransactions(ctx context.Context, ownerID, customerID int64, page, pageSize int) ([]domain.LoyaltyTransaction, int, error) {
	if page < 1 {
		page = 1
	}

	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	total, err := r.queries.CountLoyaltyTransactions(ctx, sqlc.CountLoyaltyTransactionsParams{
		OwnerID:    ownerID,
		CustomerID: customerID,
	})
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.queries.ListLoyaltyTransactions(ctx, sqlc.ListLoyaltyTransactionsParams{
		OwnerID:    ownerID,
		CustomerID: customerID,
		Limit:      int32(pageSize),
		Offset:     int32(offset),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list loyalty transactions", logger.Error(err))

		return nil, 0, err
	}

	txs := make([]domain.LoyaltyTransaction, 0, len(rows))

	for _, row := range rows {
		tx := domain.LoyaltyTransaction{
			ID:            row.ID,
			OwnerID:       row.OwnerID,
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

func (r *Repository) GetLoyaltyConfig(ctx context.Context, ownerID int64) (domain.LoyaltyConfig, error) {
	row, err := r.queries.GetLoyaltyConfig(ctx, ownerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.LoyaltyConfig{}, nil
		}

		logger.Ctx(ctx).Error("Failed to get loyalty config", logger.Error(err))

		return domain.LoyaltyConfig{}, err
	}

	ppcu, _ := strconv.ParseFloat(row.PointsPerCurrencyUnit.String, 64)
	cpp, _ := strconv.ParseFloat(row.CurrencyPerPoint.String, 64)
	bdp, _ := strconv.ParseFloat(row.BronzeDiscountPct.String, 64)
	sdp, _ := strconv.ParseFloat(row.SilverDiscountPct.String, 64)
	gdp, _ := strconv.ParseFloat(row.GoldDiscountPct.String, 64)
	pdp, _ := strconv.ParseFloat(row.PlatinumDiscountPct.String, 64)

	return domain.LoyaltyConfig{
		ID:                    row.ID,
		OwnerID:               row.OwnerID,
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
		IsEnabled:             sql.NullBool{Bool: input.IsEnabled, Valid: true},
		PointsPerCurrencyUnit: sql.NullString{String: strconv.FormatFloat(input.PointsPerCurrencyUnit, 'f', 2, 64), Valid: true},
		CurrencyPerPoint:      sql.NullString{String: strconv.FormatFloat(input.CurrencyPerPoint, 'f', 4, 64), Valid: true},
		BronzeMin:             sql.NullInt32{Int32: int32(input.BronzeMin), Valid: true},
		SilverMin:             sql.NullInt32{Int32: int32(input.SilverMin), Valid: true},
		GoldMin:               sql.NullInt32{Int32: int32(input.GoldMin), Valid: true},
		PlatinumMin:           sql.NullInt32{Int32: int32(input.PlatinumMin), Valid: true},
		BronzeDiscountPct:     sql.NullString{String: strconv.FormatFloat(input.BronzeDiscountPct, 'f', 2, 64), Valid: true},
		SilverDiscountPct:     sql.NullString{String: strconv.FormatFloat(input.SilverDiscountPct, 'f', 2, 64), Valid: true},
		GoldDiscountPct:       sql.NullString{String: strconv.FormatFloat(input.GoldDiscountPct, 'f', 2, 64), Valid: true},
		PlatinumDiscountPct:   sql.NullString{String: strconv.FormatFloat(input.PlatinumDiscountPct, 'f', 2, 64), Valid: true},
		PointsExpiryDays:      sql.NullInt32{Int32: int32(input.PointsExpiryDays), Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to upsert loyalty config", logger.Error(err))
	}

	return err
}

// ─── Segments ─────────────────────────────────────────────────────────────────

func (r *Repository) ListSegments(ctx context.Context, ownerID int64) ([]domain.Segment, error) {
	rows, err := r.queries.ListSegments(ctx, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list segments", logger.Error(err))

		return nil, err
	}

	segments := make([]domain.Segment, 0, len(rows))

	for _, row := range rows {
		segments = append(segments, segmentFromRow(row))
	}

	return segments, nil
}

func (r *Repository) GetSegmentByID(ctx context.Context, ownerID, segmentID int64) (domain.Segment, error) {
	row, err := r.queries.GetSegmentByID(ctx, sqlc.GetSegmentByIDParams{
		ID:      segmentID,
		OwnerID: ownerID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Segment{}, domain.ErrSegmentNotFound
		}

		return domain.Segment{}, err
	}

	return segmentFromRow(row), nil
}

func (r *Repository) CreateSegment(ctx context.Context, input domain.SegmentInput) (int64, error) {
	result, err := r.queries.CreateSegment(ctx, sqlc.CreateSegmentParams{
		OwnerID:     input.OwnerID,
		Name:        input.Name,
		Description: sql.NullString{String: input.Description, Valid: input.Description != ""},
		Rules:       json.RawMessage(input.Rules),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create segment", logger.Error(err))

		return 0, err
	}

	return result.LastInsertId()
}

func (r *Repository) UpdateSegment(ctx context.Context, input domain.SegmentUpdateInput) error {
	result, err := r.queries.UpdateSegment(ctx, sqlc.UpdateSegmentParams{
		Name:        input.Name,
		Description: sql.NullString{String: input.Description, Valid: input.Description != ""},
		Rules:       json.RawMessage(input.Rules),
		ID:          input.ID,
		OwnerID:     input.OwnerID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update segment", logger.Error(err))

		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return domain.ErrSegmentNotFound
	}

	return nil
}

func (r *Repository) DeleteSegment(ctx context.Context, ownerID, segmentID int64) error {
	result, err := r.queries.DeleteSegment(ctx, sqlc.DeleteSegmentParams{
		ID:      segmentID,
		OwnerID: ownerID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete segment", logger.Error(err))

		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return domain.ErrSegmentNotFound
	}

	return nil
}

// ─── Campaigns ────────────────────────────────────────────────────────────────

func (r *Repository) ListCampaigns(ctx context.Context, ownerID int64, status string) ([]domain.Campaign, error) {
	if status != "" {
		return r.listCampaignsFiltered(ctx, ownerID, status)
	}

	rows, err := r.queries.ListCampaigns(ctx, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list campaigns", logger.Error(err))

		return nil, err
	}

	campaigns := make([]domain.Campaign, 0, len(rows))

	for _, row := range rows {
		campaigns = append(campaigns, campaignFromRow(row))
	}

	return campaigns, nil
}

func (r *Repository) listCampaignsFiltered(ctx context.Context, ownerID int64, status string) ([]domain.Campaign, error) {
	listSQL, listArgs, err := sq.Select(
		"id", "owner_id", "name", "description", "channel", "status", "segment_id", "subject",
		"content", "scheduled_at", "sent_at", "audience_count", "sent_count", "delivered_count",
		"opened_count", "clicked_count", "created_at", "updated_at",
	).
		From("crm_campaigns").
		Where(sq.Eq{"owner_id": ownerID, "status": status}).
		OrderBy("created_at DESC").
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("Failed to build campaign list query", logger.Error(err))

		return nil, err
	}

	rows, err := r.pool.QueryContext(ctx, listSQL, listArgs...)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list campaigns", logger.Error(err))

		return nil, err
	}
	defer rows.Close()

	campaigns := make([]domain.Campaign, 0)

	for rows.Next() {
		c, scanErr := scanCampaignRaw(rows)
		if scanErr != nil {
			return nil, scanErr
		}

		campaigns = append(campaigns, c)
	}

	return campaigns, rows.Err()
}

func (r *Repository) GetCampaignByID(ctx context.Context, ownerID, campaignID int64) (domain.Campaign, error) {
	row, err := r.queries.GetCampaignByID(ctx, sqlc.GetCampaignByIDParams{
		ID:      campaignID,
		OwnerID: ownerID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Campaign{}, domain.ErrCampaignNotFound
		}

		return domain.Campaign{}, err
	}

	return campaignFromRow(row), nil
}

func (r *Repository) CreateCampaign(ctx context.Context, input domain.CampaignInput) (int64, error) {
	var segmentID sql.NullInt64
	if input.SegmentID != nil {
		segmentID = sql.NullInt64{Int64: *input.SegmentID, Valid: true}
	}

	result, err := r.queries.CreateCampaign(ctx, sqlc.CreateCampaignParams{
		OwnerID:     input.OwnerID,
		Name:        input.Name,
		Description: sql.NullString{String: input.Description, Valid: input.Description != ""},
		Channel:     sqlc.CrmCampaignsChannel(input.Channel),
		SegmentID:   segmentID,
		Subject:     sql.NullString{String: input.Subject, Valid: input.Subject != ""},
		Content:     input.Content,
		ScheduledAt: nullableTime(input.ScheduledAt),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create campaign", logger.Error(err))

		return 0, err
	}

	return result.LastInsertId()
}

func (r *Repository) UpdateCampaign(ctx context.Context, input domain.CampaignUpdateInput) error {
	var segmentID sql.NullInt64
	if input.SegmentID != nil {
		segmentID = sql.NullInt64{Int64: *input.SegmentID, Valid: true}
	}

	result, err := r.queries.UpdateCampaign(ctx, sqlc.UpdateCampaignParams{
		Name:        input.Name,
		Description: sql.NullString{String: input.Description, Valid: input.Description != ""},
		Channel:     sqlc.CrmCampaignsChannel(input.Channel),
		SegmentID:   segmentID,
		Subject:     sql.NullString{String: input.Subject, Valid: input.Subject != ""},
		Content:     input.Content,
		ScheduledAt: nullableTime(input.ScheduledAt),
		ID:          input.ID,
		OwnerID:     input.OwnerID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update campaign", logger.Error(err))

		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return domain.ErrCampaignNotFound
	}

	return nil
}

func (r *Repository) UpdateCampaignStatus(ctx context.Context, ownerID, campaignID int64, status string) error {
	result, err := r.queries.UpdateCampaignStatus(ctx, sqlc.UpdateCampaignStatusParams{
		Status:  sqlc.NullCrmCampaignsStatus{CrmCampaignsStatus: sqlc.CrmCampaignsStatus(status), Valid: true},
		ID:      campaignID,
		OwnerID: ownerID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update campaign status", logger.Error(err))

		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return domain.ErrCampaignNotFound
	}

	return nil
}

// ─── Dashboard ────────────────────────────────────────────────────────────────

func (r *Repository) GetDashboardStats(ctx context.Context, ownerID int64) (domain.DashboardStats, error) {
	var stats domain.DashboardStats

	totalCustomers, err := r.queries.CountActiveCustomers(ctx, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get total customers", logger.Error(err))

		return domain.DashboardStats{}, err
	}

	stats.TotalCustomers = int(totalCustomers)

	newThisMonth, err := r.queries.CountNewCustomersThisMonth(ctx, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get new customers this month", logger.Error(err))

		return domain.DashboardStats{}, err
	}

	stats.NewThisMonth = int(newThisMonth)

	activeMembers, err := r.queries.CountActiveLoyaltyMembers(ctx, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get active loyalty members", logger.Error(err))

		return domain.DashboardStats{}, err
	}

	stats.ActiveLoyaltyMembers = int(activeMembers)

	topRows, err := r.queries.ListTopCustomers(ctx, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get top customers", logger.Error(err))

		return domain.DashboardStats{}, err
	}

	stats.TopCustomers = make([]domain.TopCustomer, 0, len(topRows))
	for _, row := range topRows {
		spent, _ := strconv.ParseFloat(row.TotalSpent.String, 64)
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

// ─── Timeline ─────────────────────────────────────────────────────────────────

func (r *Repository) GetCustomerTimeline(ctx context.Context, ownerID, customerID int64) ([]domain.TimelineEntry, error) {
	rows, err := r.queries.GetCustomerTimeline(ctx, sqlc.GetCustomerTimelineParams{
		OwnerID:    ownerID,
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

func (r *Repository) InsertTimelineEntry(ctx context.Context, entry domain.TimelineEntry) error {
	err := r.queries.InsertTimelineEntry(ctx, sqlc.InsertTimelineEntryParams{
		OwnerID:     entry.OwnerID,
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

func (r *Repository) BulkCreateCustomers(ctx context.Context, ownerID int64, items []domain.CustomerImportItem) (int, int, error) {
	imported := 0
	failed := 0

	for _, item := range items {
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

func (r *Repository) ExportCustomerData(ctx context.Context, ownerID, customerID int64) (domain.CustomerExport, error) {
	customer, err := r.GetCustomerByID(ctx, ownerID, customerID)
	if err != nil {
		return domain.CustomerExport{}, err
	}

	txs, _, err := r.ListLoyaltyTransactions(ctx, ownerID, customerID, 1, 100)
	if err != nil {
		return domain.CustomerExport{}, err
	}

	timeline, err := r.GetCustomerTimeline(ctx, ownerID, customerID)
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
	secondary, err := r.GetCustomerByID(ctx, input.OwnerID, input.SecondaryID)
	if err != nil {
		return err
	}

	primary, err := r.GetCustomerByID(ctx, input.OwnerID, input.PrimaryID)
	if err != nil {
		return err
	}

	newPoints := primary.LoyaltyPoints + secondary.LoyaltyPoints
	newSpent := primary.TotalSpent + secondary.TotalSpent
	newVisits := primary.VisitCount + secondary.VisitCount

	err = r.queries.UpdateCustomerMergeStats(ctx, sqlc.UpdateCustomerMergeStatsParams{
		LoyaltyPoints: sql.NullInt32{Int32: int32(newPoints), Valid: true},
		TotalSpent:    sql.NullString{String: strconv.FormatFloat(newSpent, 'f', 2, 64), Valid: true},
		VisitCount:    sql.NullInt32{Int32: int32(newVisits), Valid: true},
		ID:            input.PrimaryID,
		OwnerID:       input.OwnerID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update primary customer during merge", logger.Error(err))

		return err
	}

	err = r.queries.MigrateLoyaltyTransactions(ctx, sqlc.MigrateLoyaltyTransactionsParams{
		CustomerID:   input.PrimaryID,
		CustomerID_2: input.SecondaryID,
		OwnerID:      input.OwnerID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to migrate loyalty transactions during merge", logger.Error(err))

		return err
	}

	err = r.queries.MigrateTimeline(ctx, sqlc.MigrateTimelineParams{
		CustomerID:   input.PrimaryID,
		CustomerID_2: input.SecondaryID,
		OwnerID:      input.OwnerID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to migrate timeline during merge", logger.Error(err))

		return err
	}

	_, err = r.queries.DeleteCustomer(ctx, sqlc.DeleteCustomerParams{
		ID:      input.SecondaryID,
		OwnerID: input.OwnerID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete secondary customer during merge", logger.Error(err))

		return err
	}

	return nil
}

// ─── Loyalty Tiers ──────────────────────────────────────────────────────────

func (r *Repository) ListLoyaltyTiers(ctx context.Context, ownerID int64) ([]domain.LoyaltyTier, error) {
	rows, err := r.queries.ListLoyaltyTiers(ctx, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list loyalty tiers", logger.Error(err))

		return nil, err
	}

	tiers := make([]domain.LoyaltyTier, 0, len(rows))

	for _, row := range rows {
		dp, _ := strconv.ParseFloat(row.DiscountPct.String, 64)

		t := domain.LoyaltyTier{
			ID:          row.ID,
			OwnerID:     row.OwnerID,
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
		Name:        input.Name,
		MinPoints:   sql.NullInt32{Int32: int32(input.MinPoints), Valid: true},
		DiscountPct: sql.NullString{String: strconv.FormatFloat(input.DiscountPct, 'f', 2, 64), Valid: true},
		Perks:       json.RawMessage(input.Perks),
		SortOrder:   sql.NullInt32{Int32: int32(input.SortOrder), Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create loyalty tier", logger.Error(err))

		return 0, err
	}

	return result.LastInsertId()
}

func (r *Repository) UpdateLoyaltyTier(ctx context.Context, input domain.LoyaltyTierUpdateInput) error {
	result, err := r.queries.UpdateLoyaltyTier(ctx, sqlc.UpdateLoyaltyTierParams{
		Name:        input.Name,
		MinPoints:   sql.NullInt32{Int32: int32(input.MinPoints), Valid: true},
		DiscountPct: sql.NullString{String: strconv.FormatFloat(input.DiscountPct, 'f', 2, 64), Valid: true},
		Perks:       json.RawMessage(input.Perks),
		SortOrder:   sql.NullInt32{Int32: int32(input.SortOrder), Valid: true},
		ID:          input.ID,
		OwnerID:     input.OwnerID,
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

func (r *Repository) DeleteLoyaltyTier(ctx context.Context, ownerID, tierID int64) error {
	tier, err := r.queries.GetLoyaltyTierByID(ctx, sqlc.GetLoyaltyTierByIDParams{
		ID:      tierID,
		OwnerID: ownerID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrTierNotFound
		}

		logger.Ctx(ctx).Error("Failed to look up loyalty tier", logger.Error(err), logger.Any("owner_id", ownerID), logger.Any("tier_id", tierID))

		return err
	}

	// cross-schema read from crm_customers — active members on this tier
	count, err := r.queries.CountActiveMembersOnTier(ctx, sqlc.CountActiveMembersOnTierParams{
		OwnerID: ownerID,
		LOWER:   tier.Name,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to count active members on tier", logger.Error(err), logger.Any("owner_id", ownerID), logger.Any("tier_id", tierID))

		return err
	}

	if count > 0 {
		return domain.ErrTierInUse
	}

	result, err := r.queries.DeleteLoyaltyTier(ctx, sqlc.DeleteLoyaltyTierParams{
		ID:      tierID,
		OwnerID: ownerID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete loyalty tier", logger.Error(err), logger.Any("owner_id", ownerID), logger.Any("tier_id", tierID))

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

func (r *Repository) GetCampaignStats(ctx context.Context, ownerID, campaignID int64) (domain.CampaignStats, error) {
	campaign, err := r.GetCampaignByID(ctx, ownerID, campaignID)
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

	sendStats, err := r.queries.GetCampaignSendStats(ctx, campaignID)
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

func (r *Repository) GetSegmentCustomers(ctx context.Context, ownerID, segmentID int64, page, pageSize int) ([]domain.Customer, int, error) {
	if page < 1 {
		page = 1
	}

	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	_, err := r.GetSegmentByID(ctx, ownerID, segmentID)
	if err != nil {
		return nil, 0, err
	}

	total, err := r.queries.CountActiveCustomers(ctx, ownerID)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.queries.ListActiveCustomers(ctx, sqlc.ListActiveCustomersParams{
		OwnerID: ownerID,
		Limit:   int32(pageSize),
		Offset:  int32(offset),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get segment customers", logger.Error(err))

		return nil, 0, err
	}

	customers := make([]domain.Customer, 0, len(rows))

	for _, row := range rows {
		customers = append(customers, customerFromRow(row))
	}

	return customers, int(total), nil
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func customerFromRow(row sqlc.CrmCustomer) domain.Customer {
	c := domain.Customer{
		ID:            row.ID,
		OwnerID:       row.OwnerID,
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

	spent, _ := strconv.ParseFloat(row.TotalSpent.String, 64)
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
		_ = json.Unmarshal(row.Tags, &c.Tags)
	}

	if c.Tags == nil {
		c.Tags = []string{}
	}

	return c
}

func segmentFromRow(row sqlc.CrmSegment) domain.Segment {
	seg := domain.Segment{
		ID:            row.ID,
		OwnerID:       row.OwnerID,
		Name:          row.Name,
		Description:   row.Description.String,
		CustomerCount: int(row.CustomerCount.Int32),
		CreatedAt:     row.CreatedAt.Time,
		UpdatedAt:     row.UpdatedAt.Time,
	}

	if row.Rules != nil {
		seg.Rules = []byte(row.Rules)
	}

	if row.LastCalculatedAt.Valid {
		t := row.LastCalculatedAt.Time
		seg.LastCalculatedAt = &t
	}

	return seg
}

func campaignFromRow(row sqlc.CrmCampaign) domain.Campaign {
	c := domain.Campaign{
		ID:             row.ID,
		OwnerID:        row.OwnerID,
		Name:           row.Name,
		Description:    row.Description.String,
		Channel:        string(row.Channel),
		Status:         string(row.Status.CrmCampaignsStatus),
		Subject:        row.Subject.String,
		Content:        row.Content,
		AudienceCount:  int(row.AudienceCount.Int32),
		SentCount:      int(row.SentCount.Int32),
		DeliveredCount: int(row.DeliveredCount.Int32),
		OpenedCount:    int(row.OpenedCount.Int32),
		ClickedCount:   int(row.ClickedCount.Int32),
		CreatedAt:      row.CreatedAt.Time,
		UpdatedAt:      row.UpdatedAt.Time,
	}

	if row.SegmentID.Valid {
		id := row.SegmentID.Int64
		c.SegmentID = &id
	}

	if row.ScheduledAt.Valid {
		t := row.ScheduledAt.Time
		c.ScheduledAt = &t
	}

	if row.SentAt.Valid {
		t := row.SentAt.Time
		c.SentAt = &t
	}

	return c
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
	)

	if err := rows.Scan(
		&c.ID, &c.OwnerID, &c.FirstName, &lastName, &email, &phone, &dob, &gender,
		&address, &city, &country, &c.LoyaltyPoints, &c.TotalSpent, &c.VisitCount, &lastVisitAt,
		&c.LoyaltyTier, &tagsJSON, &notes, &c.IsActive, &c.CreatedAt, &c.UpdatedAt,
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

	if tagsJSON.Valid && tagsJSON.String != "" {
		_ = json.Unmarshal([]byte(tagsJSON.String), &c.Tags)
	}

	if c.Tags == nil {
		c.Tags = []string{}
	}

	return c, nil
}

func scanCampaignRaw(rows *sql.Rows) (domain.Campaign, error) {
	var c domain.Campaign
	var (
		desc        sql.NullString
		segmentID   sql.NullInt64
		subject     sql.NullString
		scheduledAt sql.NullTime
		sentAt      sql.NullTime
	)

	if err := rows.Scan(
		&c.ID, &c.OwnerID, &c.Name, &desc, &c.Channel, &c.Status, &segmentID, &subject,
		&c.Content, &scheduledAt, &sentAt, &c.AudienceCount, &c.SentCount, &c.DeliveredCount,
		&c.OpenedCount, &c.ClickedCount, &c.CreatedAt, &c.UpdatedAt,
	); err != nil {
		return domain.Campaign{}, err
	}

	c.Description = desc.String
	c.Subject = subject.String

	if segmentID.Valid {
		c.SegmentID = &segmentID.Int64
	}

	if scheduledAt.Valid {
		t := scheduledAt.Time
		c.ScheduledAt = &t
	}

	if sentAt.Valid {
		t := sentAt.Time
		c.SentAt = &t
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
