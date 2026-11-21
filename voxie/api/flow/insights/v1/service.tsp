import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace flow.insights.v1;

@route("/flow/insights/v1")
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
