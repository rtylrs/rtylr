package api

import (
	"github.com/voxire/onion/internal/modules/finance/quotes/internal/application"
)

type UpdateQuoteStatusHandler struct{ Service *application.Service }

type UpdateQuoteStatusRequest struct {
	Status string `json:"status"`
}

type UpdateQuoteStatusResponse struct {
	Message string `json:"message"`
}
