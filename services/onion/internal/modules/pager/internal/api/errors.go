package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/pager/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	pagerv1 "github.com/voxire/voxie/gen/go/pager/v1"
)

// codedError is a failure the client reacts to by its code: a strict venue's
// ticket_not_found is retried, a venue_not_found is not.
type codedError struct {
	status  int
	code    string
	message string
}

func (e *codedError) Error() string         { return e.message }
func (e *codedError) HTTPStatusCode() int   { return e.status }
func (e *codedError) PublicMessage() string { return e.message }
func (e *codedError) PublicCode() string    { return e.code }

var (
	errVenueNotFound      = &codedError{http.StatusNotFound, "venue_not_found", "This pager link is not active"}
	errTicketNotFound     = &codedError{http.StatusNotFound, "ticket_not_found", "Ticket not found"}
	errInvalidNumber      = &codedError{http.StatusBadRequest, "invalid_number", "The order number must be 1 to 12 letters and digits"}
	errInvalidTransition  = &codedError{http.StatusConflict, "invalid_transition", "The ticket is no longer in a state that allows this"}
	errTooManyOpenTickets = &codedError{http.StatusTooManyRequests, "too_many_open_tickets", "This venue has too many open tickets right now"}
)

// publicError maps a service failure to what the client is told. Anything
// unexpected is logged and reported without detail.
func publicError(ctx context.Context, action string, err error) error {
	switch {
	case errors.Is(err, domain.ErrVenueNotFound):
		return errVenueNotFound
	case errors.Is(err, domain.ErrTicketNotFound):
		return errTicketNotFound
	case errors.Is(err, domain.ErrInvalidNumber):
		return errInvalidNumber
	case errors.Is(err, domain.ErrInvalidTransition):
		return errInvalidTransition
	case errors.Is(err, domain.ErrTooManyOpenTickets):
		return errTooManyOpenTickets
	}
	logger.Ctx(ctx).Error("pager request failed", logger.String("action", action), logger.Error(err))
	return onekiterrors.New(http.StatusInternalServerError, "Something went wrong, please try again")
}

// errorBody is the generated `{message, violations}` error body with the
// error's stable code added.
type errorBody struct {
	Message    string   `json:"message"`
	Code       string   `json:"code"`
	Violations []string `json:"violations,omitempty"`
}

// WriteError reports every failure of the pager service with its code.
func WriteError(w http.ResponseWriter, r *http.Request, e *pagerv1.ServerError) {
	status, code, message, _, violations, _ := e.Parts()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(errorBody{Message: message, Code: code, Violations: violations}); err != nil {
		logger.Ctx(r.Context()).Warn("failed to write pager error response", logger.Error(err))
	}
}
