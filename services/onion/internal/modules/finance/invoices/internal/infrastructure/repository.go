package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"github.com/voxire/onion/internal/modules/finance/internal/outbox"
	"strings"

	"github.com/Masterminds/squirrel"

	crmcontract "github.com/voxire/onion/internal/contracts/crm"
	"github.com/voxire/onion/internal/modules/finance/internal/invoicenumber"
	"github.com/voxire/onion/internal/modules/finance/invoices/internal/domain"
	"github.com/voxire/onion/internal/modules/finance/invoices/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/eventbus"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/searchutil"
)

type Repository struct {
	queries *sqlc.Queries
	pool    *sql.DB
}

// CreateInvoice retains the legacy header-only repository operation used by
// integration fixtures. Application writes use CreateInvoiceWithItems so the
// header and items share one transaction.
func (r *Repository) CreateInvoice(ctx context.Context, businessID int64, inv domain.Invoice) (int64, error) {
	created, err := r.CreateInvoiceWithItems(ctx, businessID, inv, nil)
	return created.ID, err
}

func NewRepository(db *sql.DB) *Repository {
	pool := db

	return &Repository{
		queries: sqlc.New(pool),
		pool:    pool,
	}
}

// ---------------------------------------------------------------------------
// Invoices
// ---------------------------------------------------------------------------

// ListInvoices returns a page of invoices using the same filter pipeline as the
// customer-scoped path. Customer filtering is optional so every caller gets
// identical status, search, and date semantics.
func (r *Repository) ListInvoices(ctx context.Context, businessID int64, status, search, dateFrom, dateTo string, limit, offset int) ([]domain.Invoice, error) {
	return r.listInvoices(ctx, businessID, 0, status, search, dateFrom, dateTo, limit, offset)
}

// ListInvoicesForCustomer is the customer-scoped variant of ListInvoices. It
// deliberately shares the full filter pipeline; a customer filter must not
// silently turn status, search, or date filters into no-ops.
func (r *Repository) ListInvoicesForCustomer(ctx context.Context, businessID, customerID int64, status, search, dateFrom, dateTo string, limit, offset int) ([]domain.Invoice, error) {
	return r.listInvoices(ctx, businessID, customerID, status, search, dateFrom, dateTo, limit, offset)
}

// dynamic-query: squirrel
func (r *Repository) listInvoices(ctx context.Context, businessID, customerID int64, status, search, dateFrom, dateTo string, limit, offset int) ([]domain.Invoice, error) {
	qb := squirrel.Select(
		"id", "owner_id", "business_id", "invoice_number", "client_name", "client_email", "client_address", "client_tax_id",
		"status", "issue_date", "due_date", "paid_date", "currency", "subtotal", "discount_type",
		"discount_value", "discount_amount", "tax_rate", "tax_amount", "total", "amount_paid", "amount_due",
		"notes", "terms", "created_at", "updated_at",
	).
		From("finance_invoices").
		Where(squirrel.Eq{"business_id": businessID})

	if customerID > 0 {
		qb = qb.Where(squirrel.Eq{"customer_id": customerID})
	}

	if status != "" {
		qb = qb.Where(squirrel.Eq{"status": status})
	}

	if search = strings.TrimSpace(search); search != "" {
		booleanQuery := searchutil.BooleanPrefixQuery(search)
		if booleanQuery != "" {
			qb = qb.Where(
				"(invoice_number = ? OR client_email = ? OR MATCH(invoice_number, client_name, client_email) AGAINST (? IN BOOLEAN MODE))",
				search,
				search,
				booleanQuery,
			)
		} else {
			qb = qb.Where(squirrel.Or{squirrel.Eq{"invoice_number": search}, squirrel.Eq{"client_email": search}})
		}
	}

	if dateFrom != "" {
		qb = qb.Where(squirrel.GtOrEq{"issue_date": dateFrom})
	}

	if dateTo != "" {
		qb = qb.Where(squirrel.LtOrEq{"issue_date": dateTo})
	}

	qb = qb.OrderBy("created_at DESC, id DESC").Limit(uint64(limit)).Offset(uint64(offset)) // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.

	rows, err := qb.PlaceholderFormat(squirrel.Question).RunWith(r.pool).QueryContext(ctx)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list invoices", logger.Error(err))

		return nil, err
	}
	defer rows.Close()

	invoices := make([]domain.Invoice, 0, limit)

	for rows.Next() {
		inv, scanErr := invoiceFromRow(rows)
		if scanErr != nil {
			logger.Ctx(ctx).Error("Failed to scan invoice row", logger.Error(scanErr))

			return nil, scanErr
		}

		invoices = append(invoices, inv)
	}

	return invoices, rows.Err()
}

