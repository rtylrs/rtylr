package api

import (
	"github.com/voxire/onion/internal/modules/finance/quotes/internal/application"
)

type ConvertQuoteHandler struct{ Service *application.Service }

type ConvertQuoteRequest struct{}

type ConvertQuoteResponse struct {
	InvoiceID int64  `json:"invoice_id"`
	Message   string `json:"message"`
}
