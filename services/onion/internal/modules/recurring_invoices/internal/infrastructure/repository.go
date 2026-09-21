package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/recurring_invoices/internal/domain"
	"github.com/voxire/onion/internal/modules/recurring_invoices/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	pool    *sql.DB
	queries *sqlc.Queries
}

type dueSchedule struct {
	id         int64
	businessID int64
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		queries: sqlc.New(db),
		pool:    db,
	}
}

func (r *Repository) List(ctx context.Context, businessID int64) ([]domain.RecurringInvoice, error) {
	rows, err := r.queries.ListRecurringInvoices(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list recurring invoices", logger.Error(err))
		return nil, err
	}

	items := make([]domain.RecurringInvoice, 0, len(rows))
	for _, row := range rows {
		ri := domain.RecurringInvoice{
			ID:                row.ID,
			OwnerID:           row.OwnerID,
			BusinessID:        row.BusinessID,
			TemplateInvoiceID: row.TemplateInvoiceID,
			Frequency:         string(row.Frequency),
			NextInvoiceDate:   row.NextInvoiceDate,
			IsActive:          nullBool(row.IsActive),
			CreatedAt:         nullTime(row.CreatedAt),
		}
		if row.EndDate.Valid {
			ri.EndDate = row.EndDate.Time
		}
		items = append(items, ri)
	}

	return items, nil
}

