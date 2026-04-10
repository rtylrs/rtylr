package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/customers/internal/domain"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	pool *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	return &Repository{
		pool: dbclient.Pool(),
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

	var (
		whereClauses []string
		args         []interface{}
	)

	whereClauses = append(whereClauses, "owner_id = ?")
	args = append(args, filter.OwnerID)

	if filter.Search != "" {
		whereClauses = append(whereClauses, "(first_name LIKE ? OR last_name LIKE ? OR email LIKE ? OR phone LIKE ?)")
		like := "%" + filter.Search + "%"
		args = append(args, like, like, like, like)
	}

	if filter.Tier != "" {
		whereClauses = append(whereClauses, "loyalty_tier = ?")
		args = append(args, filter.Tier)
	}

	if filter.Tag != "" {
		whereClauses = append(whereClauses, "JSON_CONTAINS(tags, ?)")
		tagJSON, _ := json.Marshal(filter.Tag)
		args = append(args, string(tagJSON))
	}

	where := strings.Join(whereClauses, " AND ")

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM crm_customers WHERE %s", where)
	var total int

	if err := r.pool.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		logger.Ctx(ctx).Error("Failed to count customers", logger.Error(err))

		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT id, owner_id, first_name, last_name, email, phone, date_of_birth, gender,
		       address, city, country, loyalty_points, total_spent, visit_count, last_visit_at,
		       loyalty_tier, tags, notes, is_active, created_at, updated_at
		FROM crm_customers
		WHERE %s
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`, where)

	args = append(args, filter.PageSize, offset)

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list customers", logger.Error(err))

		return nil, 0, err
	}
	defer rows.Close()

	customers := make([]domain.Customer, 0)

	for rows.Next() {
		c, scanErr := scanCustomer(rows)
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
	query := `
		SELECT id, owner_id, first_name, last_name, email, phone, date_of_birth, gender,
		       address, city, country, loyalty_points, total_spent, visit_count, last_visit_at,
		       loyalty_tier, tags, notes, is_active, created_at, updated_at
		FROM crm_customers
		WHERE id = ? AND owner_id = ?`

	row := r.pool.QueryRowContext(ctx, query, customerID, ownerID)

	c, err := scanCustomerRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Customer{}, domain.ErrCustomerNotFound
		}

		logger.Ctx(ctx).Error("Failed to get customer", logger.Error(err))

		return domain.Customer{}, err
	}

	return c, nil
}

func (r *Repository) CreateCustomer(ctx context.Context, input domain.CustomerInput) (int64, error) {
	tagsJSON, err := json.Marshal(input.Tags)
	if err != nil {
		tagsJSON = []byte("[]")
	}

	query := `
		INSERT INTO crm_customers
		  (owner_id, first_name, last_name, email, phone, date_of_birth, gender,
		   address, city, country, tags, notes, loyalty_tier, is_active)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'bronze', TRUE)`

	result, err := r.pool.ExecContext(ctx, query,
		input.OwnerID,
		input.FirstName,
		nullableString(input.LastName),
		nullableString(input.Email),
		nullableString(input.Phone),
		nullableTime(input.DateOfBirth),
		nullableString(input.Gender),
		nullableString(input.Address),
		nullableString(input.City),
		nullableString(input.Country),
		tagsJSON,
		nullableString(input.Notes),
	)
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

	query := `
		UPDATE crm_customers
		SET first_name = ?, last_name = ?, email = ?, phone = ?, date_of_birth = ?,
		    gender = ?, address = ?, city = ?, country = ?, tags = ?, notes = ?, is_active = ?
		WHERE id = ? AND owner_id = ?`

	result, err := r.pool.ExecContext(ctx, query,
		input.FirstName,
		nullableString(input.LastName),
		nullableString(input.Email),
		nullableString(input.Phone),
		nullableTime(input.DateOfBirth),
		nullableString(input.Gender),
		nullableString(input.Address),
		nullableString(input.City),
		nullableString(input.Country),
		tagsJSON,
		nullableString(input.Notes),
		isActive,
		input.ID,
		input.OwnerID,
	)
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
	result, err := r.pool.ExecContext(ctx,
		"DELETE FROM crm_customers WHERE id = ? AND owner_id = ?",
		customerID, ownerID,
	)
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
	_, err := r.pool.ExecContext(ctx,
		"UPDATE crm_customers SET loyalty_points = ?, loyalty_tier = ? WHERE id = ?",
		newPoints, newTier, customerID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update customer loyalty", logger.Error(err))
	}

	return err
}

func (r *Repository) RecordVisit(ctx context.Context, ownerID, customerID int64, amount float64) error {
	_, err := r.pool.ExecContext(ctx,
		`UPDATE crm_customers
		 SET visit_count = visit_count + 1,
		     last_visit_at = NOW(),
		     total_spent = total_spent + ?
		 WHERE id = ? AND owner_id = ?`,
		amount, customerID, ownerID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to record visit", logger.Error(err))
	}

	return err
}

// ─── Loyalty Transactions ─────────────────────────────────────────────────────

func (r *Repository) CreateLoyaltyTransaction(ctx context.Context, tx domain.LoyaltyTransaction) error {
	_, err := r.pool.ExecContext(ctx,
		`INSERT INTO crm_loyalty_transactions
		   (owner_id, customer_id, type, points, balance_before, balance_after,
		    description, reference_id, reference_type)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		tx.OwnerID,
		tx.CustomerID,
		tx.Type,
		tx.Points,
		tx.BalanceBefore,
		tx.BalanceAfter,
		nullableString(tx.Description),
		tx.ReferenceID,
		nullableString(tx.ReferenceType),
	)
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

	var total int
	if err := r.pool.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM crm_loyalty_transactions WHERE owner_id = ? AND customer_id = ?",
		ownerID, customerID,
	).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.pool.QueryContext(ctx,
		`SELECT id, owner_id, customer_id, type, points, balance_before, balance_after,
		        description, reference_id, reference_type, created_at
		 FROM crm_loyalty_transactions
		 WHERE owner_id = ? AND customer_id = ?
		 ORDER BY created_at DESC
		 LIMIT ? OFFSET ?`,
		ownerID, customerID, pageSize, offset,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list loyalty transactions", logger.Error(err))

		return nil, 0, err
	}
	defer rows.Close()

	txs := make([]domain.LoyaltyTransaction, 0)

	for rows.Next() {
		var tx domain.LoyaltyTransaction
		var desc sql.NullString
		var refID sql.NullInt64
		var refType sql.NullString

		if scanErr := rows.Scan(
			&tx.ID, &tx.OwnerID, &tx.CustomerID, &tx.Type, &tx.Points,
			&tx.BalanceBefore, &tx.BalanceAfter, &desc, &refID, &refType, &tx.CreatedAt,
		); scanErr != nil {
			logger.Ctx(ctx).Error("Failed to scan loyalty transaction", logger.Error(scanErr))

			return nil, 0, scanErr
		}

		tx.Description = desc.String
		if refID.Valid {
			tx.ReferenceID = &refID.Int64
		}
		tx.ReferenceType = refType.String
		txs = append(txs, tx)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return txs, total, nil
}

