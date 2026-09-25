package application

import (
	"context"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/invoices/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
	emailPkg "github.com/voxire/onion/internal/platform/email"
	"github.com/voxire/onion/internal/platform/logger"
)

func (s *Service) UpdateInvoice(ctx context.Context, businessID, invoiceID int64, input domain.InvoiceUpdateInput) (*domain.Invoice, error) {
	defer invalidateInvoicesScope(ctx, businessID)
	existing, err := s.repo.GetInvoiceByID(ctx, businessID, invoiceID)
	if err != nil {
		return nil, err
	}
	if existing.Status != domain.StatusDraft {
		return nil, domain.ErrInvoiceNotDraft
	}

	// Optional re-link on a draft: verify like create, snapshot when present,
	// and keep the existing link when the field is absent.
	var newCustomerID *int64
	if input.CustomerID != nil {
		name, email, address, snapErr := s.resolveCustomerSnapshot(ctx, businessID, *input.CustomerID)
		if snapErr != nil {
			return nil, snapErr
		}
		newCustomerID = input.CustomerID
		input.ClientName = name
		input.ClientEmail = email
		if address != "" {
			input.ClientAddress = address
		}
	} else {
		newCustomerID = existing.CustomerID
	}

	issueDate, err := time.Parse("2006-01-02", input.IssueDate)
	if err != nil {
		return nil, err
	}

	dueDate, err := time.Parse("2006-01-02", input.DueDate)
	if err != nil {
		return nil, err
	}

	if dueDate.Before(issueDate) {
		return nil, domain.ErrInvalidInvoiceDate
	}

	var subtotal float64
	var itemTaxAmount float64

	items := make([]domain.InvoiceItem, 0, len(input.Items))

	for i, itemInput := range input.Items {
		lineTotal := itemInput.CalculateTotal()
		lineComponents := itemInput.CalculateComponents()
		subtotal += lineComponents.Base
		itemTaxAmount += lineComponents.Tax

		unit := itemInput.Unit
		if unit == "" {
			unit = "unit"
		}

		items = append(items, domain.InvoiceItem{
			InvoiceID:   invoiceID,
			Description: itemInput.Description,
			Quantity:    itemInput.Quantity,
			Unit:        unit,
			UnitPrice:   itemInput.UnitPrice,
			TaxRate:     itemInput.TaxRate,
			DiscountPct: itemInput.DiscountPct,
			Total:       lineTotal,
			SortOrder:   i,
		})
	}

	discountType := input.DiscountType
	if discountType == "" {
		discountType = "fixed"
	}

	discountAmount, err := applyHeaderDiscount(subtotal, discountType, input.DiscountValue)
	if err != nil {
		return nil, err
	}

	afterDiscount := subtotal - discountAmount
	taxAmount := afterDiscount*(input.TaxRate/100) + itemTaxAmount
	total := afterDiscount + taxAmount

	currency := input.Currency
	if currency == "" {
		currency = "USD"
	}

	existing.ClientName = input.ClientName
	existing.ClientEmail = input.ClientEmail
	existing.ClientAddress = input.ClientAddress
	existing.ClientTaxID = input.ClientTaxID
	existing.CustomerID = newCustomerID
	existing.IssueDate = issueDate
	existing.DueDate = dueDate
	existing.Currency = currency
	existing.Subtotal = subtotal
	existing.DiscountType = discountType
	existing.DiscountValue = input.DiscountValue
	existing.DiscountAmount = discountAmount
	existing.TaxRate = input.TaxRate
	existing.TaxAmount = taxAmount
	existing.Total = total
	existing.AmountDue = total - existing.AmountPaid
	existing.Notes = input.Notes
	existing.Terms = input.Terms

	if err := s.repo.UpdateInvoiceWithItems(ctx, businessID, *existing, items); err != nil {
		return nil, err
	}

	existing.Items = items

	return existing, nil
}

func (s *Service) UpdateInvoiceStatus(ctx context.Context, businessID, invoiceID int64, status string) error {
	defer invalidateInvoicesScope(ctx, businessID)
	existing, err := s.repo.GetInvoiceByID(ctx, businessID, invoiceID)
	if err != nil {
		return err
	}
	if domain.IsTerminal(existing.Status) || (status == domain.StatusDraft && existing.Status != domain.StatusDraft) {
		return domain.ErrInvoiceNotDraft
	}
	return s.repo.UpdateInvoiceStatus(ctx, businessID, invoiceID, status)
}

func (s *Service) SendInvoice(ctx context.Context, businessID, invoiceID int64) error {
	defer invalidateInvoicesScope(ctx, businessID)
	inv, err := s.GetInvoice(ctx, businessID, invoiceID)
	if err != nil {
		return err
	}

	if err := s.repo.UpdateInvoiceStatus(ctx, businessID, invoiceID, "sent"); err != nil {
		return err
	}

	if inv.ClientEmail != "" && s.db != nil {
		subject := fmt.Sprintf("Invoice %s from rtylr", inv.InvoiceNumber)
		if enqErr := emailPkg.EnqueueTemplate(ctx, s.db, emailPkg.PriorityHigh, inv.ClientEmail, subject, emailPkg.TemplateInvoice, emailPkg.InvoiceData{
			Subject:       subject,
			CustomerName:  inv.ClientName,
			InvoiceNumber: inv.InvoiceNumber,
			Amount:        fmt.Sprintf("%.2f %s", inv.Total, inv.Currency),
			DueDate:       inv.DueDate.Format("2006-01-02"),
		}); enqErr != nil {
			logger.Ctx(ctx).Error("failed to enqueue invoice email", logger.Error(enqErr))
			return enqErr
		}
	}

	return nil
}

func (s *Service) DeleteInvoice(ctx context.Context, businessID, invoiceID int64) error {
	defer invalidateInvoicesScope(ctx, businessID)
	return s.repo.DeleteInvoice(ctx, businessID, invoiceID)
}

// ---------------------------------------------------------------------------
// Payments
// ---------------------------------------------------------------------------

func (s *Service) RecordPayment(ctx context.Context, businessID int64, input domain.PaymentInput) (*domain.Payment, error) {
	defer invalidateInvoicesScope(ctx, businessID)

	if input.Amount <= 0 {
		return nil, domain.ErrInvalidPaymentAmount
	}

	if _, err := time.Parse("2006-01-02", input.PaymentDate); err != nil {
		logger.Ctx(ctx).Error("Failed to parse payment date", logger.Error(err))

		return nil, err
	}
	return s.repo.RecordPayment(ctx, businessID, input)
}

// ---------------------------------------------------------------------------
// Expenses
// ---------------------------------------------------------------------------

func (s *Service) ListExpenses(ctx context.Context, businessID, categoryID int64, search, dateFrom, dateTo string, page, pageSize int64) ([]domain.Expense, bool, error) {
	limit, offset := paginateInvoices(page, pageSize)
	key := cacheKeys.Scoped(invoicesScope, businessID, fmt.Sprintf("expenses|c=%d|q=%s|f=%s|t=%s|p=%d|ps=%d", categoryID, search, dateFrom, dateTo, page, pageSize))
	rows, err := cacher.Of[[]domain.Expense]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) ([]domain.Expense, error) {
			return s.repo.ListExpenses(ctx, businessID, categoryID, search, dateFrom, dateTo, limit+1, offset)
		},
		cacher.WithExpiry(ctx, invoicesTTL),
	)
	if err != nil {
		return nil, false, err
	}
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}
	return rows, hasMore, nil
}
