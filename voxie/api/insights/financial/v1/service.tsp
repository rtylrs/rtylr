import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace insights.financial.v1;

/** FinancialService defines insights financial analytics endpoints. */
@route("/insights/financial/v1")
interface FinancialService {
  /** Returns cash flow analytics for a period. */
  @get
  @route("/cash-flow")
  getCashFlowAnalytics(
    ...GetCashFlowAnalyticsRequest,
  ): GetCashFlowAnalyticsResponse;

  /** Returns expense analytics for a period. */
  @get
  @route("/expenses")
  getExpenseAnalytics(
    ...GetExpenseAnalyticsRequest,
  ): GetExpenseAnalyticsResponse;

  /** Returns invoice aging analysis. */
  @get
  @route("/invoice-aging")
  getInvoiceAging(...GetInvoiceAgingRequest): GetInvoiceAgingResponse;
}
