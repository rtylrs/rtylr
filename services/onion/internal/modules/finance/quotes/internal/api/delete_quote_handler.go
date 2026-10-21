package api

import (
	"github.com/voxire/onion/internal/modules/finance/quotes/internal/application"
)

type DeleteQuoteByIDHandler struct{ Service *application.Service }

type DeleteQuoteByIDRequest struct{}

type DeleteQuoteByIDResponse struct {
	Message string `json:"message"`
}
