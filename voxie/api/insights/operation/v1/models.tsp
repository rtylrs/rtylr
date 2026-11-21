import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace insights.operation.v1;

model UtilizationMetrics {
  /** avg_utilization is the avg utilization value for this contract. */
  avg_utilization: float64;

  /** total_staff is the total staff value for this contract. */
  total_staff: int32;

  /** active_today is the active today value for this contract. */
  active_today: int32;

  /** total_hours is the total hours value for this contract. */
  total_hours: float64;

  /** available_hours is the available hours capacity for the period. */
  available_hours: float64;

  /** overtime_hours is the overtime hours value for this contract. */
  overtime_hours: float64;

  /** by_staff is the by staff value for this contract. */
  by_staff: StaffUtilization[];
}

model StaffUtilization {
  /** staff_id is the staff id value for this contract. */
  @encode(string)
  staff_id: int64;

  /** name is the name value for this contract. */
  name: string;

  /** hours_worked is the hours worked value for this contract. */
  hours_worked: float64;

  /** hours_scheduled is the hours scheduled value for this contract. */
  hours_scheduled: float64;

  /** utilization is the utilization value for this contract. */
  utilization: float64;
}

model ShiftMetrics {
  /** total_shifts is the total shifts value for this contract. */
  total_shifts: int32;

  /** covered is the covered value for this contract. */
  covered: int32;

  /** missed is the missed value for this contract. */
  missed: int32;

  /** coverage_rate is the coverage rate value for this contract. */
  coverage_rate: float64;

  /** by_day is the by day value for this contract. */
  by_day: ShiftsByDay[];
}

model ShiftsByDay {
  /** day is the day value for this contract. */
  day: string;

  /** scheduled is the scheduled value for this contract. */
  scheduled: int32;

  /** completed is the completed value for this contract. */
  completed: int32;
}

model TimeOffMetrics {
  /** total_requests is the total requests value for this contract. */
  total_requests: int32;

  /** approved is the approved value for this contract. */
  approved: int32;

  /** pending is the pending value for this contract. */
  pending: int32;

  /** rejected is the rejected value for this contract. */
  rejected: int32;

  /** by_type is the by type value for this contract. */
  by_type: TimeOffByType[];
}

model TimeOffByType {
  /** type is the type value for this contract. */
  type: string;

  /** days is the days value for this contract. */
  days: int32;

  /** requests is the requests value for this contract. */
  requests: int32;
}

model HeadcountMetrics {
  /** current is the current value for this contract. */
  current: int32;

  /** hired_this_period is the hired this period value for this contract. */
  hired_this_period: int32;

  /** departed_this_period is the departed this period value for this contract. */
  departed_this_period: int32;

  /** turnover_rate is the turnover rate value for this contract. */
  turnover_rate: float64;

  /** trend is the trend value for this contract. */
  trend: HeadcountPoint[];
}

model HeadcountPoint {
  /** date is the date value for this contract. */
  date: string;

  /** count is the count value for this contract. */
  count: int32;
}

model GetHeadcountMetricsRequest {
  /** period is the period value for this contract. */
  @query("period")
  period: string;
}

model GetHeadcountMetricsResponse {
  /** metrics is the metrics value for this contract. */
  metrics: HeadcountMetrics;
}

model GetOperationsSummaryRequest {}

model GetOperationsSummaryResponse {
  /** active_staff is the active staff count. */
  @encode(string)
  active_staff: int64;

  /** shifts_this_week is the scheduled shift count this week. */
  @encode(string)
  shifts_this_week: int64;

  /** pending_time_off is the pending time-off request count. */
  @encode(string)
  pending_time_off: int64;

  /** open_leave_requests is the open leave request count. */
  @encode(string)
  open_leave_requests: int64;
}

model GetShiftMetricsRequest {
  /** period is the period value for this contract. */
  @query("period")
  period: string;
}

model GetShiftMetricsResponse {
  /** metrics is the metrics value for this contract. */
  metrics: ShiftMetrics;
}

model GetTimeOffMetricsRequest {
  /** period is the period value for this contract. */
  @query("period")
  period: string;
}

model GetTimeOffMetricsResponse {
  /** metrics is the metrics value for this contract. */
  metrics: TimeOffMetrics;
}

model GetUtilizationRequest {
  /** period is the period value for this contract. */
  @query("period")
  period: string;
}

model GetUtilizationResponse {
  /** metrics is the metrics value for this contract. */
  metrics: UtilizationMetrics;
}
