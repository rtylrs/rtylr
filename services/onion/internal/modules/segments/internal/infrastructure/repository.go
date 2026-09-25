package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/voxire/onion/internal/modules/segments/internal/domain"
	"github.com/voxire/onion/internal/modules/segments/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/segmentrules"
)

type Repository struct {
	queries *sqlc.Queries
	db      *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		queries: sqlc.New(db),
		db:      db,
	}
}

func (r *Repository) List(ctx context.Context, businessID int64) ([]domain.Segment, error) {
	rows, err := r.queries.ListSegments(ctx, businessID)
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

func (r *Repository) GetByID(ctx context.Context, businessID, id int64) (domain.Segment, error) {
	row, err := r.queries.GetSegmentByID(ctx, sqlc.GetSegmentByIDParams{
		ID:         id,
		BusinessID: businessID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Segment{}, domain.ErrSegmentNotFound
		}

		return domain.Segment{}, err
	}

	return segmentFromRow(row), nil
}

func (r *Repository) Create(ctx context.Context, input domain.SegmentInput) (int64, error) {
	filters := input.Filters
	if filters == nil {
		filters = []byte("{}")
	}

	result, err := r.queries.CreateSegment(ctx, sqlc.CreateSegmentParams{
		OwnerID:     input.OwnerID,
		BusinessID:  input.BusinessID,
		Name:        input.Name,
		Description: sql.NullString{String: input.Description, Valid: input.Description != ""},
		Rules:       json.RawMessage(filters),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create segment", logger.Error(err))

		return 0, err
	}

	return result.LastInsertId()
}

func (r *Repository) Update(ctx context.Context, input domain.SegmentUpdateInput) error {
	businessID := input.BusinessID
	if businessID == 0 {
		businessID = input.OwnerID
	}

	filters := input.Filters
	if filters == nil {
		filters = []byte("{}")
	}

	result, err := r.queries.UpdateSegment(ctx, sqlc.UpdateSegmentParams{
		Name:        input.Name,
		Description: sql.NullString{String: input.Description, Valid: input.Description != ""},
		Rules:       json.RawMessage(filters),
		ID:          input.ID,
		BusinessID:  businessID,
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

// ListCustomers returns a paginated view of the business's active customers,
// filtered by the segment's stored rules when they are present and valid.
// Segments with no rules keep the legacy "all active customers" behavior.
func (r *Repository) ListCustomers(ctx context.Context, businessID int64, filters []byte, page, pageSize int) ([]domain.SegmentCustomer, int, error) {
	if page < 1 {
		page = 1
	}

	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	where := "business_id = ? AND is_active = TRUE"
	args := []any{businessID}
	if frag, fargs := segmentrules.Filter(filters); frag != "" {
		where += " AND " + frag
		args = append(args, fargs...)
	}

	// nosemgrep: go.lang.security.audit.database.string-formatted-query.string-formatted-query
	countSQL := "SELECT COUNT(*) FROM crm_customers WHERE " + where
	var total int
	if err := r.db.QueryRowContext(ctx, countSQL, args...).Scan(&total); err != nil {
		logger.Ctx(ctx).Error("Failed to count segment customers", logger.Error(err))

		return nil, 0, err
	}

	listSQL := `SELECT id, owner_id, first_name, last_name, email, phone, date_of_birth, gender,
       address, city, country, loyalty_points, total_spent, visit_count, last_visit_at,
       loyalty_tier, tags, notes, is_active, created_at, updated_at
FROM crm_customers WHERE ` + where + `
ORDER BY created_at DESC LIMIT ? OFFSET ?`
	rows, err := r.db.QueryContext(ctx, listSQL, append(args, pageSize, offset)...)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list segment customers", logger.Error(err))

		return nil, 0, err
	}
	defer rows.Close()

	customers := make([]domain.SegmentCustomer, 0, pageSize)

	for rows.Next() {
		var (
			c           domain.SegmentCustomer
			lastName    sql.NullString
			email       sql.NullString
			phone       sql.NullString
			dob         sql.NullTime
			gender      sql.NullString
			address     sql.NullString
			city        sql.NullString
			country     sql.NullString
			loyalPoints sql.NullInt64
			totalSpent  sql.NullString
			visitCount  sql.NullInt64
			lastVisitAt sql.NullTime
			tier        sql.NullString
			tagsJSON    sql.NullString
			notes       sql.NullString
		)
		if err := rows.Scan(&c.ID, &c.OwnerID, &c.FirstName, &lastName, &email, &phone, &dob, &gender,
			&address, &city, &country, &loyalPoints, &totalSpent, &visitCount, &lastVisitAt,
			&tier, &tagsJSON, &notes, &c.IsActive, &c.CreatedAt, &c.UpdatedAt); err != nil {
			logger.Ctx(ctx).Error("Failed to scan segment customer", logger.Error(err))

			return nil, 0, err
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
		if loyalPoints.Valid {
			c.LoyaltyPoints = loyalPoints.Int64
		}
		if totalSpent.Valid {
			c.TotalSpent, _ = strconv.ParseFloat(totalSpent.String, 64)
		}
		if visitCount.Valid {
			c.VisitCount = visitCount.Int64
		}
		if lastVisitAt.Valid {
			t := lastVisitAt.Time
			c.LastVisitAt = &t
		}
		c.LoyaltyTier = tier.String
		if tagsJSON.Valid && tagsJSON.String != "" {
			var tags []string
			if err := json.Unmarshal([]byte(tagsJSON.String), &tags); err == nil {
				c.Tags = tags
			}
		}
		if c.Tags == nil {
			c.Tags = []string{}
		}
		c.Notes = notes.String

		customers = append(customers, c)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return customers, total, nil
}

func (r *Repository) Delete(ctx context.Context, businessID, id int64) error {
	_, err := r.queries.DeleteSegment(ctx, sqlc.DeleteSegmentParams{
		ID:         id,
		BusinessID: businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete segment", logger.Error(err))

		return err
	}

	return nil
}

func segmentFromRow(row sqlc.CrmSegment) domain.Segment {
	seg := domain.Segment{
		ID:            row.ID,
		OwnerID:       row.OwnerID,
		BusinessID:    row.BusinessID,
		Name:          row.Name,
		Description:   row.Description.String,
		CustomerCount: int(row.CustomerCount.Int32),
		CreatedAt:     row.CreatedAt.Time,
		UpdatedAt:     row.UpdatedAt.Time,
	}

	if row.Rules != nil {
		seg.Filters = []byte(row.Rules)
	}
	if row.LastCalculatedAt.Valid {
		t := row.LastCalculatedAt.Time
		seg.LastCalculatedAt = &t
	}

	return seg
}
