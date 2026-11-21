// Package application holds the pager's rules: how a scan opens a ticket, what
// a phone following one is told, and which moves the staff can make.
package application

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"
	"slices"
	"time"

	"github.com/voxire/onion/internal/modules/pager/internal/domain"
)

// repository is the pager's storage. Every ticket method takes the business,
// so a ticket of another business is simply not found. openSince is the
// creation time before which a labelled ticket no longer counts as open, and
// unlabelledSince the same for a ticket nobody labelled.
type repository interface {
	// Settings returns domain.DefaultSettings for a business that never saved any.
	Settings(ctx context.Context, businessID int64) (domain.Settings, error)
	SaveSettings(ctx context.Context, businessID int64, settings domain.Settings) error

	// CreateTicket opens an unlabelled ticket showing ticket.Code as its match
	// word. It returns domain.ErrCodeTaken when another open unlabelled ticket
	// of the business shows the word, whatever its age.
	CreateTicket(ctx context.Context, ticket domain.Ticket) (domain.Ticket, error)
	// OpenMatchCodes lists what the business's open unlabelled tickets show:
	// words, and the two character codes of tickets opened before them.
	OpenMatchCodes(ctx context.Context, businessID int64, unlabelledSince time.Time) ([]string, error)
	// ExpireStaleUnlabelled closes the unlabelled tickets older than
	// unlabelledSince, which frees their words.
	ExpireStaleUnlabelled(ctx context.Context, businessID int64, unlabelledSince time.Time) error
	// ExpireStaleTicket closes the open ticket holding number if it is older
	// than openSince, which frees the number.
	ExpireStaleTicket(ctx context.Context, businessID int64, number string, openSince time.Time) error
	ListOpenTickets(ctx context.Context, businessID int64, openSince, unlabelledSince time.Time) ([]domain.Ticket, error)
	Ticket(ctx context.Context, businessID, id int64) (domain.Ticket, bool, error)
	TicketByToken(ctx context.Context, token string) (domain.Ticket, bool, error)
	// CountWaitingAhead counts the labelled waiting tickets opened before id.
	CountWaitingAhead(ctx context.Context, businessID, id int64, openSince time.Time) (int64, error)

	// The moves below change nothing unless the ticket is open and in the
	// state the move starts from.

	// Label returns domain.ErrLabelTaken when kind is a number and another
	// open ticket of the business holds it.
	Label(ctx context.Context, businessID, id int64, label, kind string, openSince, unlabelledSince time.Time) error
	// MarkReady needs a labelled ticket.
	MarkReady(ctx context.Context, businessID, id int64, at, openSince time.Time) error
	Remind(ctx context.Context, businessID, id int64, at, openSince time.Time) error
	UndoReady(ctx context.Context, businessID, id int64, openSince time.Time) error
	Collect(ctx context.Context, businessID, id int64, openSince, unlabelledSince time.Time) error
	Acknowledge(ctx context.Context, businessID, id int64, at, openSince time.Time) error

	// SavePushSubscription returns domain.ErrTooManyPushSubscriptions once the
	// ticket holds limit subscriptions.
	SavePushSubscription(ctx context.Context, sub domain.PushSubscription, at time.Time, limit int) error
	PushSubscriptions(ctx context.Context, businessID, ticketID int64) ([]domain.PushSubscription, error)
	DeletePushSubscription(ctx context.Context, businessID, id int64) error
	DeleteTicketPushSubscriptions(ctx context.Context, businessID, ticketID int64) error
	DeleteStalePushSubscriptions(ctx context.Context, openSince time.Time) error
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

	// push is nil while lock-screen notifications are off.
	push    PushSender
	runPush func(func())
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

// matchCodeAttempts bounds the retries when scans at the same moment draw the
// same free word. There are only a few dozen words, so a rush of phones at one
// venue does collide; each retry reads the taken words again.
const matchCodeAttempts = 10

// Scan opens a waiting, unlabelled ticket for the phone that scanned the
// venue's QR, with a match word no other open unlabelled ticket shows.
func (s *Service) Scan(ctx context.Context, slug string) (domain.Ticket, error) {
	venue, _, err := s.PublicVenue(ctx, slug)
	if err != nil {
		return domain.Ticket{}, err
	}
	now := s.clock()
	unlabelledSince := domain.UnlabelledSince(now)
	// A ticket past its lifetime reads as expired but still holds its word in
	// storage until it is closed.
	if err := s.repo.ExpireStaleUnlabelled(ctx, venue.BusinessID, unlabelledSince); err != nil {
		return domain.Ticket{}, err
	}
	for range matchCodeAttempts {
		codes, err := s.repo.OpenMatchCodes(ctx, venue.BusinessID, unlabelledSince)
		if err != nil {
			return domain.Ticket{}, err
		}
		if len(codes) >= domain.MaxOpenUnlabelledTickets {
			return domain.Ticket{}, domain.ErrTooManyOpenTickets
		}
		used := make(map[string]bool, len(codes))
		for _, code := range codes {
			used[code] = true
		}
		code, err := freeMatchWord(used)
		if err != nil {
			return domain.Ticket{}, err
		}
		token, err := newPublicToken()
		if err != nil {
			return domain.Ticket{}, err
		}
		ticket, err := s.repo.CreateTicket(ctx, domain.Ticket{
			BusinessID:  venue.BusinessID,
			Code:        code,
			Status:      domain.StatusWaiting,
			PublicToken: token,
			CreatedAt:   now,
		})
		if errors.Is(err, domain.ErrCodeTaken) {
			continue
		}
		return ticket, err
	}
	return domain.Ticket{}, domain.ErrTooManyOpenTickets
}

// freeMatchWord draws a random word that is not in used: a key on its own
// while any is free, then the keys with the lowest number still free ("panda2",
// then "panda3").
func freeMatchWord(used map[string]bool) (string, error) {
	for number := 1; number <= domain.MaxOpenUnlabelledTickets; number++ {
		free := make([]string, 0, len(domain.MatchWords))
		for _, key := range domain.MatchWords {
			if code := domain.MatchWordCode(key, number); !used[code] {
				free = append(free, code)
			}
		}
		if len(free) == 0 {
			continue
		}
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(free))))
		if err != nil {
			return "", fmt.Errorf("draw pager match word: %w", err)
		}
		return free[n.Int64()], nil
	}
	return "", domain.ErrTooManyOpenTickets
}