// ─── Loyalty Config ───────────────────────────────────────────────────────────

func (r *Repository) GetLoyaltyConfig(ctx context.Context, ownerID int64) (domain.LoyaltyConfig, error) {
	var cfg domain.LoyaltyConfig

	err := r.pool.QueryRowContext(ctx,
		`SELECT id, owner_id, is_enabled, points_per_currency_unit, currency_per_point,
		        bronze_min, silver_min, gold_min, platinum_min,
		        bronze_discount_pct, silver_discount_pct, gold_discount_pct, platinum_discount_pct,
		        points_expiry_days, updated_at
		 FROM crm_loyalty_config WHERE owner_id = ?`, ownerID,
	).Scan(
		&cfg.ID, &cfg.OwnerID, &cfg.IsEnabled,
		&cfg.PointsPerCurrencyUnit, &cfg.CurrencyPerPoint,
		&cfg.BronzeMin, &cfg.SilverMin, &cfg.GoldMin, &cfg.PlatinumMin,
		&cfg.BronzeDiscountPct, &cfg.SilverDiscountPct, &cfg.GoldDiscountPct, &cfg.PlatinumDiscountPct,
		&cfg.PointsExpiryDays, &cfg.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.LoyaltyConfig{}, nil
		}

		logger.Ctx(ctx).Error("Failed to get loyalty config", logger.Error(err))

		return domain.LoyaltyConfig{}, err
	}

	return cfg, nil
}

