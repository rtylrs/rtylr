package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"math"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/customers/internal/domain"
	"github.com/voxire/onion/internal/modules/customers/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

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
