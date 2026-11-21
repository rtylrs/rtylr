import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace insights.customer.v1;

/** CustomerService defines insights customer analytics endpoints. */
@route("/insights/customer/v1")
interface CustomerService {
  /** Returns customer analytics for a period. */
  @get
  @route("/analytics")
  getCustomerAnalytics(
    ...GetCustomerAnalyticsRequest,
  ): GetCustomerAnalyticsResponse;

  /** Returns churn analysis for a period. */
  @get
  @route("/churn")
  getChurnAnalysis(...GetChurnAnalysisRequest): GetChurnAnalysisResponse;

  /** Returns lifetime value analysis. */
  @get
  @route("/ltv")
  getLTVAnalysis(...GetLTVAnalysisRequest): GetLTVAnalysisResponse;

  /** Returns top customers by spend or order count. */
  @get
  @route("/top")
  getTopCustomers(...GetTopCustomersRequest): GetTopCustomersResponse;
}