func (r *Repository) UpsertLoyaltyConfig(ctx context.Context, input domain.LoyaltyConfigInput) error {
	_, err := r.pool.ExecContext(ctx,
		`INSERT INTO crm_loyalty_config
		   (owner_id, is_enabled, points_per_currency_unit, currency_per_point,
		    bronze_min, silver_min, gold_min, platinum_min,
		    bronze_discount_pct, silver_discount_pct, gold_discount_pct, platinum_discount_pct,
		    points_expiry_days)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		 ON DUPLICATE KEY UPDATE
		   is_enabled = VALUES(is_enabled),
		   points_per_currency_unit = VALUES(points_per_currency_unit),
		   currency_per_point = VALUES(currency_per_point),
		   bronze_min = VALUES(bronze_min),
		   silver_min = VALUES(silver_min),
		   gold_min = VALUES(gold_min),
		   platinum_min = VALUES(platinum_min),
		   bronze_discount_pct = VALUES(bronze_discount_pct),
		   silver_discount_pct = VALUES(silver_discount_pct),
		   gold_discount_pct = VALUES(gold_discount_pct),
		   platinum_discount_pct = VALUES(platinum_discount_pct),
		   points_expiry_days = VALUES(points_expiry_days)`,
		input.OwnerID, input.IsEnabled,
		input.PointsPerCurrencyUnit, input.CurrencyPerPoint,
		input.BronzeMin, input.SilverMin, input.GoldMin, input.PlatinumMin,
		input.BronzeDiscountPct, input.SilverDiscountPct, input.GoldDiscountPct, input.PlatinumDiscountPct,
		input.PointsExpiryDays,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to upsert loyalty config", logger.Error(err))
	}

	return err
}

// ─── Segments ─────────────────────────────────────────────────────────────────

func (r *Repository) ListSegments(ctx context.Context, ownerID int64) ([]domain.Segment, error) {
	rows, err := r.pool.QueryContext(ctx,
		`SELECT id, owner_id, name, description, rules, customer_count, last_calculated_at, created_at, updated_at
		 FROM crm_segments WHERE owner_id = ? ORDER BY created_at DESC`, ownerID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list segments", logger.Error(err))

		return nil, err
	}
	defer rows.Close()

	segments := make([]domain.Segment, 0)

	for rows.Next() {
		seg, scanErr := scanSegment(rows)
		if scanErr != nil {
			return nil, scanErr
		}

		segments = append(segments, seg)
	}

	return segments, rows.Err()
}

func (r *Repository) GetSegmentByID(ctx context.Context, ownerID, segmentID int64) (domain.Segment, error) {
	row := r.pool.QueryRowContext(ctx,
		`SELECT id, owner_id, name, description, rules, customer_count, last_calculated_at, created_at, updated_at
		 FROM crm_segments WHERE id = ? AND owner_id = ?`, segmentID, ownerID,
	)

	seg, err := scanSegmentRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Segment{}, domain.ErrSegmentNotFound
		}

		return domain.Segment{}, err
	}

	return seg, nil
}

func (r *Repository) CreateSegment(ctx context.Context, input domain.SegmentInput) (int64, error) {
	result, err := r.pool.ExecContext(ctx,
		`INSERT INTO crm_segments (owner_id, name, description, rules) VALUES (?, ?, ?, ?)`,
		input.OwnerID,
		input.Name,
		nullableString(input.Description),
		input.Rules,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create segment", logger.Error(err))

		return 0, err
	}

	return result.LastInsertId()
}

