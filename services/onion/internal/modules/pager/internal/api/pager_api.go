// Package api serves the order pager over the pager.v1 OneKit contract.
package api

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/modules/pager/internal/application"
	"github.com/voxire/onion/internal/modules/pager/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	pagerv1 "github.com/voxire/voxie/gen/go/pager/v1"
)

// PagerAPI implements pagerv1.PagerServiceServer. The public methods identify
// a venue by its slug or a ticket by its token and never return a ticket id;
// the staff methods act on the signed-in caller's business only.
type PagerAPI struct {
	Service *application.Service
}

var _ pagerv1.PagerServiceServer = (*PagerAPI)(nil)

func (a *PagerAPI) GetPublicVenue(ctx context.Context, req *pagerv1.GetPublicVenueRequest) (*pagerv1.GetPublicVenueResponse, error) {
	venue, settings, err := a.Service.PublicVenue(ctx, req.GetSlug())
	if err != nil {
		return nil, publicError(ctx, "get public venue", err)
	}
	return &pagerv1.GetPublicVenueResponse{
		Name:    venue.Name,
		LogoUrl: venue.LogoURL,
		Enabled: settings.Enabled,
		Strict:  settings.Strict,
	}, nil
}

func (a *PagerAPI) JoinTicket(ctx context.Context, req *pagerv1.JoinTicketRequest) (*pagerv1.JoinTicketResponse, error) {
	ticket, err := a.Service.Join(ctx, req.GetSlug(), req.GetNumber())
	if err != nil {
		return nil, publicError(ctx, "join ticket", err)
	}
	return &pagerv1.JoinTicketResponse{
		PublicToken: ticket.PublicToken,
		Number:      ticket.Number,
		Status:      ticket.Status,
	}, nil
}

func (a *PagerAPI) GetPublicTicket(ctx context.Context, req *pagerv1.GetPublicTicketRequest) (*pagerv1.GetPublicTicketResponse, error) {
	view, err := a.Service.PublicTicket(ctx, req.GetPublicToken())
	if err != nil {
		return nil, publicError(ctx, "get public ticket", err)
	}
	return &pagerv1.GetPublicTicketResponse{
		Number:         view.Ticket.Number,
		Status:         view.Ticket.Status,
		VenueName:      view.VenueName,
		Ahead:          int32(view.Ahead),
		ReadyAt:        timestamp(view.Ticket.ReadyAt),
		RemindedAt:     timestamp(view.Ticket.RemindedAt),
		AcknowledgedAt: timestamp(view.Ticket.AcknowledgedAt),
		ServerTime:     timestamp(view.ServerTime),
	}, nil
}

func (a *PagerAPI) AcknowledgeTicket(ctx context.Context, req *pagerv1.AcknowledgeTicketRequest) (*pagerv1.AcknowledgeTicketResponse, error) {
	if err := a.Service.Acknowledge(ctx, req.GetPublicToken()); err != nil {
		return nil, publicError(ctx, "acknowledge ticket", err)
	}
	return &pagerv1.AcknowledgeTicketResponse{}, nil
}

func (a *PagerAPI) ListTickets(ctx context.Context, _ *pagerv1.ListTicketsRequest) (*pagerv1.ListTicketsResponse, error) {
	businessID, err := staffBusinessID(ctx)
	if err != nil {
		return nil, err
	}
	tickets, readAt, err := a.Service.ListOpen(ctx, businessID)
	if err != nil {
		return nil, publicError(ctx, "list tickets", err)
	}
	out := make([]*pagerv1.Ticket, 0, len(tickets))
	for _, ticket := range tickets {
		out = append(out, toTicket(ticket))
	}
	return &pagerv1.ListTicketsResponse{Tickets: out, ServerTime: timestamp(readAt)}, nil
}

func (a *PagerAPI) CreateTicket(ctx context.Context, req *pagerv1.CreateTicketRequest) (*pagerv1.CreateTicketResponse, error) {
	businessID, err := staffBusinessID(ctx)
	if err != nil {
		return nil, err
	}
	ticket, err := a.Service.Create(ctx, businessID, req.GetNumber())
	if err != nil {
		return nil, publicError(ctx, "create ticket", err)
	}
	return &pagerv1.CreateTicketResponse{Ticket: toTicket(ticket)}, nil
}

func (a *PagerAPI) MarkReady(ctx context.Context, req *pagerv1.MarkReadyRequest) (*pagerv1.MarkReadyResponse, error) {
	ticket, err := a.move(ctx, "mark ready", req.GetId(), a.Service.MarkReady)
	if err != nil {
		return nil, err
	}
	return &pagerv1.MarkReadyResponse{Ticket: ticket}, nil
}