func (r *Repository) GetInvoiceByID(ctx context.Context, businessID, invoiceID int64) (*domain.Invoice, error) {
	row, err := r.queries.GetInvoiceByID(ctx, sqlc.GetInvoiceByIDParams{
		ID:         invoiceID,
		BusinessID: businessID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrInvoiceNotFound
		}
		logger.Ctx(ctx).Error("Failed to get invoice by ID", logger.Error(err))

		return nil, err
	}

	inv := invoiceFromSqlc(row)

	return &inv, nil
}

func (r *Repository) CreateInvoiceWithItems(ctx context.Context, businessID int64, inv domain.Invoice, items []domain.InvoiceItem) (domain.Invoice, error) {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to begin invoice transaction", logger.Error(err))
		return domain.Invoice{}, err
	}
	// Rollback is a no-op after a successful commit and also fires on panic,
	// so an interrupted transaction never lingers on the pooled connection.
	defer func() { _ = tx.Rollback() }()

	if inv.InvoiceNumber == "" {
		inv.InvoiceNumber, err = invoicenumber.Allocate(ctx, tx, businessID, inv.IssueDate.Year())
		if err != nil {
			return domain.Invoice{}, err
		}
	}

	result, err := r.queries.WithTx(tx).CreateInvoice(ctx, sqlc.CreateInvoiceParams{
		OwnerID:        businessID,
		BusinessID:     businessID,
		InvoiceNumber:  inv.InvoiceNumber,
		ClientName:     inv.ClientName,
		ClientEmail:    sqlNullString(inv.ClientEmail),
		ClientAddress:  sqlNullString(inv.ClientAddress),
		ClientTaxID:    sqlNullString(inv.ClientTaxID),
		CustomerID:     nullableInt64FromPtr(inv.CustomerID),
		Status:         sqlc.NullFinanceInvoicesStatus{FinanceInvoicesStatus: sqlc.FinanceInvoicesStatus(inv.Status), Valid: inv.Status != ""},
		IssueDate:      inv.IssueDate,
		DueDate:        inv.DueDate,
		Currency:       sqlNullString(inv.Currency),
		Subtotal:       sqlNullString(fmtFloat(inv.Subtotal)),
		DiscountType:   sqlc.NullFinanceInvoicesDiscountType{FinanceInvoicesDiscountType: sqlc.FinanceInvoicesDiscountType(inv.DiscountType), Valid: inv.DiscountType != ""},
		DiscountValue:  sqlNullString(fmtFloat(inv.DiscountValue)),
		DiscountAmount: sqlNullString(fmtFloat(inv.DiscountAmount)),
		TaxRate:        sqlNullString(fmtFloat(inv.TaxRate)),
		TaxAmount:      sqlNullString(fmtFloat(inv.TaxAmount)),
		Total:          sqlNullString(fmtFloat(inv.Total)),
		AmountPaid:     sqlNullString(fmtFloat(inv.AmountPaid)),
		AmountDue:      sqlNullString(fmtFloat(inv.AmountDue)),
		Notes:          sqlNullString(inv.Notes),
		Terms:          sqlNullString(inv.Terms),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create invoice", logger.Error(err))

		return domain.Invoice{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		logger.Ctx(ctx).Error("Failed to read created invoice ID", logger.Error(err))

		return domain.Invoice{}, err
	}
	if err = createInvoiceItems(ctx, tx, id, items); err != nil {
		logger.Ctx(ctx).Error("Failed to create invoice items", logger.Error(err))
		return domain.Invoice{}, err
	}
	contact := crmcontract.CustomerContactCaptured{Name: inv.ClientName, Email: inv.ClientEmail}
	if err = outbox.Publish(ctx, tx, eventbus.Event{
		BusinessID:  businessID,
		Type:        crmcontract.CustomerContactCapturedType,
		SubjectType: "finance_invoice",
		SubjectID:   id,
		Payload:     contact,
		DedupeKey:   crmcontract.CustomerContactCaptureDedupeKey("finance_invoice", id, contact),
	}); err != nil {
		return domain.Invoice{}, err
	}

	if err = tx.Commit(); err != nil {
		logger.Ctx(ctx).Error("Failed to commit invoice transaction", logger.Error(err))
		return domain.Invoice{}, err
	}

	inv.ID = id
	inv.BusinessID = businessID
	inv.Items = items
	return inv, nil
}