// Label gives a waiting ticket of the business its order number or customer
// name, as the venue's label mode allows. A labelled waiting ticket can be
// labelled again to fix a typo.
func (s *Service) Label(ctx context.Context, businessID, id int64, raw string) (domain.Ticket, error) {
	settings, err := s.repo.Settings(ctx, businessID)
	if err != nil {
		return domain.Ticket{}, err
	}
	label, kind, err := domain.ParseLabel(raw, settings.LabelMode)
	if err != nil {
		return domain.Ticket{}, err
	}
	now := s.clock()
	openSince := domain.OpenSince(now)
	ticket, err := s.staffTicket(ctx, businessID, id, now)
	if err != nil {
		return domain.Ticket{}, err
	}
	if ticket.Status != domain.StatusWaiting {
		return domain.Ticket{}, domain.ErrInvalidTransition
	}
	if ticket.Label == label && ticket.LabelKind == kind {
		return ticket, nil
	}
	if kind == domain.LabelNumber {
		// Yesterday's order 42 still holds the number in storage.
		if err := s.repo.ExpireStaleTicket(ctx, businessID, label, openSince); err != nil {
			return domain.Ticket{}, err
		}
	}
	if err := s.repo.Label(ctx, businessID, id, label, kind, openSince, domain.UnlabelledSince(now)); err != nil {
		return domain.Ticket{}, err
	}
	ticket, err = s.staffTicket(ctx, businessID, id, now)
	if err != nil {
		return domain.Ticket{}, err
	}
	// Another request moved the ticket first.
	if ticket.Status != domain.StatusWaiting || ticket.Label != label || ticket.LabelKind != kind {
		return domain.Ticket{}, domain.ErrInvalidTransition
	}
	return ticket, nil
}