func (r *Repository) UpdateSegment(ctx context.Context, input domain.SegmentUpdateInput) error {
	result, err := r.pool.ExecContext(ctx,
		`UPDATE crm_segments SET name = ?, description = ?, rules = ? WHERE id = ? AND owner_id = ?`,
		input.Name,
		nullableString(input.Description),
		input.Rules,
		input.ID,
		input.OwnerID,
	)
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
	result, err := r.pool.ExecContext(ctx,
		"DELETE FROM crm_segments WHERE id = ? AND owner_id = ?",
		segmentID, ownerID,
	)
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
	var (
		whereClauses []string
		args         []interface{}
	)

	whereClauses = append(whereClauses, "owner_id = ?")
	args = append(args, ownerID)

	if status != "" {
		whereClauses = append(whereClauses, "status = ?")
		args = append(args, status)
	}

	where := strings.Join(whereClauses, " AND ")

	rows, err := r.pool.QueryContext(ctx,
		fmt.Sprintf(`SELECT id, owner_id, name, description, channel, status, segment_id, subject,
		        content, scheduled_at, sent_at, audience_count, sent_count, delivered_count,
		        opened_count, clicked_count, created_at, updated_at
		 FROM crm_campaigns WHERE %s ORDER BY created_at DESC`, where),
		args...,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list campaigns", logger.Error(err))

		return nil, err
	}
	defer rows.Close()

	campaigns := make([]domain.Campaign, 0)

	for rows.Next() {
		c, scanErr := scanCampaign(rows)
		if scanErr != nil {
			return nil, scanErr
		}

		campaigns = append(campaigns, c)
	}

	return campaigns, rows.Err()
}

func (r *Repository) GetCampaignByID(ctx context.Context, ownerID, campaignID int64) (domain.Campaign, error) {
	row := r.pool.QueryRowContext(ctx,
		`SELECT id, owner_id, name, description, channel, status, segment_id, subject,
		        content, scheduled_at, sent_at, audience_count, sent_count, delivered_count,
		        opened_count, clicked_count, created_at, updated_at
		 FROM crm_campaigns WHERE id = ? AND owner_id = ?`, campaignID, ownerID,
	)

	c, err := scanCampaignRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Campaign{}, domain.ErrCampaignNotFound
		}

		return domain.Campaign{}, err
	}

	return c, nil
}

