import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace insights.revenue.v1;

model RevenueOverview {
  /** total is the total number of matching records. */
  total: float64;

  /** change is the change value for this contract. */
  change: float64;

  /** trend is the trend value for this contract. */
  trend: RevenuePoint[];
}

model RevenuePoint {
  /** date is the date value for this contract. */
  date: string;

  /** amount is the amount value for this contract. */
  amount: float64;
}

model RevenueByApp {
  /** app is the app value for this contract. */
  app: string;

  /** amount is the amount value for this contract. */
  amount: float64;

  /** percentage is the percentage value for this contract. */
  percentage: float64;
}

model RevenueComparison {
  /** period is the period value for this contract. */
  period: string;

  /** current is the current value for this contract. */
  current: float64;

  /** previous is the previous value for this contract. */
  previous: float64;

  /** change is the change value for this contract. */
  change: float64;
}

model GetRevenueOverviewRequest {
  /** period is the period value for this contract. */
  @query("period")
  period: string;
}

model GetRevenueOverviewResponse {
  /** overview is the overview value for this contract. */
  overview: RevenueOverview;
}

model GetRevenueByAppRequest {
  /** period is the period value for this contract. */
  @query("period")
  period: string;
}

model GetRevenueByAppResponse {
  /** apps is the apps value for this contract. */
  apps: RevenueByApp[];
}

model GetRevenueComparisonRequest {
  /** period is the period value for this contract. */
  @query("period")
  period: string;

  /**
   * compare_to is the compare to value for this contract.
   * previous_period, same_period_last_year
   */
  @Onkt.required
  @query("compare_to")
  compare_to: string;
}

model GetRevenueComparisonResponse {
  /** comparisons is the comparisons value for this contract. */
  comparisons: RevenueComparison[];
}