// PublicTicket is what a phone following a ticket is told on every poll.
type PublicTicket struct {
	// Ticket carries the status as read at ServerTime.
	Ticket    domain.Ticket
	VenueName string
	// VenueLogoURL is empty when the business has no logo.
	VenueLogoURL string
	// Ahead counts the labelled waiting tickets opened before this one.
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
	if ticket.Status == domain.StatusWaiting && ticket.Labelled() {
		ahead, err = s.repo.CountWaitingAhead(ctx, ticket.BusinessID, ticket.ID, domain.OpenSince(now))
		if err != nil {
			return PublicTicket{}, err
		}
	}
	return PublicTicket{Ticket: ticket, VenueName: venue.Name, VenueLogoURL: venue.LogoURL, Ahead: ahead, ServerTime: now}, nil
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

// ListOpen returns the business's open tickets, labelled or not, oldest first,
// and the time they were read at.
func (s *Service) ListOpen(ctx context.Context, businessID int64) ([]domain.Ticket, time.Time, error) {
	now := s.clock()
	tickets, err := s.repo.ListOpenTickets(ctx, businessID, domain.OpenSince(now), domain.UnlabelledSince(now))
	return tickets, now, err
}

func (s *Service) MarkReady(ctx context.Context, businessID, id int64) (domain.Ticket, error) {
	ticket, err := s.move(ctx, businessID, id, domain.StatusReady, []string{domain.StatusWaiting},
		func(now, openSince time.Time) error { return s.repo.MarkReady(ctx, businessID, id, now, openSince) })
	if err == nil {
		s.notify(ctx, ticket, domain.PushReady)
	}
	return ticket, err
}

func (s *Service) Remind(ctx context.Context, businessID, id int64) (domain.Ticket, error) {
	ticket, err := s.move(ctx, businessID, id, domain.StatusReady, []string{domain.StatusReady},
		func(now, openSince time.Time) error { return s.repo.Remind(ctx, businessID, id, now, openSince) })
	if err == nil {
		s.notify(ctx, ticket, domain.PushRemind)
	}
	return ticket, err
}

func (s *Service) UndoReady(ctx context.Context, businessID, id int64) (domain.Ticket, error) {
	return s.move(ctx, businessID, id, domain.StatusWaiting, []string{domain.StatusReady},
		func(_, openSince time.Time) error { return s.repo.UndoReady(ctx, businessID, id, openSince) })
}

func (s *Service) Collect(ctx context.Context, businessID, id int64) (domain.Ticket, error) {
	ticket, err := s.move(ctx, businessID, id, domain.StatusCollected, []string{domain.StatusWaiting, domain.StatusReady},
		func(now, openSince time.Time) error {
			return s.repo.Collect(ctx, businessID, id, openSince, domain.UnlabelledSince(now))
		})
	if err == nil {
		s.forgetPushes(ctx, ticket)
	}
	return ticket, err
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
	// Only a labelled ticket can be ready: the board has nothing to call it by.
	if !slices.Contains(from, ticket.Status) || to == domain.StatusReady && !ticket.Labelled() {
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

// UpdateSettings changes the settings that are given and leaves the rest.
func (s *Service) UpdateSettings(ctx context.Context, businessID int64, enabled *bool, labelMode *string) (domain.Settings, error) {
	if labelMode != nil && !domain.ValidLabelMode(*labelMode) {
		return domain.Settings{}, domain.ErrInvalidLabelMode
	}
	settings, err := s.repo.Settings(ctx, businessID)
	if err != nil {
		return domain.Settings{}, err
	}
	if enabled != nil {
		settings.Enabled = *enabled
	}
	if labelMode != nil {
		settings.LabelMode = *labelMode
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
