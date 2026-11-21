// Package application holds the pager's rules: who may open a ticket, what a
// phone following one is told, and which moves the staff can make.
package application

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/voxire/onion/internal/modules/pager/internal/domain"
)

// repository is the pager's storage. Every ticket method takes the business,
// so a ticket of another business is simply not found. openSince is the
// creation time before which an open ticket no longer counts as open.
type repository interface {
	// Settings returns the zero value for a business that never saved any.
	Settings(ctx context.Context, businessID int64) (domain.Settings, error)
	SaveSettings(ctx context.Context, businessID int64, settings domain.Settings) error

	// CreateTicket returns domain.ErrNumberTaken when the business already has
	// an open ticket with the number, whatever its age.
	CreateTicket(ctx context.Context, ticket domain.Ticket) (domain.Ticket, error)
	OpenTicketByNumber(ctx context.Context, businessID int64, number string, openSince time.Time) (domain.Ticket, bool, error)
	// ExpireStaleTicket closes the open ticket holding number if it is older
	// than openSince, which frees the number.
	ExpireStaleTicket(ctx context.Context, businessID int64, number string, openSince time.Time) error
	CountOpenCustomerTickets(ctx context.Context, businessID int64, openSince time.Time) (int64, error)
	ListOpenTickets(ctx context.Context, businessID int64, openSince time.Time) ([]domain.Ticket, error)
	Ticket(ctx context.Context, businessID, id int64) (domain.Ticket, bool, error)
	TicketByToken(ctx context.Context, token string) (domain.Ticket, bool, error)
	CountWaitingAhead(ctx context.Context, businessID, id int64, openSince time.Time) (int64, error)
	LinkPhone(ctx context.Context, businessID, id int64) error

	// The moves below change nothing unless the ticket is open and in the
	// state the move starts from.
	MarkReady(ctx context.Context, businessID, id int64, at, openSince time.Time) error
	Remind(ctx context.Context, businessID, id int64, at, openSince time.Time) error
	UndoReady(ctx context.Context, businessID, id int64, openSince time.Time) error
	Collect(ctx context.Context, businessID, id int64, openSince time.Time) error
	Acknowledge(ctx context.Context, businessID, id int64, at, openSince time.Time) error
}

// venueReader resolves the business behind a pager link. Both methods return
// domain.ErrVenueNotFound for an unknown or inactive business.
type venueReader interface {
	VenueBySlug(ctx context.Context, slug string) (domain.Venue, error)
	VenueByID(ctx context.Context, businessID int64) (domain.Venue, error)
}

type Service struct {
	repo   repository
	venues venueReader
	now    func() time.Time
}

func NewService(repo repository, venues venueReader, now func() time.Time) *Service {
	return &Service{repo: repo, venues: venues, now: now}
}

// clock is the current time at the precision the tickets are stored with.
func (s *Service) clock() time.Time {
	return s.now().UTC().Truncate(time.Millisecond)
}

// PublicVenue resolves a pager link. A venue whose pager is off is not found.
func (s *Service) PublicVenue(ctx context.Context, slug string) (domain.Venue, domain.Settings, error) {
	venue, err := s.venues.VenueBySlug(ctx, slug)
	if err != nil {
		return domain.Venue{}, domain.Settings{}, err
	}
	settings, err := s.repo.Settings(ctx, venue.BusinessID)
	if err != nil {
		return domain.Venue{}, domain.Settings{}, err
	}
	if !settings.Enabled {
		return domain.Venue{}, domain.Settings{}, domain.ErrVenueNotFound
	}
	return venue, settings, nil
}

// Join links a customer's phone to the open ticket with their number. A venue
// that is not strict creates the ticket when the staff have not entered it yet.
func (s *Service) Join(ctx context.Context, slug, rawNumber string) (domain.Ticket, error) {
	number, err := domain.NormalizeNumber(rawNumber)
	if err != nil {
		return domain.Ticket{}, err
	}
	venue, settings, err := s.PublicVenue(ctx, slug)
	if err != nil {
		return domain.Ticket{}, err
	}
	return s.open(ctx, venue.BusinessID, number, domain.SourceCustomer, !settings.Strict)
}

// Create enters a number from the staff board, returning the ticket that
// already holds it when there is one.
func (s *Service) Create(ctx context.Context, businessID int64, rawNumber string) (domain.Ticket, error) {
	number, err := domain.NormalizeNumber(rawNumber)
	if err != nil {
		return domain.Ticket{}, err
	}
	return s.open(ctx, businessID, number, domain.SourceStaff, true)
}

