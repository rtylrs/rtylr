import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace insights.operation.v1;

/** OperationService defines insights operations analytics endpoints. */
@route("/insights/operation/v1")
interface OperationService {
  /** Returns headcount metrics for a period. */
  @get
  @route("/headcount")
  getHeadcountMetrics(
    ...GetHeadcountMetricsRequest,
  ): GetHeadcountMetricsResponse;

  /** Returns shift metrics for a period. */
  @get
  @route("/shifts")
  getShiftMetrics(...GetShiftMetricsRequest): GetShiftMetricsResponse;

  /** Returns time-off metrics for a period. */
  @get
  @route("/time-off")
  getTimeOffMetrics(...GetTimeOffMetricsRequest): GetTimeOffMetricsResponse;

  /** Returns utilization metrics for a period. */
  @get
  @route("/utilization")
  getUtilization(...GetUtilizationRequest): GetUtilizationResponse;

  /** Returns headline HR operations KPI counters. */
  @get
  @route("/summary")
  getOperationsSummary(
    ...GetOperationsSummaryRequest,
  ): GetOperationsSummaryResponse;
}
