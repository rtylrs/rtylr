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

// codedError is a failure the client reacts to by its code, such as a
// label_taken the board explains or a venue_not_found the phone shows.
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
	errInvalidLabel       = &codedError{http.StatusBadRequest, "invalid_label", "Enter an order number of 1 to 12 letters and digits, or a name of up to 40 characters"}
	errLabelTaken         = &codedError{http.StatusConflict, "label_taken", "Another open order already has this number"}
	errInvalidLabelMode   = &codedError{http.StatusBadRequest, "invalid_label_mode", "The label mode must be number, name or either"}
	errInvalidTransition  = &codedError{http.StatusConflict, "invalid_transition", "The ticket is no longer in a state that allows this"}
	errTooManyOpenTickets = &codedError{http.StatusTooManyRequests, "too_many_open_tickets", "This venue has too many open tickets right now"}
	errInvalidPush        = &codedError{http.StatusBadRequest, "invalid_push_subscription", "This browser's push subscription is not accepted"}
	errTooManyPushes      = &codedError{http.StatusTooManyRequests, "too_many_push_subscriptions", "Too many browsers are following this ticket"}
	errJoinByNumberOff    = &codedError{http.StatusConflict, "join_by_number_off", "This venue gives you a code to show at the counter instead"}
	errInvalidReviewURL   = &codedError{http.StatusBadRequest, "invalid_review_url", "Enter a review link that starts with https://"}
	errInvalidPrepTime    = &codedError{http.StatusBadRequest, "invalid_prep_minutes", "Enter a time from 0 to 120 minutes"}
	errTypeNumber         = &codedError{http.StatusConflict, "type_your_number", "Type the order number on your receipt"}
	errInvalidRingtone    = &codedError{http.StatusBadRequest, "invalid_ringtone", "Choose bell, chime, marimba, ring or beep"}
)

// publicError maps a service failure to what the client is told. Anything
// unexpected is logged and reported without detail.
func publicError(ctx context.Context, action string, err error) error {
	switch {
	case errors.Is(err, domain.ErrVenueNotFound):
		return errVenueNotFound
	case errors.Is(err, domain.ErrTicketNotFound):
		return errTicketNotFound
	case errors.Is(err, domain.ErrInvalidLabel):
		return errInvalidLabel
	case errors.Is(err, domain.ErrLabelTaken):
		return errLabelTaken
	case errors.Is(err, domain.ErrInvalidLabelMode):
		return errInvalidLabelMode
	case errors.Is(err, domain.ErrInvalidTransition):
		return errInvalidTransition
	case errors.Is(err, domain.ErrTooManyOpenTickets):
		return errTooManyOpenTickets
	case errors.Is(err, domain.ErrInvalidPushSubscription):
		return errInvalidPush
	case errors.Is(err, domain.ErrTooManyPushSubscriptions):
		return errTooManyPushes
	case errors.Is(err, domain.ErrJoinByNumberOff):
		return errJoinByNumberOff
	case errors.Is(err, domain.ErrInvalidReviewURL):
		return errInvalidReviewURL
	case errors.Is(err, domain.ErrInvalidPrepTime):
		return errInvalidPrepTime
	case errors.Is(err, domain.ErrTypeNumber):
		return errTypeNumber
	case errors.Is(err, domain.ErrInvalidRingtone):
		return errInvalidRingtone
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