func (a *PagerAPI) RemindTicket(ctx context.Context, req *pagerv1.RemindTicketRequest) (*pagerv1.RemindTicketResponse, error) {
	ticket, err := a.move(ctx, "remind ticket", req.GetId(), a.Service.Remind)
	if err != nil {
		return nil, err
	}
	return &pagerv1.RemindTicketResponse{Ticket: ticket}, nil
}

func (a *PagerAPI) UndoReady(ctx context.Context, req *pagerv1.UndoReadyRequest) (*pagerv1.UndoReadyResponse, error) {
	ticket, err := a.move(ctx, "undo ready", req.GetId(), a.Service.UndoReady)
	if err != nil {
		return nil, err
	}
	return &pagerv1.UndoReadyResponse{Ticket: ticket}, nil
}

func (a *PagerAPI) CollectTicket(ctx context.Context, req *pagerv1.CollectTicketRequest) (*pagerv1.CollectTicketResponse, error) {
	ticket, err := a.move(ctx, "collect ticket", req.GetId(), a.Service.Collect)
	if err != nil {
		return nil, err
	}
	return &pagerv1.CollectTicketResponse{Ticket: ticket}, nil
}

// move runs one staff action on a ticket of the caller's business.
func (a *PagerAPI) move(ctx context.Context, action string, id int64, apply func(ctx context.Context, businessID, id int64) (domain.Ticket, error)) (*pagerv1.Ticket, error) {
	businessID, err := staffBusinessID(ctx)
	if err != nil {
		return nil, err
	}
	ticket, err := apply(ctx, businessID, id)
	if err != nil {
		return nil, publicError(ctx, action, err)
	}
	return toTicket(ticket), nil
}

func (a *PagerAPI) GetSettings(ctx context.Context, _ *pagerv1.GetSettingsRequest) (*pagerv1.GetSettingsResponse, error) {
	businessID, err := staffBusinessID(ctx)
	if err != nil {
		return nil, err
	}
	settings, err := a.Service.Settings(ctx, businessID)
	if err != nil {
		return nil, publicError(ctx, "get settings", err)
	}
	return &pagerv1.GetSettingsResponse{Enabled: settings.Enabled, Strict: settings.Strict}, nil
}

// UpdateSettings is for owners and admins: turning the pager on publishes the
// business's name and logo on an anonymous page.
func (a *PagerAPI) UpdateSettings(ctx context.Context, req *pagerv1.UpdateSettingsRequest) (*pagerv1.UpdateSettingsResponse, error) {
	businessID, err := staffBusinessID(ctx)
	if err != nil {
		return nil, err
	}
	if !authPkg.IsBusinessAdmin(authPkg.DeviceCtx(ctx).User) {
		return nil, onekiterrors.PermissionDenied()
	}
	settings, err := a.Service.UpdateSettings(ctx, businessID, req.Enabled, req.Strict)
	if err != nil {
		return nil, publicError(ctx, "update settings", err)
	}
	return &pagerv1.UpdateSettingsResponse{Enabled: settings.Enabled, Strict: settings.Strict}, nil
}

// staffBusinessID is the business of the signed-in caller, verified against
// their current membership. It is the only source of a staff request's tenant.
func staffBusinessID(ctx context.Context) (int64, error) {
	user := authPkg.DeviceCtx(ctx).User
	if user == nil {
		return 0, onekiterrors.PermissionDenied()
	}
	businessID, err := tenant.ResolveBusinessID(ctx, user.BusinessID, user.UserID)
	if err != nil {
		return 0, onekiterrors.PermissionDenied()
	}
	return businessID, nil
}

func toTicket(ticket domain.Ticket) *pagerv1.Ticket {
	return &pagerv1.Ticket{
		Id:             ticket.ID,
		Number:         ticket.Number,
		Status:         ticket.Status,
		Source:         ticket.Source,
		PhoneLinked:    ticket.PhoneLinked,
		CreatedAt:      timestamp(ticket.CreatedAt),
		ReadyAt:        timestamp(ticket.ReadyAt),
		RemindedAt:     timestamp(ticket.RemindedAt),
		AcknowledgedAt: timestamp(ticket.AcknowledgedAt),
	}
}

// timestamp formats a time as RFC 3339 UTC with milliseconds, and a time that
// has not happened as the empty string the contract leaves out.
func timestamp(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format("2006-01-02T15:04:05.000Z")
}
