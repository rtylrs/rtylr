import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hr.report.v1;

/** ReportService defines HR report generation endpoints. */
@route("/hr/report/v1")
interface ReportService {
  /** Generates an HR report. */
  @post
  @route("/reports/generate")
  generateReport(...GenerateReportRequest): GenerateReportResponse;
}
