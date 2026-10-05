package api

import (
	"github.com/voxire/onion/internal/modules/finance/invoices/internal/application"
	"github.com/voxire/onion/internal/modules/finance/invoices/internal/domain"
	financeinvoicesv1 "github.com/voxire/voxie/gen/go/finance/invoices/v1"
)

type CreateInvoiceHandler struct {
	FinanceService *application.Service
}

type InvoiceItemPayload struct {
	Description string  `json:"description"`
	Quantity    float64 `json:"quantity"`
	Unit        string  `json:"unit"`
	UnitPrice   float64 `json:"unit_price"`
	TaxRate     float64 `json:"tax_rate"`
	DiscountPct float64 `json:"discount_pct"`
	SortOrder   int     `json:"sort_order"`
	Total       float64 `json:"total"`
}

type CreateInvoiceRequest struct {
	ClientName    string               `json:"client_name"`
	ClientEmail   string               `json:"client_email"`
	ClientAddress string               `json:"client_address"`
	ClientTaxID   string               `json:"client_tax_id"`
	IssueDate     string               `json:"issue_date"`
	DueDate       string               `json:"due_date"`
	Currency      string               `json:"currency"`
	DiscountType  string               `json:"discount_type"`
	DiscountValue float64              `json:"discount_value"`
	TaxRate       float64              `json:"tax_rate"`
	Notes         string               `json:"notes"`
	Terms         string               `json:"terms"`
	Items         []InvoiceItemPayload `json:"items"`
}

type CreateInvoiceResponse struct {
	Invoice *financeinvoicesv1.InvoiceDetail `json:"invoice"`
}

func newInvoiceDetailPayload(inv *domain.Invoice) *financeinvoicesv1.InvoiceDetail {
	items := make([]*financeinvoicesv1.InvoiceItem, 0, len(inv.Items))
	for _, item := range inv.Items {
		items = append(items, &financeinvoicesv1.InvoiceItem{
			Description: item.Description,
			Quantity:    item.Quantity,
			Unit:        item.Unit,
			UnitPrice:   item.UnitPrice,
			TaxRate:     item.TaxRate,
			DiscountPct: item.DiscountPct,
			SortOrder:   int64(item.SortOrder),
			Total:       item.Total,
		})
	}

	paidDate := ""
	if inv.PaidDate != nil {
		paidDate = inv.PaidDate.Format("2006-01-02")
	}

	return &financeinvoicesv1.InvoiceDetail{
		Id:             inv.ID,
		OwnerId:        inv.OwnerID,
		InvoiceNumber:  inv.InvoiceNumber,
		ClientName:     inv.ClientName,
		ClientEmail:    inv.ClientEmail,
		ClientAddress:  inv.ClientAddress,
		ClientTaxId:    inv.ClientTaxID,
		CustomerId:     inv.CustomerID,
		Status:         inv.Status,
		IssueDate:      inv.IssueDate.Format("2006-01-02"),
		DueDate:        inv.DueDate.Format("2006-01-02"),
		PaidDate:       paidDate,
		Currency:       inv.Currency,
		Subtotal:       inv.Subtotal,
		DiscountType:   inv.DiscountType,
		DiscountValue:  inv.DiscountValue,
		DiscountAmount: inv.DiscountAmount,
		TaxRate:        inv.TaxRate,
		TaxAmount:      inv.TaxAmount,
		Total:          inv.Total,
		AmountPaid:     inv.AmountPaid,
		AmountDue:      inv.AmountDue,
		Notes:          inv.Notes,
		Terms:          inv.Terms,
		CreatedAt:      inv.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:      inv.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		Items:          items,
	}
}
