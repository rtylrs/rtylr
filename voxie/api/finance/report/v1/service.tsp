import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace finance.report.v1;

/** ReportService defines finance reporting endpoints. */
@route("/finance/report/v1")
interface ReportService {
  @get
  @route("/pl")
  getPLReport(...GetPLReportRequest): GetPLReportResponse;

  @get
  @route("/tax")
  getTaxReport(...GetTaxReportRequest): GetTaxReportResponse;

  @get
  @route("/expense-categories")
  listExpenseCategoryDetails(
    ...ListExpenseCategoryDetailsRequest,
  ): ListExpenseCategoryDetailsResponse;

  @post
  @route("/expense-categories")
  createExpenseCategoryDetail(
    ...CreateExpenseCategoryDetailRequest,
  ): CreateExpenseCategoryDetailResponse;

  @get
  @route("/tax-config")
  listTaxConfig(...ListTaxConfigRequest): ListTaxConfigResponse;

  @post
  @route("/tax-config")
  createTaxConfig(...CreateTaxConfigRequest): CreateTaxConfigResponse;
}