func (r *Repository) Create(ctx context.Context, businessID int64, input domain.RecurringInvoiceInput) (int64, error) {
	if input.TemplateInvoiceID == nil {
		return 0, errors.New("template_invoice_id is required")
	}
	nextDate, err := time.Parse("2006-01-02", input.NextInvoiceDate)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to parse next_invoice_date", logger.Error(err))
		return 0, err
	}
	endDate, err := parseEndDate(input.EndDate)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to parse end_date", logger.Error(err))
		return 0, err
	}

	// is_active is optional in the contract and defaults to TRUE in the schema:
	// an omitted flag means "active unless told otherwise".
	isActive := true
	if input.IsActive != nil {
		isActive = *input.IsActive
	}

	res, err := r.queries.CreateRecurringInvoice(ctx, sqlc.CreateRecurringInvoiceParams{
		OwnerID:           businessID,
		BusinessID:        businessID,
		TemplateInvoiceID: *input.TemplateInvoiceID,
		Frequency:         sqlc.FinanceRecurringInvoicesFrequency(input.Frequency),
		NextInvoiceDate:   nextDate,
		EndDate:           endDate,
		IsActive:          sql.NullBool{Bool: isActive, Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create recurring invoice", logger.Error(err))
		return 0, err
	}

	return res.LastInsertId()
}

func (r *Repository) Update(ctx context.Context, businessID, id int64, input domain.RecurringInvoiceInput) error {
	nextDate, err := time.Parse("2006-01-02", input.NextInvoiceDate)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to parse next_invoice_date", logger.Error(err))
		return err
	}
	endDate, err := parseEndDate(input.EndDate)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to parse end_date", logger.Error(err))
		return err
	}

	// The optional fields only overwrite the stored row when the caller sent
	// them: an omitted template_invoice_id must not clear the template, an
	// omitted is_active must not pause the schedule, and an omitted end_date
	// must not turn a bounded schedule into an open-ended one.
	setEndDate := int64(0)
	if input.EndDate != nil {
		setEndDate = 1
	}
	res, err := r.queries.UpdateRecurringInvoice(ctx, sqlc.UpdateRecurringInvoiceParams{
		TemplateInvoiceID: nullInt64Ptr(input.TemplateInvoiceID),
		Frequency:         sqlc.FinanceRecurringInvoicesFrequency(input.Frequency),
		NextInvoiceDate:   nextDate,
		SetEndDate:        setEndDate,
		EndDate:           endDate,
		IsActive:          nullBoolPtr(input.IsActive),
		ID:                id,
		BusinessID:        businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update recurring invoice", logger.Error(err))
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows > 0 {
		return nil
	}
	// An update that changes no column value reports zero affected rows just
	// like an update that matches no row, so probe for the row to tell the two
	// apart instead of reporting a 404 for a no-op update.
	return r.requireExists(ctx, businessID, id)
}

func (r *Repository) SetActive(ctx context.Context, businessID, id int64, active bool) error {
	res, err := r.queries.SetRecurringInvoiceActive(ctx, sqlc.SetRecurringInvoiceActiveParams{
		IsActive:   sql.NullBool{Bool: active, Valid: true},
		ID:         id,
		BusinessID: businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update recurring invoice active flag", logger.Error(err))
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows > 0 {
		return nil
	}
	// Pausing an already-paused schedule (or resuming an active one) changes no
	// column value, so zero affected rows is not proof that the row is missing.
	return r.requireExists(ctx, businessID, id)
}

// requireExists maps a zero-affected update onto the domain not-found error
// unless the row is actually present.
func (r *Repository) requireExists(ctx context.Context, businessID, id int64) error {
	exists, err := r.queries.RecurringInvoiceExists(ctx, sqlc.RecurringInvoiceExistsParams{
		ID:         id,
		BusinessID: businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to check recurring invoice existence", logger.Error(err))
		return err
	}
	if !exists {
		return domain.ErrRecurringInvoiceNotFound
	}
	return nil
}

// ProcessDue creates the next invoice for each due recurring schedule. The
// schedule row, copied invoice, copied line items, and next date are committed
// together so a retry after a failure cannot duplicate a billing period.
func (r *Repository) ProcessDue(ctx context.Context, now time.Time) ([]int64, error) {
	rows, err := r.pool.QueryContext(ctx, `
		SELECT r.id, r.business_id
		FROM finance_recurring_invoices r
		JOIN finance_invoices template ON template.id = r.template_invoice_id
		  AND template.business_id = r.business_id
		WHERE r.is_active = TRUE
		  AND r.next_invoice_date <= ?
		  AND (r.end_date IS NULL OR r.end_date >= ?)
		ORDER BY r.next_invoice_date ASC, r.id ASC
		LIMIT 100`, now, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var due []dueSchedule
	for rows.Next() {
		var item dueSchedule
		if err := rows.Scan(&item.id, &item.businessID); err != nil {
			return nil, err
		}
		due = append(due, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return processDueSchedules(ctx, now, due, r.generateDue), nil
}

func processDueSchedules(
	ctx context.Context,
	now time.Time,
	due []dueSchedule,
	generate func(context.Context, int64, int64, time.Time) (bool, error),
) []int64 {
	businessIDs := make([]int64, 0, len(due))
	for _, item := range due {
		created, err := generate(ctx, item.id, item.businessID, now)
		if errors.Is(err, sql.ErrNoRows) {
			continue
		}
		if err != nil {
			logger.Ctx(ctx).Error("Failed to process recurring invoice schedule",
				logger.Int("schedule_id", int(item.id)),
				logger.Int("business_id", int(item.businessID)),
				logger.Error(err),
			)
			continue
		}
		if created {
			businessIDs = append(businessIDs, item.businessID)
		}
	}
	return businessIDs
}

type recurringTemplate struct {
	clientName                                       string
	clientEmail, clientAddress, clientTaxID          sql.NullString
	currency, discountType, notes, terms             sql.NullString
	issueDate, dueDate                               time.Time
	subtotal, discountValue, discountAmount, taxRate string
	taxAmount, total                                 string
}

func (r *Repository) generateDue(ctx context.Context, id, businessID int64, now time.Time) (bool, error) {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	var schedule struct {
		templateID int64
		frequency  string
		nextDate   time.Time
		endDate    sql.NullTime
	}
	err = tx.QueryRowContext(ctx, `
		SELECT template_invoice_id, frequency, next_invoice_date, end_date
		FROM finance_recurring_invoices
		WHERE id = ? AND business_id = ? AND is_active = TRUE
		  AND next_invoice_date <= ?
		  AND (end_date IS NULL OR end_date >= ?)
		FOR UPDATE`, id, businessID, now, now).Scan(
		&schedule.templateID, &schedule.frequency, &schedule.nextDate, &schedule.endDate,
	)
	if err != nil {
		return false, err
	}

	var template recurringTemplate
	err = tx.QueryRowContext(ctx, `
		SELECT client_name, client_email, client_address, client_tax_id,
		       issue_date, due_date, currency, subtotal, discount_type,
		       discount_value, discount_amount, tax_rate, tax_amount, total,
		       notes, terms
		FROM finance_invoices
		WHERE id = ? AND business_id = ?`, schedule.templateID, businessID).Scan(
		&template.clientName, &template.clientEmail, &template.clientAddress, &template.clientTaxID,
		&template.issueDate, &template.dueDate, &template.currency, &template.subtotal,
		&template.discountType, &template.discountValue, &template.discountAmount, &template.taxRate,
		&template.taxAmount, &template.total, &template.notes, &template.terms,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return false, domain.ErrRecurringInvoiceTemplateMissing
	}
	if err != nil {
		return false, err
	}

	dueDate := schedule.nextDate.Add(template.dueDate.Sub(template.issueDate))
	invoiceNumber := fmt.Sprintf("REC-%d-%s", id, schedule.nextDate.Format("20060102"))
	result, err := tx.ExecContext(ctx, `
		INSERT INTO finance_invoices
		(owner_id, business_id, invoice_number, client_name, client_email, client_address, client_tax_id,
		 status, issue_date, due_date, currency, subtotal, discount_type, discount_value,
		 discount_amount, tax_rate, tax_amount, total, amount_paid, amount_due, notes, terms)
		VALUES (?, ?, ?, ?, ?, ?, ?, 'draft', ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, ?, ?, ?)`,
		businessID, businessID, invoiceNumber, template.clientName, template.clientEmail, template.clientAddress,
		template.clientTaxID, schedule.nextDate, dueDate, template.currency, template.subtotal, template.discountType,
		template.discountValue, template.discountAmount, template.taxRate, template.taxAmount, template.total,
		template.total, template.notes, template.terms,
	)
	if err != nil {
		return false, err
	}
	newInvoiceID, err := result.LastInsertId()
	if err != nil {
		return false, err
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO finance_invoice_items
		(invoice_id, description, quantity, unit, unit_price, tax_rate, discount_pct, total, sort_order)
		SELECT ?, description, quantity, unit, unit_price, tax_rate, discount_pct, total, sort_order
		FROM finance_invoice_items WHERE invoice_id = ?`, newInvoiceID, schedule.templateID); err != nil {
		return false, err
	}

	nextDate, err := advanceDate(schedule.nextDate, schedule.frequency)
	if err != nil {
		return false, err
	}
	active := !schedule.endDate.Valid || !nextDate.After(schedule.endDate.Time)
	if _, err := tx.ExecContext(ctx, `
		UPDATE finance_recurring_invoices
		SET next_invoice_date = ?, is_active = ?
		WHERE id = ? AND business_id = ?`, nextDate, active, id, businessID); err != nil {
		return false, err
	}

	if err := tx.Commit(); err != nil {
		return false, err
	}
	committed = true
	return true, nil
}

// ---------------------------------------------------------------------------
// Conversion helpers
// ---------------------------------------------------------------------------

func nullBool(b sql.NullBool) bool {
	if b.Valid {
		return b.Bool
	}
	return false
}

func nullTime(t sql.NullTime) time.Time {
	if t.Valid {
		return t.Time
	}
	return time.Time{}
}

// parseEndDate maps the optional end_date contract field onto the NULL-able
// end_date column. nil means "not supplied" (on update that leaves the stored
// value untouched); an empty value clears the bound. A malformed value is an
// error rather than a silently open-ended schedule.
func parseEndDate(value *string) (sql.NullTime, error) {
	if value == nil {
		return sql.NullTime{}, nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return sql.NullTime{}, nil
	}
	parsed, err := time.Parse("2006-01-02", trimmed)
	if err != nil {
		return sql.NullTime{}, fmt.Errorf("end_date must be in YYYY-MM-DD format: %w", err)
	}
	return sql.NullTime{Time: parsed, Valid: true}, nil
}

func nullBoolPtr(b *bool) sql.NullBool {
	if b == nil {
		return sql.NullBool{}
	}
	return sql.NullBool{Bool: *b, Valid: true}
}

func nullInt64Ptr(i *int64) sql.NullInt64 {
	if i == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *i, Valid: true}
}

func advanceDate(date time.Time, frequency string) (time.Time, error) {
	switch frequency {
	case "weekly":
		return date.AddDate(0, 0, 7), nil
	case "monthly":
		return advanceMonths(date, 1), nil
	case "quarterly":
		return advanceMonths(date, 3), nil
	case "yearly":
		return advanceMonths(date, 12), nil
	default:
		return time.Time{}, fmt.Errorf("unsupported recurring invoice frequency %q", frequency)
	}
}

func advanceMonths(date time.Time, months int) time.Time {
	target := time.Date(date.Year(), date.Month(), 1, date.Hour(), date.Minute(), date.Second(), date.Nanosecond(), date.Location()).AddDate(0, months, 0)
	lastDay := time.Date(target.Year(), target.Month()+1, 0, date.Hour(), date.Minute(), date.Second(), date.Nanosecond(), date.Location()).Day()
	day := date.Day()
	if day > lastDay {
		day = lastDay
	}
	return time.Date(target.Year(), target.Month(), day, date.Hour(), date.Minute(), date.Second(), date.Nanosecond(), date.Location())
}
