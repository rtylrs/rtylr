import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace insights.revenue.v1;

/** RevenueService defines insights revenue analytics endpoints. */
@route("/insights/revenue/v1")
interface RevenueService {
  /** Returns revenue overview for a period. */
  @get
  @route("/overview")
  getRevenueOverview(...GetRevenueOverviewRequest): GetRevenueOverviewResponse;

  /** Returns revenue broken down by app. */
  @get
  @route("/by-app")
  getRevenueByApp(...GetRevenueByAppRequest): GetRevenueByAppResponse;

  /** Compares revenue across periods. */
  @get
  @route("/comparison")
  getRevenueComparison(
    ...GetRevenueComparisonRequest,
  ): GetRevenueComparisonResponse;
}
