package application

import (
	"context"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/invoices/internal/domain"
	"github.com/voxire/onion/internal/modules/invoices/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/logger"
)

type Service struct {
	repo *infrastructure.Repository
}

func NewService(ctx context.Context) *Service {
	repo := infrastructure.NewRepository(ctx)

	return &Service{
		repo: repo,
	}
}

// ---------------------------------------------------------------------------
// Invoices
// ---------------------------------------------------------------------------

func (s *Service) ListInvoices(ctx context.Context, ownerID int64, status, search, dateFrom, dateTo string) ([]domain.Invoice, error) {
	return s.repo.ListInvoices(ctx, ownerID, status, search, dateFrom, dateTo)
}

func (s *Service) GetInvoice(ctx context.Context, ownerID, invoiceID int64) (*domain.Invoice, error) {
	inv, err := s.repo.GetInvoiceByID(ctx, ownerID, invoiceID)
	if err != nil {
		return nil, err
	}

	items, err := s.repo.GetInvoiceItems(ctx, inv.ID)
	if err != nil {
		return nil, err
	}

	inv.Items = items

	payments, err := s.repo.GetPaymentsByInvoice(ctx, inv.ID)
	if err != nil {
		return nil, err
	}

	inv.Payments = payments

	return inv, nil
}

func (s *Service) CreateInvoice(ctx context.Context, ownerID int64, input domain.InvoiceInput) (*domain.Invoice, error) {
	issueDate, err := time.Parse("2006-01-02", input.IssueDate)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to parse issue date", logger.Error(err))

		return nil, err
	}

	dueDate, err := time.Parse("2006-01-02", input.DueDate)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to parse due date", logger.Error(err))

		return nil, err
	}

	if dueDate.Before(issueDate) {
		return nil, domain.ErrInvalidInvoiceDate
	}

	// Auto-generate invoice number: INV-YYYY-XXXX
	year := issueDate.Year()

	seq, err := s.repo.CountInvoiceSequence(ctx, ownerID, year)
	if err != nil {
		return nil, err
	}

	invoiceNumber := fmt.Sprintf("INV-%d-%04d", year, seq+1)

	// Calculate line-item totals
	var subtotal float64

	items := make([]domain.InvoiceItem, 0, len(input.Items))

	for i, itemInput := range input.Items {
		lineTotal := itemInput.CalculateTotal()
		subtotal += lineTotal

		unit := itemInput.Unit
		if unit == "" {
			unit = "unit"
		}

		items = append(items, domain.InvoiceItem{
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

	// Apply header-level discount
	var discountAmount float64

	discountType := input.DiscountType
	if discountType == "" {
		discountType = "fixed"
	}

	switch discountType {
	case "percentage":
		discountAmount = subtotal * (input.DiscountValue / 100)
	default:
		discountAmount = input.DiscountValue
	}

	afterDiscount := subtotal - discountAmount

	// Header-level tax is applied on top of the item totals (which may already include per-item tax)
	taxAmount := afterDiscount * (input.TaxRate / 100)
	total := afterDiscount + taxAmount

	currency := input.Currency
	if currency == "" {
		currency = "USD"
	}

	inv := domain.Invoice{
		OwnerID:        ownerID,
		InvoiceNumber:  invoiceNumber,
		ClientName:     input.ClientName,
		ClientEmail:    input.ClientEmail,
		ClientAddress:  input.ClientAddress,
		ClientTaxID:    input.ClientTaxID,
		Status:         "draft",
		IssueDate:      issueDate,
		DueDate:        dueDate,
		Currency:       currency,
		Subtotal:       subtotal,
		DiscountType:   discountType,
		DiscountValue:  input.DiscountValue,
		DiscountAmount: discountAmount,
		TaxRate:        input.TaxRate,
		TaxAmount:      taxAmount,
		Total:          total,
		AmountPaid:     0,
		AmountDue:      total,
		Notes:          input.Notes,
		Terms:          input.Terms,
	}

	invID, err := s.repo.CreateInvoice(ctx, ownerID, inv)
	if err != nil {
		return nil, err
	}

	inv.ID = invID
	inv.Items = items

	if err := s.repo.CreateInvoiceItems(ctx, invID, items); err != nil {
		return nil, err
	}

	return &inv, nil
}

func (s *Service) UpdateInvoice(ctx context.Context, ownerID, invoiceID int64, input domain.InvoiceUpdateInput) (*domain.Invoice, error) {
	existing, err := s.repo.GetInvoiceByID(ctx, ownerID, invoiceID)
	if err != nil {
		return nil, err
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

	items := make([]domain.InvoiceItem, 0, len(input.Items))

	for i, itemInput := range input.Items {
		lineTotal := itemInput.CalculateTotal()
		subtotal += lineTotal

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

	var discountAmount float64

	switch discountType {
	case "percentage":
		discountAmount = subtotal * (input.DiscountValue / 100)
	default:
		discountAmount = input.DiscountValue
	}

	afterDiscount := subtotal - discountAmount
	taxAmount := afterDiscount * (input.TaxRate / 100)
	total := afterDiscount + taxAmount

	currency := input.Currency
	if currency == "" {
		currency = "USD"
	}

	existing.ClientName = input.ClientName
	existing.ClientEmail = input.ClientEmail
	existing.ClientAddress = input.ClientAddress
	existing.ClientTaxID = input.ClientTaxID
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

	if err := s.repo.UpdateInvoice(ctx, ownerID, *existing); err != nil {
		return nil, err
	}

	if err := s.repo.DeleteInvoiceItems(ctx, invoiceID); err != nil {
		return nil, err
	}

	if err := s.repo.CreateInvoiceItems(ctx, invoiceID, items); err != nil {
		return nil, err
	}

	existing.Items = items

	return existing, nil
}

func (s *Service) UpdateInvoiceStatus(ctx context.Context, ownerID, invoiceID int64, status string) error {
	return s.repo.UpdateInvoiceStatus(ctx, ownerID, invoiceID, status)
}

func (s *Service) DeleteInvoice(ctx context.Context, ownerID, invoiceID int64) error {
	return s.repo.DeleteInvoice(ctx, ownerID, invoiceID)
}

// ---------------------------------------------------------------------------
// Payments
// ---------------------------------------------------------------------------

func (s *Service) RecordPayment(ctx context.Context, ownerID int64, input domain.PaymentInput) (*domain.Payment, error) {
	inv, err := s.repo.GetInvoiceByID(ctx, ownerID, input.InvoiceID)
	if err != nil {
		return nil, err
	}

	paymentDate, err := time.Parse("2006-01-02", input.PaymentDate)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to parse payment date", logger.Error(err))

		return nil, err
	}

	currency := input.Currency
	if currency == "" {
		currency = inv.Currency
	}

	p := domain.Payment{
		OwnerID:     ownerID,
		InvoiceID:   input.InvoiceID,
		Amount:      input.Amount,
		Currency:    currency,
		PaymentDate: paymentDate,
		Method:      input.Method,
		Reference:   input.Reference,
		Notes:       input.Notes,
	}

	paymentID, err := s.repo.CreatePayment(ctx, ownerID, p)
	if err != nil {
		return nil, err
	}

	p.ID = paymentID

	// Update invoice payment totals and status
	newAmountPaid := inv.AmountPaid + input.Amount
	newAmountDue := inv.Total - newAmountPaid

	if newAmountDue < 0 {
		newAmountDue = 0
	}

	var newStatus string
	var paidDate *time.Time

	switch {
	case newAmountDue <= 0:
		newStatus = "paid"
		now := paymentDate
		paidDate = &now
	case newAmountPaid > 0:
		newStatus = "partial"
	default:
		newStatus = inv.Status
	}

	if err := s.repo.UpdateInvoicePaymentTotals(ctx, input.InvoiceID, newAmountPaid, newAmountDue, newStatus, paidDate); err != nil {
		return nil, err
	}

	return &p, nil
}

// ---------------------------------------------------------------------------
// Expenses
// ---------------------------------------------------------------------------

func (s *Service) ListExpenses(ctx context.Context, ownerID int64, categoryID int64, search, dateFrom, dateTo string) ([]domain.Expense, error) {
	return s.repo.ListExpenses(ctx, ownerID, categoryID, search, dateFrom, dateTo)
}

func (s *Service) CreateExpense(ctx context.Context, ownerID int64, input domain.ExpenseInput) (*domain.Expense, error) {
	expenseDate, err := time.Parse("2006-01-02", input.ExpenseDate)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to parse expense date", logger.Error(err))

		return nil, err
	}

	currency := input.Currency
	if currency == "" {
		currency = "USD"
	}

	paymentMethod := input.PaymentMethod
	if paymentMethod == "" {
		paymentMethod = "cash"
	}

	e := domain.Expense{
		OwnerID:         ownerID,
		CategoryID:      input.CategoryID,
		Description:     input.Description,
		Amount:          input.Amount,
		Currency:        currency,
		ExpenseDate:     expenseDate,
		Vendor:          input.Vendor,
		PaymentMethod:   paymentMethod,
		ReceiptURL:      input.ReceiptURL,
		IsTaxDeductible: input.IsTaxDeductible,
		TaxAmount:       input.TaxAmount,
		Notes:           input.Notes,
	}

	id, err := s.repo.CreateExpense(ctx, ownerID, e)
	if err != nil {
		return nil, err
	}

	e.ID = id

	return &e, nil
}

func (s *Service) UpdateExpense(ctx context.Context, ownerID, expenseID int64, input domain.ExpenseInput) (*domain.Expense, error) {
	existing, err := s.repo.GetExpenseByID(ctx, ownerID, expenseID)
	if err != nil {
		return nil, err
	}

	expenseDate, err := time.Parse("2006-01-02", input.ExpenseDate)
	if err != nil {
		return nil, err
	}

	currency := input.Currency
	if currency == "" {
		currency = "USD"
	}

	paymentMethod := input.PaymentMethod
	if paymentMethod == "" {
		paymentMethod = "cash"
	}

	existing.CategoryID = input.CategoryID
	existing.Description = input.Description
	existing.Amount = input.Amount
	existing.Currency = currency
	existing.ExpenseDate = expenseDate
	existing.Vendor = input.Vendor
	existing.PaymentMethod = paymentMethod
	existing.ReceiptURL = input.ReceiptURL
	existing.IsTaxDeductible = input.IsTaxDeductible
	existing.TaxAmount = input.TaxAmount
	existing.Notes = input.Notes

	if err := s.repo.UpdateExpense(ctx, ownerID, *existing); err != nil {
		return nil, err
	}

	return existing, nil
}

func (s *Service) DeleteExpense(ctx context.Context, ownerID, expenseID int64) error {
	return s.repo.DeleteExpense(ctx, ownerID, expenseID)
}

// ---------------------------------------------------------------------------
// Expense Categories
// ---------------------------------------------------------------------------

func (s *Service) ListExpenseCategories(ctx context.Context, ownerID int64) ([]domain.ExpenseCategory, error) {
	return s.repo.ListExpenseCategories(ctx, ownerID)
}

func (s *Service) CreateExpenseCategory(ctx context.Context, ownerID int64, input domain.ExpenseCategoryInput) (*domain.ExpenseCategory, error) {
	c := domain.ExpenseCategory{
		OwnerID: ownerID,
		Name:    input.Name,
		Color:   input.Color,
		Icon:    input.Icon,
	}

	id, err := s.repo.CreateExpenseCategory(ctx, ownerID, c)
	if err != nil {
		return nil, err
	}

	c.ID = id

	return &c, nil
}

// ---------------------------------------------------------------------------
// Tax Config
// ---------------------------------------------------------------------------

func (s *Service) ListTaxConfigs(ctx context.Context, ownerID int64) ([]domain.TaxConfig, error) {
	return s.repo.ListTaxConfigs(ctx, ownerID)
}

func (s *Service) CreateTaxConfig(ctx context.Context, ownerID int64, input domain.TaxConfigInput) (*domain.TaxConfig, error) {
	taxType := input.Type
	if taxType == "" {
		taxType = "vat"
	}

	t := domain.TaxConfig{
		OwnerID:   ownerID,
		Name:      input.Name,
		Rate:      input.Rate,
		Type:      taxType,
		IsDefault: input.IsDefault,
		IsActive:  true,
	}

	id, err := s.repo.CreateTaxConfig(ctx, ownerID, t)
	if err != nil {
		return nil, err
	}

	t.ID = id

	return &t, nil
}

// ---------------------------------------------------------------------------
// Reports
// ---------------------------------------------------------------------------

func (s *Service) GetPLReport(ctx context.Context, ownerID int64, period, date string) (domain.PLReport, error) {
	dateFrom, dateTo := resolvePeriodRange(period, date)

	revenue, err := s.repo.SumPaidInvoices(ctx, ownerID, dateFrom, dateTo)
	if err != nil {
		return domain.PLReport{}, err
	}

	expenses, err := s.repo.SumExpenses(ctx, ownerID, dateFrom, dateTo)
	if err != nil {
		return domain.PLReport{}, err
	}

	// For simplicity COGS = 0 (can be linked to ERP in the future).
	// Operating expenses = all recorded expenses.
	grossProfit := revenue
	netProfit := grossProfit - expenses

	var marginPct float64
	if revenue > 0 {
		marginPct = (netProfit / revenue) * 100
	}

	return domain.PLReport{
		Period:            period,
		Revenue:           revenue,
		COGS:              0,
		GrossProfit:       grossProfit,
		OperatingExpenses: expenses,
		NetProfit:         netProfit,
		MarginPct:         marginPct,
	}, nil
}

func (s *Service) GetDashboard(ctx context.Context, ownerID int64) (domain.Dashboard, error) {
	totalReceivable, err := s.repo.SumTotalReceivable(ctx, ownerID)
	if err != nil {
		return domain.Dashboard{}, err
	}

	overdueAmount, err := s.repo.SumOverdueAmount(ctx, ownerID)
	if err != nil {
		return domain.Dashboard{}, err
	}

	now := time.Now()
	mtdFrom := fmt.Sprintf("%d-%02d-01", now.Year(), now.Month())
	mtdTo := now.Format("2006-01-02")

	totalExpensesMTD, err := s.repo.SumExpenses(ctx, ownerID, mtdFrom, mtdTo)
	if err != nil {
		return domain.Dashboard{}, err
	}

	revenueMTD, err := s.repo.SumPaidInvoices(ctx, ownerID, mtdFrom, mtdTo)
	if err != nil {
		return domain.Dashboard{}, err
	}

	netProfitMTD := revenueMTD - totalExpensesMTD

	return domain.Dashboard{
		TotalReceivable:  totalReceivable,
		OverdueAmount:    overdueAmount,
		TotalExpensesMTD: totalExpensesMTD,
		NetProfitMTD:     netProfitMTD,
	}, nil
}

func (s *Service) GetTaxReport(ctx context.Context, ownerID int64, period, date string) (domain.TaxReport, error) {
	dateFrom, dateTo := resolvePeriodRange(period, date)

	taxCollected, err := s.repo.SumTaxCollected(ctx, ownerID, dateFrom, dateTo)
	if err != nil {
		return domain.TaxReport{}, err
	}

	taxOwed, err := s.repo.SumTaxOwed(ctx, ownerID, dateFrom, dateTo)
	if err != nil {
		return domain.TaxReport{}, err
	}

	return domain.TaxReport{
		Period:       period,
		TaxCollected: taxCollected,
		TaxOwed:      taxOwed,
	}, nil
}

// resolvePeriodRange converts a period keyword + anchor date into a date range.
// date should be in YYYY-MM-DD format; if empty, today is used.
func resolvePeriodRange(period, date string) (string, string) {
	var anchor time.Time

	if date != "" {
		if t, err := time.Parse("2006-01-02", date); err == nil {
			anchor = t
		}
	}

	if anchor.IsZero() {
		anchor = time.Now()
	}

	switch period {
	case "month":
		from := time.Date(anchor.Year(), anchor.Month(), 1, 0, 0, 0, 0, time.UTC)
		to := from.AddDate(0, 1, -1)

		return from.Format("2006-01-02"), to.Format("2006-01-02")

	case "quarter":
		quarter := (int(anchor.Month()) - 1) / 3
		fromMonth := time.Month(quarter*3 + 1)
		from := time.Date(anchor.Year(), fromMonth, 1, 0, 0, 0, 0, time.UTC)
		to := from.AddDate(0, 3, -1)

		return from.Format("2006-01-02"), to.Format("2006-01-02")

	case "year":
		from := time.Date(anchor.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(anchor.Year(), 12, 31, 0, 0, 0, 0, time.UTC)

		return from.Format("2006-01-02"), to.Format("2006-01-02")

	default:
		// default to current month
		from := time.Date(anchor.Year(), anchor.Month(), 1, 0, 0, 0, 0, time.UTC)
		to := from.AddDate(0, 1, -1)

		return from.Format("2006-01-02"), to.Format("2006-01-02")
	}
}
