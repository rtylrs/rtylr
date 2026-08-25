package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/voxire/onion/internal/modules/insights/internal/infrastructure"
)

// ScheduledReportPayload is the JSON shape returned from every scheduled-
// report endpoint. Timestamps are emitted in RFC3339 so the frontend can
// parse them with native Date without ambiguity.
type ScheduledReportPayload struct {
	ID         int64  `json:"id"`
	TemplateID string `json:"template_id"`
	Cadence    string `json:"cadence"`
	Channel    string `json:"channel"`
	Recipient  string `json:"recipient"`
	IsActive   bool   `json:"is_active"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

// ScheduledReportErrNotFound is re-exported so handlers can 404 cleanly
// without depending on the infrastructure package.
var ScheduledReportErrNotFound = infrastructure.ErrScheduledReportNotFound

// validCadences/validChannels enforce a small, product-level whitelist.
// These are not per-tenant configurable — they're the set the product
// supports. A new cadence requires a schema change anyway (worker cron).
var validCadences = map[string]bool{
	"daily":   true,
	"weekly":  true,
	"monthly": true,
}

var validChannels = map[string]bool{
	"email": true,
}

// ListScheduledReports returns all schedules for the owner.
func (s *Service) ListScheduledReports(ctx context.Context, businessID int64) ([]ScheduledReportPayload, error) {
	rows, err := s.repo.ListScheduledReports(ctx, businessID)
	if err != nil {
		return nil, err
	}
	out := make([]ScheduledReportPayload, 0, len(rows))
	for _, row := range rows {
		out = append(out, toScheduledReportPayload(row))
	}
	return out, nil
}

// CreateScheduledReport validates input and persists a new schedule.
func (s *Service) CreateScheduledReport(ctx context.Context, businessID int64, input ScheduledReportPayload) (ScheduledReportPayload, error) {
	if err := validateScheduledReport(input); err != nil {
		return ScheduledReportPayload{}, err
	}
	id, err := s.repo.CreateScheduledReport(ctx, infrastructure.ScheduledReportRow{
		OwnerID:    businessID,
		BusinessID: businessID,
		TemplateID: input.TemplateID,
		Cadence:    input.Cadence,
		Channel:    input.Channel,
		Recipient:  input.Recipient,
		IsActive:   input.IsActive,
	})
	if err != nil {
		return ScheduledReportPayload{}, err
	}
	row, err := s.repo.GetScheduledReport(ctx, businessID, id)
	if err != nil {
		return ScheduledReportPayload{}, err
	}
	return toScheduledReportPayload(row), nil
}

// UpdateScheduledReport validates input and applies the update.
func (s *Service) UpdateScheduledReport(ctx context.Context, businessID, id int64, input ScheduledReportPayload) (ScheduledReportPayload, error) {
	if err := validateScheduledReport(input); err != nil {
		return ScheduledReportPayload{}, err
	}
	uerr := s.repo.UpdateScheduledReport(ctx, infrastructure.ScheduledReportRow{
		ID:         id,
		OwnerID:    businessID,
		BusinessID: businessID,
		TemplateID: input.TemplateID,
		Cadence:    input.Cadence,
		Channel:    input.Channel,
		Recipient:  input.Recipient,
		IsActive:   input.IsActive,
	})
	if uerr != nil {
		return ScheduledReportPayload{}, uerr
	}
	row, err := s.repo.GetScheduledReport(ctx, businessID, id)
	if err != nil {
		return ScheduledReportPayload{}, err
	}
	return toScheduledReportPayload(row), nil
}

// DeleteScheduledReport removes a schedule by id.
func (s *Service) DeleteScheduledReport(ctx context.Context, businessID, id int64) error {
	return s.repo.DeleteScheduledReport(ctx, businessID, id)
}

func validateScheduledReport(p ScheduledReportPayload) error {
	if p.TemplateID == "" {
		return errors.New("template_id is required")
	}
	if !validCadences[p.Cadence] {
		return fmt.Errorf("invalid cadence %q", p.Cadence)
	}
	if !validChannels[p.Channel] {
		return fmt.Errorf("invalid channel %q", p.Channel)
	}
	if p.Recipient == "" {
		return errors.New("recipient is required")
	}
	return nil
}

func toScheduledReportPayload(row infrastructure.ScheduledReportRow) ScheduledReportPayload {
	return ScheduledReportPayload{
		ID:         row.ID,
		TemplateID: row.TemplateID,
		Cadence:    row.Cadence,
		Channel:    row.Channel,
		Recipient:  row.Recipient,
		IsActive:   row.IsActive,
		CreatedAt:  row.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:  row.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
