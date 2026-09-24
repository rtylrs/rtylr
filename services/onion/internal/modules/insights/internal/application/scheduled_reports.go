package application

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/voxire/onion/internal/modules/insights/internal/infrastructure"
)

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

type ScheduledReportInput struct {
	TemplateID string
	Cadence    string
	Channel    string
	Recipient  string
	IsActive   *bool
}

type ScheduledReportListOptions struct {
	Page    int32
	Limit   int32
	SortBy  string
	SortDir string
}

type ScheduledReportPage struct {
	Rows  []ScheduledReportPayload
	Total int64
}

var ScheduledReportErrNotFound = infrastructure.ErrScheduledReportNotFound

var validCadences = map[string]bool{
	"daily": true, "weekly": true, "monthly": true,
}

var validChannels = map[string]bool{"email": true}

func NormalizeScheduledReportListOptions(options ScheduledReportListOptions) (ScheduledReportListOptions, error) {
	options.Page = normalizePage(options.Page)
	options.Limit = normalizeListLimit(options.Limit)
	options.SortBy = strings.ToLower(strings.TrimSpace(options.SortBy))
	switch options.SortBy {
	case "":
		options.SortBy = "created_at"
	case "createdat", "created_at", "updatedat", "updated_at", "recipient", "recipient_email", "cadence", "frequency", "channel", "is_active", "isactive":
		switch options.SortBy {
		case "createdat":
			options.SortBy = "created_at"
		case "updatedat":
			options.SortBy = "updated_at"
		case "recipient_email":
			options.SortBy = "recipient"
		case "frequency":
			options.SortBy = "cadence"
		case "isactive":
			options.SortBy = "is_active"
		}
	default:
		return ScheduledReportListOptions{}, fmt.Errorf("unsupported sort_by %q", options.SortBy)
	}
	options.SortDir = strings.ToLower(strings.TrimSpace(options.SortDir))
	switch options.SortDir {
	case "":
		options.SortDir = "desc"
	case "asc", "desc":
	default:
		return ScheduledReportListOptions{}, fmt.Errorf("unsupported sort_dir %q", options.SortDir)
	}
	return options, nil
}

func (s *Service) ListScheduledReports(ctx context.Context, businessID int64, options ScheduledReportListOptions) (ScheduledReportPage, error) {
	options, err := NormalizeScheduledReportListOptions(options)
	if err != nil {
		return ScheduledReportPage{}, err
	}
	rows, total, err := s.repo.ListScheduledReports(ctx, businessID, options.SortBy, options.SortDir, options.Limit, pageOffset(options.Page, options.Limit))
	if err != nil {
		return ScheduledReportPage{}, err
	}
	out := make([]ScheduledReportPayload, 0, len(rows))
	for _, row := range rows {
		out = append(out, toScheduledReportPayload(row))
	}
	return ScheduledReportPage{Rows: out, Total: total}, nil
}

func (s *Service) CreateScheduledReport(ctx context.Context, businessID int64, input ScheduledReportInput) (ScheduledReportPayload, error) {
	if err := validateScheduledReport(input.validationPayload()); err != nil {
		return ScheduledReportPayload{}, err
	}
	id, err := s.repo.CreateScheduledReport(ctx, infrastructure.ScheduledReportWrite{
		OwnerID: businessID, BusinessID: businessID, TemplateID: input.TemplateID,
		Cadence: input.Cadence, Channel: input.Channel, Recipient: input.Recipient, IsActive: input.IsActive,
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

func (s *Service) UpdateScheduledReport(ctx context.Context, businessID, id int64, input ScheduledReportInput) (ScheduledReportPayload, error) {
	if err := validateScheduledReport(input.validationPayload()); err != nil {
		return ScheduledReportPayload{}, err
	}
	err := s.repo.UpdateScheduledReport(ctx, infrastructure.ScheduledReportWrite{
		ID: id, OwnerID: businessID, BusinessID: businessID, TemplateID: input.TemplateID,
		Cadence: input.Cadence, Channel: input.Channel, Recipient: input.Recipient, IsActive: input.IsActive,
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

func (s *Service) DeleteScheduledReport(ctx context.Context, businessID, id int64) error {
	return s.repo.DeleteScheduledReport(ctx, businessID, id)
}

func (p ScheduledReportInput) validationPayload() ScheduledReportPayload {
	return ScheduledReportPayload{
		TemplateID: p.TemplateID,
		Cadence:    p.Cadence,
		Channel:    p.Channel,
		Recipient:  p.Recipient,
	}
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
		ID: row.ID, TemplateID: row.TemplateID, Cadence: row.Cadence, Channel: row.Channel,
		Recipient: row.Recipient, IsActive: row.IsActive,
		CreatedAt: row.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: row.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