// open returns the open ticket for number, creating it when create allows.
// Storage holds the rule that a number has one open ticket, so when the
// cashier and the customer enter it at the same moment the loser of that race
// attaches to the winner's ticket.
func (s *Service) open(ctx context.Context, businessID int64, number, source string, create bool) (domain.Ticket, error) {
	now := s.clock()
	openSince := domain.OpenSince(now)
	fromCustomer := source == domain.SourceCustomer

	for range 2 {
		ticket, found, err := s.repo.OpenTicketByNumber(ctx, businessID, number, openSince)
		if err != nil {
			return domain.Ticket{}, err
		}
		if found {
			if fromCustomer && !ticket.PhoneLinked {
				if err := s.repo.LinkPhone(ctx, businessID, ticket.ID); err != nil {
					return domain.Ticket{}, err
				}
				ticket.PhoneLinked = true
			}
			return ticket, nil
		}
		if !create {
			return domain.Ticket{}, domain.ErrTicketNotFound
		}
		if fromCustomer {
			count, err := s.repo.CountOpenCustomerTickets(ctx, businessID, openSince)
			if err != nil {
				return domain.Ticket{}, err
			}
			if count >= domain.MaxOpenCustomerTickets {
				return domain.Ticket{}, domain.ErrTooManyOpenTickets
			}
		}
		// A ticket past its lifetime reads as expired but still holds the
		// number in storage until it is closed.
		if err := s.repo.ExpireStaleTicket(ctx, businessID, number, openSince); err != nil {
			return domain.Ticket{}, err
		}
		token, err := newPublicToken()
		if err != nil {
			return domain.Ticket{}, err
		}
		ticket, err = s.repo.CreateTicket(ctx, domain.Ticket{
			BusinessID:  businessID,
			Number:      number,
			Status:      domain.StatusWaiting,
			Source:      source,
			PublicToken: token,
			PhoneLinked: fromCustomer,
			CreatedAt:   now,
		})
		if errors.Is(err, domain.ErrNumberTaken) {
			continue
		}
		return ticket, err
	}
	return domain.Ticket{}, domain.ErrNumberTaken
}

// PublicTicket is what a phone following a ticket is told on every poll.
type PublicTicket struct {
	// Ticket carries the status as read at ServerTime.
	Ticket    domain.Ticket
	VenueName string
	// Ahead counts the waiting tickets entered before this one.
	Ahead      int64
	ServerTime time.Time
}

func (s *Service) PublicTicket(ctx context.Context, token string) (PublicTicket, error) {
	now := s.clock()
	ticket, err := s.ticketByToken(ctx, token)
	if err != nil {
		return PublicTicket{}, err
	}
	venue, err := s.venues.VenueByID(ctx, ticket.BusinessID)
	if errors.Is(err, domain.ErrVenueNotFound) {
		return PublicTicket{}, domain.ErrTicketNotFound
	}
	if err != nil {
		return PublicTicket{}, err
	}
	ticket.Status = ticket.StatusAt(now)
	var ahead int64
	if ticket.Status == domain.StatusWaiting {
		ahead, err = s.repo.CountWaitingAhead(ctx, ticket.BusinessID, ticket.ID, domain.OpenSince(now))
		if err != nil {
			return PublicTicket{}, err
		}
	}
	return PublicTicket{Ticket: ticket, VenueName: venue.Name, Ahead: ahead, ServerTime: now}, nil
}

// Acknowledge records that a customer silenced the bell. It leaves the status
// alone, and a ticket that is not ringing is left as it is.
func (s *Service) Acknowledge(ctx context.Context, token string) error {
	now := s.clock()
	ticket, err := s.ticketByToken(ctx, token)
	if err != nil {
		return err
	}
	if ticket.StatusAt(now) != domain.StatusReady || !ticket.AcknowledgedAt.IsZero() {
		return nil
	}
	return s.repo.Acknowledge(ctx, ticket.BusinessID, ticket.ID, now, domain.OpenSince(now))
}

// ticketByToken resolves a public token while the venue's pager is on.
func (s *Service) ticketByToken(ctx context.Context, token string) (domain.Ticket, error) {
	if !validPublicToken(token) {
		return domain.Ticket{}, domain.ErrTicketNotFound
	}
	ticket, found, err := s.repo.TicketByToken(ctx, token)
	if err != nil {
		return domain.Ticket{}, err
	}
	if !found {
		return domain.Ticket{}, domain.ErrTicketNotFound
	}
	settings, err := s.repo.Settings(ctx, ticket.BusinessID)
	if err != nil {
		return domain.Ticket{}, err
	}
	if !settings.Enabled {
		return domain.Ticket{}, domain.ErrTicketNotFound
	}
	return ticket, nil
}

