package api

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/voxire/onion/internal/modules/insights/internal/application"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	insightsreportv1 "github.com/voxire/voxie/gen/go/insights/report/v1"
)

type ReportAPI struct {
	Service *application.Service
}

func NewReportAPI(service *application.Service) *ReportAPI {
	return &ReportAPI{Service: service}
}

func scheduledReportToProto(row application.ScheduledReportPayload) *insightsreportv1.ScheduledReport {
	// TemplateID is stored as a slug (e.g. "revenue"); non-numeric slugs map to
	// a zero numeric id and the slug itself is carried in TemplateName below.
	var templateID int64
	if id, err := strconv.ParseInt(row.TemplateID, 10, 64); err == nil {
		templateID = id
	}
	return &insightsreportv1.ScheduledReport{
		Id:             row.ID,
		TemplateId:     templateID,
		TemplateName:   row.TemplateID,
		Frequency:      row.Cadence,
		Channel:        row.Channel,
		RecipientEmail: row.Recipient,
		IsActive:       row.IsActive,
		CreatedAt:      row.CreatedAt,
	}
}

func scheduledReportInput(reqTemplateSlug, frequency, channel, recipient string, isActive *bool) application.ScheduledReportInput {
	return application.ScheduledReportInput{
		TemplateID: reqTemplateSlug,
		Cadence:    frequency,
		Channel:    channel,
		Recipient:  recipient,
		IsActive:   isActive,
	}
}

func (a *ReportAPI) ListReportTemplates(ctx context.Context, req *insightsreportv1.ListReportTemplatesRequest) (*insightsreportv1.ListReportTemplatesResponse, error) {
	_, err := resolveBusinessIDFor(ctx, "insights:reports:create")
	if err != nil {
		return nil, err
	}
	page, err := a.Service.ListReportTemplates(ctx, application.ReportTemplateListOptions{
		Page: req.GetPage(), Limit: req.GetLimit(), SortBy: req.GetSortBy(), SortDir: req.GetSortDir(), Type: req.GetType(),
	})
	if err != nil {
		return nil, onekiterrors.New(http.StatusBadRequest, err.Error())
	}

	templates := make([]*insightsreportv1.ReportTemplate, 0, len(page.Rows))
	for i, row := range page.Rows {
		templates = append(templates, &insightsreportv1.ReportTemplate{
			Id:          int64(i + 1),
			Name:        row.Name,
			Description: row.Description,
			Type:        row.Type,
		})
	}
	limit := normalizeLimit(req.GetLimit())
	return &insightsreportv1.ListReportTemplatesResponse{
		Templates:  templates,
		Pagination: paginationResponse(req.GetPage(), limit, page.Total),
	}, nil
}

func (a *ReportAPI) CreateReportTemplate(ctx context.Context, req *insightsreportv1.CreateReportTemplateRequest) (*insightsreportv1.CreateReportTemplateResponse, error) {
	if _, err := resolveBusinessIDFor(ctx, "insights:reports:create"); err != nil {
		return nil, err
	}
	if req == nil || strings.TrimSpace(req.GetName()) == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "report template name is required")
	}

	// Insights templates are a product-owned catalog, not tenant-managed
	// records. The create operation returns the requested template shape for
	// clients that use the common report-service contract while the durable
	// tenant-owned operation remains CreateScheduledReport.
	return &insightsreportv1.CreateReportTemplateResponse{Template: &insightsreportv1.ReportTemplate{
		Name:        req.GetName(),
		Description: req.GetDescription(),
		Type:        req.GetType(),
		Metrics:     req.GetMetrics(),
		Format:      req.GetFormat(),
	}}, nil
}

func (a *ReportAPI) GetReportTemplate(context.Context, *insightsreportv1.GetReportTemplateRequest) (*insightsreportv1.GetReportTemplateResponse, error) {
	return nil, notImplemented("GetReportTemplate")
}

func (a *ReportAPI) ListScheduledReports(ctx context.Context, req *insightsreportv1.ListScheduledReportsRequest) (*insightsreportv1.ListScheduledReportsResponse, error) {
	businessID, err := resolveBusinessIDFor(ctx, "insights:reports:create")
	if err != nil {
		return nil, err
	}
	options, err := application.NormalizeScheduledReportListOptions(application.ScheduledReportListOptions{
		Page: req.GetPage(), Limit: req.GetLimit(), SortBy: req.GetSortBy(), SortDir: req.GetSortDir(),
	})
	if err != nil {
		return nil, onekiterrors.New(http.StatusBadRequest, err.Error())
	}
	page, err := a.Service.ListScheduledReports(ctx, businessID, options)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list scheduled reports")
	}

	reports := make([]*insightsreportv1.ScheduledReport, 0, len(page.Rows))
	for _, row := range page.Rows {
		reports = append(reports, scheduledReportToProto(row))
	}
	limit := normalizeLimit(req.GetLimit())
	return &insightsreportv1.ListScheduledReportsResponse{
		Reports:    reports,
		Pagination: paginationResponse(req.GetPage(), limit, page.Total),
	}, nil
}