func (r *Repository) CreateCampaign(ctx context.Context, input domain.CampaignInput) (int64, error) {
	result, err := r.pool.ExecContext(ctx,
		`INSERT INTO crm_campaigns
		   (owner_id, name, description, channel, segment_id, subject, content, scheduled_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		input.OwnerID,
		input.Name,
		nullableString(input.Description),
		input.Channel,
		input.SegmentID,
		nullableString(input.Subject),
		input.Content,
		nullableTime(input.ScheduledAt),
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create campaign", logger.Error(err))

		return 0, err
	}

	return result.LastInsertId()
}

func (r *Repository) UpdateCampaign(ctx context.Context, input domain.CampaignUpdateInput) error {
	result, err := r.pool.ExecContext(ctx,
		`UPDATE crm_campaigns
		 SET name = ?, description = ?, channel = ?, segment_id = ?, subject = ?,
		     content = ?, scheduled_at = ?
		 WHERE id = ? AND owner_id = ?`,
		input.Name,
		nullableString(input.Description),
		input.Channel,
		input.SegmentID,
		nullableString(input.Subject),
		input.Content,
		nullableTime(input.ScheduledAt),
		input.ID,
		input.OwnerID,
	)
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
	result, err := r.pool.ExecContext(ctx,
		"UPDATE crm_campaigns SET status = ? WHERE id = ? AND owner_id = ?",
		status, campaignID, ownerID,
	)
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

	if err := r.pool.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM crm_customers WHERE owner_id = ? AND is_active = TRUE", ownerID,
	).Scan(&stats.TotalCustomers); err != nil {
		logger.Ctx(ctx).Error("Failed to get total customers", logger.Error(err))

		return domain.DashboardStats{}, err
	}

	if err := r.pool.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM crm_customers
		 WHERE owner_id = ? AND MONTH(created_at) = MONTH(NOW()) AND YEAR(created_at) = YEAR(NOW())`,
		ownerID,
	).Scan(&stats.NewThisMonth); err != nil {
		logger.Ctx(ctx).Error("Failed to get new customers this month", logger.Error(err))

		return domain.DashboardStats{}, err
	}

	if err := r.pool.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM crm_customers WHERE owner_id = ? AND loyalty_points > 0 AND is_active = TRUE",
		ownerID,
	).Scan(&stats.ActiveLoyaltyMembers); err != nil {
		logger.Ctx(ctx).Error("Failed to get active loyalty members", logger.Error(err))

		return domain.DashboardStats{}, err
	}

	rows, err := r.pool.QueryContext(ctx,
		`SELECT id, first_name, last_name, email, total_spent, loyalty_points, loyalty_tier
		 FROM crm_customers WHERE owner_id = ? AND is_active = TRUE
		 ORDER BY total_spent DESC LIMIT 10`, ownerID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get top customers", logger.Error(err))

		return domain.DashboardStats{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var tc domain.TopCustomer
		var lastName, email sql.NullString

		if scanErr := rows.Scan(&tc.ID, &tc.FirstName, &lastName, &email, &tc.TotalSpent, &tc.LoyaltyPoints, &tc.LoyaltyTier); scanErr != nil {
			return domain.DashboardStats{}, scanErr
		}

		tc.LastName = lastName.String
		tc.Email = email.String
		stats.TopCustomers = append(stats.TopCustomers, tc)
	}

	if err := rows.Err(); err != nil {
		return domain.DashboardStats{}, err
	}

	if stats.TopCustomers == nil {
		stats.TopCustomers = []domain.TopCustomer{}
	}

	return stats, nil
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

type customerScanner interface {
	Scan(dest ...interface{}) error
}

func scanCustomer(s *sql.Rows) (domain.Customer, error) {
	return scanCustomerColumns(s)
}

func scanCustomerRow(s *sql.Row) (domain.Customer, error) {
	return scanCustomerColumns(s)
}

func scanCustomerColumns(s customerScanner) (domain.Customer, error) {
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

	if err := s.Scan(
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

type segmentScanner interface {
	Scan(dest ...interface{}) error
}

func scanSegment(s *sql.Rows) (domain.Segment, error) {
	return scanSegmentColumns(s)
}

func scanSegmentRow(s *sql.Row) (domain.Segment, error) {
	return scanSegmentColumns(s)
}

func scanSegmentColumns(s segmentScanner) (domain.Segment, error) {
	var seg domain.Segment
	var desc sql.NullString
	var lastCalc sql.NullTime

	if err := s.Scan(
		&seg.ID, &seg.OwnerID, &seg.Name, &desc, &seg.Rules, &seg.CustomerCount, &lastCalc,
		&seg.CreatedAt, &seg.UpdatedAt,
	); err != nil {
		return domain.Segment{}, err
	}

	seg.Description = desc.String

	if lastCalc.Valid {
		t := lastCalc.Time
		seg.LastCalculatedAt = &t
	}

	return seg, nil
}

type campaignScanner interface {
	Scan(dest ...interface{}) error
}

func scanCampaign(s *sql.Rows) (domain.Campaign, error) {
	return scanCampaignColumns(s)
}

func scanCampaignRow(s *sql.Row) (domain.Campaign, error) {
	return scanCampaignColumns(s)
}

func scanCampaignColumns(s campaignScanner) (domain.Campaign, error) {
	var c domain.Campaign
	var (
		desc        sql.NullString
		segmentID   sql.NullInt64
		subject     sql.NullString
		scheduledAt sql.NullTime
		sentAt      sql.NullTime
	)

	if err := s.Scan(
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

func nullableString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

func nullableTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{}
	}

	return sql.NullTime{Time: *t, Valid: true}
}

// -- Timeline ----------------------------------------------------------------

func (r *Repository) GetCustomerTimeline(ctx context.Context, ownerID, customerID int64) ([]domain.TimelineEntry, error) {
	rows, err := r.pool.QueryContext(ctx,
		`SELECT id, owner_id, customer_id, type, description, metadata, created_at
		 FROM crm_customer_timeline
		 WHERE owner_id = ? AND customer_id = ?
		 ORDER BY created_at DESC
		 LIMIT 100`,
		ownerID, customerID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get customer timeline", logger.Error(err))

		return nil, err
	}
	defer rows.Close()

	entries := make([]domain.TimelineEntry, 0)

	for rows.Next() {
		var e domain.TimelineEntry
		var metadata sql.NullString

		if scanErr := rows.Scan(&e.ID, &e.OwnerID, &e.CustomerID, &e.Type, &e.Description, &metadata, &e.CreatedAt); scanErr != nil {
			logger.Ctx(ctx).Error("Failed to scan timeline entry", logger.Error(scanErr))

			return nil, scanErr
		}

		if metadata.Valid {
			e.Metadata = []byte(metadata.String)
		} else {
			e.Metadata = []byte("{}")
		}

		entries = append(entries, e)
	}

	return entries, rows.Err()
}

func (r *Repository) InsertTimelineEntry(ctx context.Context, entry domain.TimelineEntry) error {
	_, err := r.pool.ExecContext(ctx,
		`INSERT INTO crm_customer_timeline (owner_id, customer_id, type, description, metadata)
		 VALUES (?, ?, ?, ?, ?)`,
		entry.OwnerID, entry.CustomerID, entry.Type, entry.Description, string(entry.Metadata),
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to insert timeline entry", logger.Error(err))
	}

	return err
}

// -- Import ------------------------------------------------------------------

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

		_, err = r.pool.ExecContext(ctx,
			`INSERT INTO crm_customers
			  (owner_id, first_name, last_name, email, phone, date_of_birth, gender,
			   address, city, country, tags, notes, loyalty_tier, is_active)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'bronze', TRUE)`,
			ownerID,
			item.FirstName,
			nullableString(item.LastName),
			nullableString(item.Email),
			nullableString(item.Phone),
			nullableTime(dob),
			nullableString(item.Gender),
			nullableString(item.Address),
			nullableString(item.City),
			nullableString(item.Country),
			tagsJSON,
			nullableString(item.Notes),
		)
		if err != nil {
			logger.Ctx(ctx).Error("Failed to import customer", logger.Error(err))
			failed++
		} else {
			imported++
		}
	}

	return imported, failed, nil
}

// -- Export (GDPR) -----------------------------------------------------------

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

// -- Merge -------------------------------------------------------------------

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

	_, err = r.pool.ExecContext(ctx,
		`UPDATE crm_customers
		 SET loyalty_points = ?, total_spent = ?, visit_count = ?
		 WHERE id = ? AND owner_id = ?`,
		newPoints, newSpent, newVisits, input.PrimaryID, input.OwnerID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update primary customer during merge", logger.Error(err))

		return err
	}

	_, err = r.pool.ExecContext(ctx,
		`UPDATE crm_loyalty_transactions SET customer_id = ? WHERE customer_id = ? AND owner_id = ?`,
		input.PrimaryID, input.SecondaryID, input.OwnerID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to migrate loyalty transactions during merge", logger.Error(err))

		return err
	}

	_, err = r.pool.ExecContext(ctx,
		`UPDATE crm_customer_timeline SET customer_id = ? WHERE customer_id = ? AND owner_id = ?`,
		input.PrimaryID, input.SecondaryID, input.OwnerID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to migrate timeline during merge", logger.Error(err))

		return err
	}

	_, err = r.pool.ExecContext(ctx,
		"DELETE FROM crm_customers WHERE id = ? AND owner_id = ?",
		input.SecondaryID, input.OwnerID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete secondary customer during merge", logger.Error(err))

		return err
	}

	return nil
}

