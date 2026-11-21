import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;
using common.pagination.v1;

namespace insights.report.v1;

model ReportTemplate {
  /** id is the unique identifier for the resource. */
  @encode(string)
  id: int64;

  /** name is the name value for this contract. */
  name: string;

  /** description is the description value for this contract. */
  description: string;

  /**
   * type is the type value for this contract.
   * revenue, customers, products, operations, financial
   */
  type: string;

  /** metrics is the metrics value for this contract. */
  metrics: string[];

  /**
   * format is the format value for this contract.
   * pdf, csv, xlsx
   */
  format: string;

  /** created_at is when the record was created. */
  created_at: string;
}

model ScheduledReport {
  /** id is the unique identifier for the resource. */
  @encode(string)
  id: int64;

  /** template_id is the template id value for this contract. */
  @encode(string)
  template_id: int64;

  /** template_name is the template name value for this contract. */
  template_name: string;

  /**
   * frequency is the frequency value for this contract.
   * daily, weekly, monthly
   */
  frequency: string;

  /** recipient_email is the recipient email value for this contract. */
  recipient_email: string;

  /** channel is the delivery channel (email, slack). */
  channel: string;

  /** is_active is the is active value for this contract. */
  is_active: boolean;

  /** last_sent_at is the last sent at value for this contract. */
  last_sent_at: string;

  /** next_send_at is the next send at value for this contract. */
  next_send_at: string;

  /** created_at is when the record was created. */
  created_at: string;
}

model ListReportTemplatesRequest {
  /** page is the current result page. */
  @query("page")
  page: int32;

  /** limit is the maximum number of records to return. */
  @query("limit")
  limit: int32;

  /** sort_by is the field to sort results by. */
  @query("sort_by")
  sort_by: string;

  /** sort_dir is the sort direction ("asc" or "desc"). */
  @query("sort_dir")
  sort_dir: string;

  /** type is the type value for this contract. */
  @query("type")
  type: string;
}

model ListReportTemplatesResponse {
  /** templates is the templates value for this contract. */
  templates: ReportTemplate[];

  /** pagination is the pagination value for this contract. */
  pagination: PaginationResponse;
}

model CreateReportTemplateRequest {
  /** name is the name value for this contract. */
  @Onkt.required
  name: string;

  /** description is the description value for this contract. */
  @Onkt.required
  description: string;

  /** type is the type value for this contract. */
  @Onkt.required
  type: string;

  /** metrics is the metrics value for this contract. */
  @minItems(1)
  metrics: string[];

  /** format is the format value for this contract. */
  @Onkt.required
  format: string;
}

model CreateReportTemplateResponse {
  /** template is the template value for this contract. */
  template: ReportTemplate;
}

model GetReportTemplateRequest {
  /** id is the unique identifier for the resource. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model GetReportTemplateResponse {
  /** template is the template value for this contract. */
  template: ReportTemplate;
}

model ListScheduledReportsRequest {
  /** page is the current result page. */
  @query("page")
  page: int32;

  /** limit is the maximum number of records to return. */
  @query("limit")
  limit: int32;

  /** sort_by is the field to sort results by. */
  @query("sort_by")
  sort_by: string;

  /** sort_dir is the sort direction ("asc" or "desc"). */
  @query("sort_dir")
  sort_dir: string;
}

model ListScheduledReportsResponse {
  /** reports is the reports value for this contract. */
  reports: ScheduledReport[];

  /** pagination is the pagination value for this contract. */
  pagination: PaginationResponse;
}

model CreateScheduledReportRequest {
  /** template_slug is the template slug (e.g. revenue-monthly). */
  @Onkt.required
  template_slug: string;

  /** frequency is the delivery frequency (daily, weekly, monthly). */
  @Onkt.required
  frequency: string;

  /** channel is the delivery channel (email, slack). */
  @Onkt.required
  channel: string;

  /** recipient_email is the delivery recipient. */
  @Onkt.required
  @format("email")
  recipient_email: string;

  /** is_active toggles whether the schedule is active. */
  is_active: boolean;
}

model CreateScheduledReportResponse {
  /** report is the report value for this contract. */
  report: ScheduledReport;
}

model UpdateScheduledReportRequest {
  /** id is the scheduled report identifier. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  /** template_slug is the template slug (e.g. revenue-monthly). */
  @Onkt.required
  template_slug: string;

  /** frequency is the delivery frequency (daily, weekly, monthly). */
  @Onkt.required
  frequency: string;

  /** channel is the delivery channel (email, slack). */
  @Onkt.required
  channel: string;

  /** recipient_email is the delivery recipient. */
  @Onkt.required
  @format("email")
  recipient_email: string;

  /** is_active toggles whether the schedule is active. */
  is_active: boolean;
}

model UpdateScheduledReportResponse {
  /** report is the updated scheduled report. */
  report: ScheduledReport;
}

model DeleteScheduledReportRequest {
  /** id is the scheduled report identifier. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model DeleteScheduledReportResponse {
  /** deleted confirms the schedule was removed. */
  deleted: boolean;
}

model GenerateRevenueReportRequest {
  /** from_date is the inclusive start date (YYYY-MM-DD). */
  from_date: string;

  /** to_date is the inclusive end date (YYYY-MM-DD). */
  to_date: string;

  /** format is the export format (csv, pdf). */
  format: string;
}

model GenerateRevenueReportResponse {
  /** report_id is the queued report identifier. */
  report_id: string;

  /** url is the download URL when ready. */
  url: string;

  /** status is the queue status. */
  status: string;
}

model ExportRevenueReportRequest {
  /** from_date is the inclusive start date (YYYY-MM-DD). */
  @query("from")
  from_date: string;

  /** to_date is the inclusive end date (YYYY-MM-DD). */
  @query("to")
  to_date: string;
}

model ExportRevenueRow {
  /** date is the revenue date (YYYY-MM-DD). */
  date: string;

  /** amount_minor_units is the revenue amount in minor units. */
  @encode(string)
  amount_minor_units: int64;
}

model ExportRevenueReportResponse {
  /** filename is the suggested download filename. */
  filename: string;

  /** content_type is the MIME type for the export payload. */
  content_type: string;

  /** csv is the CSV document body. */
  csv: string;

  /** rows is the structured export rows. */
  rows: ExportRevenueRow[];
}
