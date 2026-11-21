import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace insights.report.v1;

/** ReportService defines insights report template endpoints. */
@route("/insights/report/v1")
interface ReportService {
  /** Returns saved report templates. */
  @get
  @route("/templates")
  listReportTemplates(
    ...ListReportTemplatesRequest,
  ): ListReportTemplatesResponse;

  /** Creates a report template. */
  @post
  @route("/templates")
  createReportTemplate(
    ...CreateReportTemplateRequest,
  ): CreateReportTemplateResponse;

  /** Returns a report template by ID. */
  @get
  @route("/templates/{id}")
  getReportTemplate(...GetReportTemplateRequest): GetReportTemplateResponse;

  /** Returns scheduled reports. */
  @get
  @route("/scheduled")
  listScheduledReports(
    ...ListScheduledReportsRequest,
  ): ListScheduledReportsResponse;

  /** Creates a scheduled report. */
  @post
  @route("/scheduled")
  createScheduledReport(
    ...CreateScheduledReportRequest,
  ): CreateScheduledReportResponse;

  /** Updates a scheduled report. */
  @put
  @route("/scheduled/{id}")
  updateScheduledReport(
    ...UpdateScheduledReportRequest,
  ): UpdateScheduledReportResponse;

  /** Removes a scheduled report. */
  @delete
  @route("/scheduled/{id}")
  deleteScheduledReport(
    ...DeleteScheduledReportRequest,
  ): DeleteScheduledReportResponse;

  /** Queues a revenue report export. */
  @post
  @route("/reports/revenue/generate")
  generateRevenueReport(
    ...GenerateRevenueReportRequest,
  ): GenerateRevenueReportResponse;

  /** Returns a CSV export for a revenue date range. */
  @get
  @route("/reports/revenue/export")
  exportRevenueReport(
    ...ExportRevenueReportRequest,
  ): ExportRevenueReportResponse;
}
