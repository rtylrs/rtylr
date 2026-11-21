import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hr.report.v1;

/**
 * ReportRequestPayload describes the parameters for an HR report.
 * Report types: attendance | timesheet | leave_utilization | leave_calendar
 * | overtime | payroll_cost | headcount | turnover
 * | checklist_completion | performance_review_status
 */
model ReportRequestPayload {
  /** Type of report to generate (see report type constants above). */
  report_type: string;

  /** Start of the reporting period (ISO-8601 date). */
  date_from: string;

  /** End of the reporting period (ISO-8601 date). */
  date_to: string;

  /** Filter by location; absent to include all locations. */
  @encode(string)
  location_id?: int64;

  /** Filter by department; absent to include all departments. */
  @encode(string)
  department_id?: int64;

  /** Filter by employee; absent to include all employees. */
  @encode(string)
  employee_id?: int64;

  /** Output format (json | csv | pdf). */
  format: string;
}

/** ReportPayload is a generated HR report. */
model ReportPayload {
  /** Type of report that was generated. */
  report_type: string;

  /** Unix timestamp when the report was generated. */
  @encode(string)
  generated_at: int64;

  /** Serialised report content (CSV, JSON, or PDF binary depending on format). */
  data: bytes;

  /** Format of the serialised data (json | csv | pdf). */
  format: string;
}

model GenerateReportRequest {
  /** Parameters describing the report to generate. */
  @Onkt.required
  request: ReportRequestPayload;
}

model GenerateReportResponse {
  /** The generated report. */
  report: ReportPayload;
}