// ListOpen returns the business's open tickets, oldest first, and the time
// they were read at.
func (s *Service) ListOpen(ctx context.Context, businessID int64) ([]domain.Ticket, time.Time, error) {
	now := s.clock()
	tickets, err := s.repo.ListOpenTickets(ctx, businessID, domain.OpenSince(now))
	return tickets, now, err
}

func (s *Service) MarkReady(ctx context.Context, businessID, id int64) (domain.Ticket, error) {
	return s.move(ctx, businessID, id, domain.StatusReady, []string{domain.StatusWaiting},
		func(now, openSince time.Time) error { return s.repo.MarkReady(ctx, businessID, id, now, openSince) })
}

func (s *Service) Remind(ctx context.Context, businessID, id int64) (domain.Ticket, error) {
	return s.move(ctx, businessID, id, domain.StatusReady, []string{domain.StatusReady},
		func(now, openSince time.Time) error { return s.repo.Remind(ctx, businessID, id, now, openSince) })
}

func (s *Service) UndoReady(ctx context.Context, businessID, id int64) (domain.Ticket, error) {
	return s.move(ctx, businessID, id, domain.StatusWaiting, []string{domain.StatusReady},
		func(_, openSince time.Time) error { return s.repo.UndoReady(ctx, businessID, id, openSince) })
}

func (s *Service) Collect(ctx context.Context, businessID, id int64) (domain.Ticket, error) {
	return s.move(ctx, businessID, id, domain.StatusCollected, []string{domain.StatusWaiting, domain.StatusReady},
		func(_, openSince time.Time) error { return s.repo.Collect(ctx, businessID, id, openSince) })
}

// move takes a ticket of the business from one of the from states to to. The
// ticket is read again afterwards because apply changes nothing when another
// request moved the ticket first.
func (s *Service) move(ctx context.Context, businessID, id int64, to string, from []string, apply func(now, openSince time.Time) error) (domain.Ticket, error) {
	now := s.clock()
	ticket, err := s.staffTicket(ctx, businessID, id, now)
	if err != nil {
		return domain.Ticket{}, err
	}
	if !slices.Contains(from, ticket.Status) {
		return domain.Ticket{}, domain.ErrInvalidTransition
	}
	if err := apply(now, domain.OpenSince(now)); err != nil {
		return domain.Ticket{}, err
	}
	ticket, err = s.staffTicket(ctx, businessID, id, now)
	if err != nil {
		return domain.Ticket{}, err
	}
	if ticket.Status != to {
		return domain.Ticket{}, domain.ErrInvalidTransition
	}
	return ticket, nil
}

// staffTicket reads a ticket of the business with its status as read at now.
func (s *Service) staffTicket(ctx context.Context, businessID, id int64, now time.Time) (domain.Ticket, error) {
	ticket, found, err := s.repo.Ticket(ctx, businessID, id)
	if err != nil {
		return domain.Ticket{}, err
	}
	if !found {
		return domain.Ticket{}, domain.ErrTicketNotFound
	}
	ticket.Status = ticket.StatusAt(now)
	return ticket, nil
}

func (s *Service) Settings(ctx context.Context, businessID int64) (domain.Settings, error) {
	return s.repo.Settings(ctx, businessID)
}

// UpdateSettings changes the switches that are given and leaves the rest.
func (s *Service) UpdateSettings(ctx context.Context, businessID int64, enabled, strict *bool) (domain.Settings, error) {
	settings, err := s.repo.Settings(ctx, businessID)
	if err != nil {
		return domain.Settings{}, err
	}
	if enabled != nil {
		settings.Enabled = *enabled
	}
	if strict != nil {
		settings.Strict = *strict
	}
	if err := s.repo.SaveSettings(ctx, businessID, settings); err != nil {
		return domain.Settings{}, err
	}
	return settings, nil
}

// publicTokenBytes of randomness make a ticket's token: 192 bits, well past
// guessing, and the only credential a phone holds.
const publicTokenBytes = 24

func newPublicToken() (string, error) {
	raw := make([]byte, publicTokenBytes)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("generate pager token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

// validPublicToken rejects anything that cannot be a token before storage is
// asked about it.
func validPublicToken(token string) bool {
	if len(token) != base64.RawURLEncoding.EncodedLen(publicTokenBytes) {
		return false
	}
	for i := range len(token) {
		c := token[i]
		letter := c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
		if !letter && (c < '0' || c > '9') && c != '-' && c != '_' {
			return false
		}
	}
	return true
}
