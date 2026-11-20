import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace flow.insights.v2;

/**
 * InsightsService aggregates business-wide delivery metrics. Every request
 * implicitly scopes to the caller's business; per-user visibility of named
 * projects (roadmap, workload drill-downs) is applied server-side.
 */
@route("/flow/insights/v2")
interface InsightsService {
  @get
  @route("/overview")
  getInsights(...GetInsightsRequest): GetInsightsResponse;

  @get
  @route("/team/workload")
  getTeamWorkload(...GetTeamWorkloadRequest): GetTeamWorkloadResponse;

  @get
  @route("/roadmap")
  getRoadmap(...GetRoadmapRequest): GetRoadmapResponse;
}
