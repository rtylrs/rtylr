// Package domain holds the pager's tickets and the rules that depend on
// neither storage nor transport.
package domain

import (
	"errors"
	"strings"
	"time"
)

const (
	StatusWaiting   = "waiting"
	StatusReady     = "ready"
	StatusCollected = "collected"
	StatusExpired   = "expired"

	SourceStaff    = "staff"
	SourceCustomer = "customer"
)

// OpenLifetime is how long a ticket stays open. Receipt numbers repeat every
// day, so an older open ticket reads as expired and its number is free again.
const OpenLifetime = 6 * time.Hour

// MaxOpenCustomerTickets bounds what anonymous joins can create at one venue.
// Tickets the staff enter are not counted and are never refused.
const MaxOpenCustomerTickets = 2000

const maxNumberLength = 12

var (
	ErrVenueNotFound      = errors.New("pager venue not found")
	ErrTicketNotFound     = errors.New("pager ticket not found")
	ErrInvalidNumber      = errors.New("pager number is invalid")
	ErrInvalidTransition  = errors.New("pager ticket is not in a state that allows this")
	ErrTooManyOpenTickets = errors.New("pager venue has too many open tickets")
	// ErrNumberTaken is the storage layer reporting that another open ticket
	// already holds the number.
	ErrNumberTaken = errors.New("pager number already has an open ticket")
)

// Ticket is one order number being waited for. The time fields are zero until
// the moment they describe has happened.
type Ticket struct {
	ID             int64
	BusinessID     int64
	Number         string
	Status         string
	Source         string
	PublicToken    string
	PhoneLinked    bool
	CreatedAt      time.Time
	ReadyAt        time.Time
	RemindedAt     time.Time
	AcknowledgedAt time.Time
}

// OpenSince is the oldest creation time a ticket can have and still be open.
func OpenSince(now time.Time) time.Time {
	return now.Add(-OpenLifetime)
}

// StatusAt is the status as read at now: an open ticket past its lifetime is
// expired whatever the stored status says.
func (t Ticket) StatusAt(now time.Time) string {
	open := t.Status == StatusWaiting || t.Status == StatusReady
	if open && t.CreatedAt.Before(OpenSince(now)) {
		return StatusExpired
	}
	return t.Status
}

type Settings struct {
	// Enabled makes the venue reachable on the public pager routes.
	Enabled bool
	// Strict lets customers join only numbers the staff already entered.
	Strict bool
}

// Venue is the business a pager link points at.
type Venue struct {
	BusinessID int64
	Name       string
	LogoURL    string
}

// NormalizeNumber turns what a person typed from a receipt into the stored
// form: trimmed, Arabic-Indic and Persian digits as ASCII, letters in capitals.
// Anything other than 1 to 12 ASCII letters and digits is refused.
func NormalizeNumber(raw string) (string, error) {
	var out strings.Builder
	for _, r := range strings.TrimSpace(raw) {
		switch {
		case r >= '0' && r <= '9', r >= 'A' && r <= 'Z':
		case r >= 'a' && r <= 'z':
			r -= 'a' - 'A'
		case r >= '٠' && r <= '٩':
			r = '0' + (r - '٠')
		case r >= '۰' && r <= '۹':
			r = '0' + (r - '۰')
		default:
			return "", ErrInvalidNumber
		}
		if out.Len() == maxNumberLength {
			return "", ErrInvalidNumber
		}
		out.WriteRune(r)
	}
	if out.Len() == 0 {
		return "", ErrInvalidNumber
	}
	return out.String(), nil
}
