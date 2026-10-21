package api

import (
	"github.com/voxire/onion/internal/modules/finance/quotes/internal/application"
)

type UpdateQuoteByIDHandler struct{ Service *application.Service }

type UpdateQuoteByIDRequest struct {
	CustomerName  string  `json:"customer_name"`
	CustomerEmail string  `json:"customer_email"`
	Status        string  `json:"status"`
	Items         string  `json:"items"`
	Subtotal      float64 `json:"subtotal"`
	Tax           float64 `json:"tax"`
	Discount      float64 `json:"discount"`
	Total         float64 `json:"total"`
	ValidUntil    string  `json:"valid_until"`
	Notes         string  `json:"notes"`
}

type UpdateQuoteByIDResponse struct {
	Message string `json:"message"`
}