func (a *ReportAPI) CreateScheduledReport(ctx context.Context, req *insightsreportv1.CreateScheduledReportRequest) (*insightsreportv1.CreateScheduledReportResponse, error) {
	businessID, err := resolveBusinessIDFor(ctx, "insights:reports:create")
	if err != nil {
		return nil, err
	}

	out, err := a.Service.CreateScheduledReport(ctx, businessID, scheduledReportInput(
		req.GetTemplateSlug(),
		req.GetFrequency(),
		req.GetChannel(),
		req.GetRecipientEmail(),
		scheduledReportIsActive(ctx, req.GetIsActive()),
	))
	if err != nil {
		return nil, onekiterrors.New(http.StatusBadRequest, err.Error())
	}

	return &insightsreportv1.CreateScheduledReportResponse{Report: scheduledReportToProto(out)}, nil
}

func (a *ReportAPI) UpdateScheduledReport(ctx context.Context, req *insightsreportv1.UpdateScheduledReportRequest) (*insightsreportv1.UpdateScheduledReportResponse, error) {
	businessID, err := resolveBusinessIDFor(ctx, "insights:reports:share")
	if err != nil {
		return nil, err
	}

	out, err := a.Service.UpdateScheduledReport(ctx, businessID, req.GetId(), scheduledReportInput(
		req.GetTemplateSlug(),
		req.GetFrequency(),
		req.GetChannel(),
		req.GetRecipientEmail(),
		scheduledReportIsActive(ctx, req.GetIsActive()),
	))
	if err != nil {
		return nil, onekiterrors.New(http.StatusBadRequest, err.Error())
	}

	return &insightsreportv1.UpdateScheduledReportResponse{Report: scheduledReportToProto(out)}, nil
}

func (a *ReportAPI) DeleteScheduledReport(ctx context.Context, req *insightsreportv1.DeleteScheduledReportRequest) (*insightsreportv1.DeleteScheduledReportResponse, error) {
	businessID, err := resolveBusinessIDFor(ctx, "insights:reports:create")
	if err != nil {
		return nil, err
	}

	if err := a.Service.DeleteScheduledReport(ctx, businessID, req.GetId()); err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to delete scheduled report")
	}

	return &insightsreportv1.DeleteScheduledReportResponse{Deleted: true}, nil
}

func (a *ReportAPI) GenerateRevenueReport(ctx context.Context, _ *insightsreportv1.GenerateRevenueReportRequest) (*insightsreportv1.GenerateRevenueReportResponse, error) {
	businessID, err := resolveBusinessIDFor(ctx, "insights:reports:create")
	if err != nil {
		return nil, err
	}

	result := a.Service.QueueRevenueReport(ctx, businessID)
	var url string
	if result.URL != nil {
		url = *result.URL
	}

	return &insightsreportv1.GenerateRevenueReportResponse{
		ReportId: result.ReportID,
		Url:      url,
		Status:   result.Status,
	}, nil
}

func (a *ReportAPI) ExportRevenueReport(ctx context.Context, req *insightsreportv1.ExportRevenueReportRequest) (*insightsreportv1.ExportRevenueReportResponse, error) {
	businessID, err := resolveBusinessIDFor(ctx, "insights:reports:export")
	if err != nil {
		return nil, err
	}

	from, to, err := application.ResolveDateRange(strings.TrimSpace(req.GetFromDate()), strings.TrimSpace(req.GetToDate()))
	if err != nil {
		return nil, onekiterrors.New(http.StatusBadRequest, "Invalid date range")
	}

	series, err := a.Service.GetRevenueSeries(ctx, businessID, from, to)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to load revenue export")
	}

	rows := make([]*insightsreportv1.ExportRevenueRow, 0, len(series))
	var b strings.Builder
	b.WriteString("date,amount_minor_units\n")
	for _, point := range series {
		rows = append(rows, &insightsreportv1.ExportRevenueRow{
			Date:             point.Date,
			AmountMinorUnits: point.AmountMinorUnits,
		})
		b.WriteString(point.Date)
		b.WriteByte(',')
		b.WriteString(strconv.FormatInt(point.AmountMinorUnits, 10))
		b.WriteByte('\n')
	}

	filename := fmt.Sprintf("revenue_%s_%s.csv", from.Format("2006-01-02"), to.Format("2006-01-02"))
	return &insightsreportv1.ExportRevenueReportResponse{
		Filename:    filename,
		ContentType: "text/csv",
		Csv:         b.String(),
		Rows:        rows,
	}, nil
}