// -- Loyalty Tiers -----------------------------------------------------------

func (r *Repository) ListLoyaltyTiers(ctx context.Context, ownerID int64) ([]domain.LoyaltyTier, error) {
	rows, err := r.pool.QueryContext(ctx,
		`SELECT id, owner_id, name, min_points, discount_pct, perks, sort_order, created_at, updated_at
		 FROM crm_loyalty_tiers WHERE owner_id = ? ORDER BY sort_order ASC`, ownerID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list loyalty tiers", logger.Error(err))

		return nil, err
	}
	defer rows.Close()

	tiers := make([]domain.LoyaltyTier, 0)

	for rows.Next() {
		var t domain.LoyaltyTier
		var perks sql.NullString

		if scanErr := rows.Scan(&t.ID, &t.OwnerID, &t.Name, &t.MinPoints, &t.DiscountPct, &perks, &t.SortOrder, &t.CreatedAt, &t.UpdatedAt); scanErr != nil {
			return nil, scanErr
		}

		if perks.Valid {
			t.Perks = []byte(perks.String)
		} else {
			t.Perks = []byte("[]")
		}

		tiers = append(tiers, t)
	}

	return tiers, rows.Err()
}

func (r *Repository) CreateLoyaltyTier(ctx context.Context, input domain.LoyaltyTierInput) (int64, error) {
	result, err := r.pool.ExecContext(ctx,
		`INSERT INTO crm_loyalty_tiers (owner_id, name, min_points, discount_pct, perks, sort_order)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		input.OwnerID, input.Name, input.MinPoints, input.DiscountPct, string(input.Perks), input.SortOrder,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create loyalty tier", logger.Error(err))

		return 0, err
	}

	return result.LastInsertId()
}

func (r *Repository) UpdateLoyaltyTier(ctx context.Context, input domain.LoyaltyTierUpdateInput) error {
	result, err := r.pool.ExecContext(ctx,
		`UPDATE crm_loyalty_tiers SET name = ?, min_points = ?, discount_pct = ?, perks = ?, sort_order = ?
		 WHERE id = ? AND owner_id = ?`,
		input.Name, input.MinPoints, input.DiscountPct, string(input.Perks), input.SortOrder,
		input.ID, input.OwnerID,
	)
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

// -- Campaign Stats ----------------------------------------------------------

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

	var bounced, failed int

	err = r.pool.QueryRowContext(ctx,
		`SELECT
			COALESCE(SUM(CASE WHEN status = 'bounced' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END), 0)
		 FROM crm_campaign_sends WHERE campaign_id = ?`,
		campaignID,
	).Scan(&bounced, &failed)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get campaign send stats", logger.Error(err))
	}

	stats.BouncedCount = bounced
	stats.FailedCount = failed

	return stats, nil
}

// -- Segment Customers -------------------------------------------------------

func (r *Repository) GetSegmentCustomers(ctx context.Context, ownerID, segmentID int64, page, pageSize int) ([]domain.Customer, int, error) {
	if page < 1 {
		page = 1
	}

	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	seg, err := r.GetSegmentByID(ctx, ownerID, segmentID)
	if err != nil {
		return nil, 0, err
	}

	_ = seg

	var total int

	if err := r.pool.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM crm_customers WHERE owner_id = ? AND is_active = TRUE",
		ownerID,
	).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, owner_id, first_name, last_name, email, phone, date_of_birth, gender,
		       address, city, country, loyalty_points, total_spent, visit_count, last_visit_at,
		       loyalty_tier, tags, notes, is_active, created_at, updated_at
		FROM crm_customers
		WHERE owner_id = ? AND is_active = TRUE
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`

	rows, err := r.pool.QueryContext(ctx, query, ownerID, pageSize, offset)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get segment customers", logger.Error(err))

		return nil, 0, err
	}
	defer rows.Close()

	customers := make([]domain.Customer, 0)

	for rows.Next() {
		c, scanErr := scanCustomer(rows)
		if scanErr != nil {
			return nil, 0, scanErr
		}

		customers = append(customers, c)
	}

	return customers, total, rows.Err()
}
