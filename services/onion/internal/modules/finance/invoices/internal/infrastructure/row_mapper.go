package infrastructure

import (
	"database/sql"

	"github.com/voxire/onion/internal/modules/finance/invoices/internal/domain"
)

type invoiceRowScanner interface {
	Scan(...any) error
}

func invoiceFromRow(row invoiceRowScanner) (domain.Invoice, error) {
	var invoice domain.Invoice
	var clientEmail, clientAddress, clientTaxID sql.NullString
	var paidDate sql.NullTime
	var currency, discountType, notes, terms sql.NullString
	if err := row.Scan(
		&invoice.ID,
		&invoice.OwnerID,
		&invoice.BusinessID,
		&invoice.InvoiceNumber,
		&invoice.ClientName,
		&clientEmail,
		&clientAddress,
		&clientTaxID,
		&invoice.Status,
		&invoice.IssueDate,
		&invoice.DueDate,
		&paidDate,
		&currency,
		&invoice.Subtotal,
		&discountType,
		&invoice.DiscountValue,
		&invoice.DiscountAmount,
		&invoice.TaxRate,
		&invoice.TaxAmount,
		&invoice.Total,
		&invoice.AmountPaid,
		&invoice.AmountDue,
		&notes,
		&terms,
		&invoice.CreatedAt,
		&invoice.UpdatedAt,
	); err != nil {
		return domain.Invoice{}, err
	}

	if clientEmail.Valid {
		invoice.ClientEmail = clientEmail.String
	}
	if clientAddress.Valid {
		invoice.ClientAddress = clientAddress.String
	}
	if clientTaxID.Valid {
		invoice.ClientTaxID = clientTaxID.String
	}
	if paidDate.Valid {
		invoice.PaidDate = &paidDate.Time
	}
	if currency.Valid {
		invoice.Currency = currency.String
	} else {
		invoice.Currency = "USD"
	}
	if discountType.Valid {
		invoice.DiscountType = discountType.String
	} else {
		invoice.DiscountType = "fixed"
	}
	if notes.Valid {
		invoice.Notes = notes.String
	}
	if terms.Valid {
		invoice.Terms = terms.String
	}
	return invoice, nil
}

func expenseFromRow(row invoiceRowScanner) (domain.Expense, error) {
	var expense domain.Expense
	var categoryID sql.NullInt64
	var currency, vendor, paymentMethod, receiptURL, notes, status sql.NullString
	if err := row.Scan(
		&expense.ID,
		&expense.OwnerID,
		&expense.BusinessID,
		&categoryID,
		&expense.Description,
		&expense.Amount,
		&currency,
		&expense.ExpenseDate,
		&vendor,
		&paymentMethod,
		&receiptURL,
		&expense.IsTaxDeductible,
		&expense.TaxAmount,
		&notes,
		&status,
		&expense.CreatedAt,
		&expense.UpdatedAt,
	); err != nil {
		return domain.Expense{}, err
	}

	if categoryID.Valid {
		expense.CategoryID = categoryID.Int64
	}
	if currency.Valid {
		expense.Currency = currency.String
	} else {
		expense.Currency = "USD"
	}
	if vendor.Valid {
		expense.Vendor = vendor.String
	}
	if paymentMethod.Valid {
		expense.PaymentMethod = paymentMethod.String
	} else {
		expense.PaymentMethod = "cash"
	}
	if receiptURL.Valid {
		expense.ReceiptURL = receiptURL.String
	}
	if notes.Valid {
		expense.Notes = notes.String
	}
	if status.Valid {
		expense.Status = status.String
	}
	return expense, nil
}
