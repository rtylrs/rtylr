import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;
using common.pagination.v1;

namespace insights.customer.v1;

model CustomerAnalytics {
  /** total is the total number of matching records. */
  total: int32;

  /** new_this_period is the new this period value for this contract. */
  new_this_period: int32;

  /** retention_rate is the retention rate value for this contract. */
  retention_rate: float64;

  /** avg_lifetime_value is the avg lifetime value value for this contract. */
  avg_lifetime_value: float64;

  /** churn_rate is the churn rate value for this contract. */
  churn_rate: float64;

  /** growth is the growth value for this contract. */
  growth: CustomerGrowthPoint[];
}

model CustomerGrowthPoint {
  /** date is the date value for this contract. */
  date: string;

  /** new_customers is the new customers value for this contract. */
  new_customers: int32;

  /** total is the total number of matching records. */
  total: int32;
}

model TopCustomerEntry {
  /** id is the unique identifier for the resource. */
  @encode(string)
  id: int64;

  /** name is the name value for this contract. */
  name: string;

  /** total_spent is the total spent value for this contract. */
  total_spent: float64;

  /** order_count is the order count value for this contract. */
  order_count: int32;

  /** last_order_at is the last order at value for this contract. */
  last_order_at: string;
}

model ChurnAnalysis {
  /** churn_rate is the churn rate value for this contract. */
  churn_rate: float64;

  /** churned_count is the churned count value for this contract. */
  churned_count: int32;

  /** reasons is the reasons value for this contract. */
  reasons: ChurnReason[];
}

model ChurnReason {
  /** reason is the reason value for this contract. */
  reason: string;

  /** count is the count value for this contract. */
  count: int32;

  /** percentage is the percentage value for this contract. */
  percentage: float64;
}

model LTVAnalysis {
  /** average is the average value for this contract. */
  average: float64;

  /** median is the median value for this contract. */
  median: float64;

  /** segments is the segments value for this contract. */
  segments: LTVSegment[];
}

model LTVSegment {
  /** segment is the segment value for this contract. */
  segment: string;

  /** avg_ltv is the avg ltv value for this contract. */
  avg_ltv: float64;

  /** customer_count is the customer count value for this contract. */
  customer_count: int32;
}

model GetCustomerAnalyticsRequest {
  /** period is the period value for this contract. */
  @query("period")
  period: string;
}

model GetCustomerAnalyticsResponse {
  /** analytics is the analytics value for this contract. */
  analytics: CustomerAnalytics;
}

model GetChurnAnalysisRequest {
  /** period is the period value for this contract. */
  @query("period")
  period: string;
}

model GetChurnAnalysisResponse {
  /** analysis is the analysis value for this contract. */
  analysis: ChurnAnalysis;
}

model GetLTVAnalysisRequest {}

model GetLTVAnalysisResponse {
  /** analysis is the analysis value for this contract. */
  analysis: LTVAnalysis;
}

model GetTopCustomersRequest {
  /** page is the current result page. */
  @query("page")
  page: int32;

  /** limit is the maximum number of records to return. */
  @query("limit")
  limit: int32;

  /** sort_by is the sort field (total_spent, order_count). */
  @query("sort_by")
  sort_by: string;

  /** sort_dir is the sort direction ("asc" or "desc"). */
  @query("sort_dir")
  sort_dir: string;

  /** period is the period value for this contract. */
  @query("period")
  period: string;
}

model GetTopCustomersResponse {
  /** customers is the customers value for this contract. */
  customers: TopCustomerEntry[];

  /** pagination is the pagination value for this contract. */
  pagination: PaginationResponse;
}
