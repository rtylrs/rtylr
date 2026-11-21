import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace insights.dashboard.v1;

/** DashboardService defines insights dashboard endpoints. */
@route("/insights/dashboard/v1")
interface DashboardService {
  /** Returns high-level insights summary. */
  @get
  @route("/insights")
  getInsights(...GetInsightsRequest): GetInsightsResponse;

  /** Returns the product roadmap view. */
  @get
  @route("/roadmap")
  getRoadmap(...GetRoadmapRequest): GetRoadmapResponse;

  /** Returns team workload metrics. */
  @get
  @route("/workload")
  getWorkload(...GetWorkloadRequest): GetWorkloadResponse;

  /** Returns MTD headline KPI cards for the insights home dashboard. */
  @get
  @route("/summary")
  getDashboardSummary(
    ...GetDashboardSummaryRequest,
  ): GetDashboardSummaryResponse;

  /** Returns CRM open pipeline, month-to-date won deal value and recurring revenue for the insights home dashboard. */
  @get
  @route("/sales")
  getSalesSummary(...GetSalesSummaryRequest): GetSalesSummaryResponse;

  /** Returns the hours logged in Work this month to date, in total and per public project, for the insights home dashboard. */
  @get
  @route("/hours")
  getHoursSummary(...GetHoursSummaryRequest): GetHoursSummaryResponse;

  /** Lists the Work projects the caller may pick a project dashboard for: every public project, the private projects they own, created or are a member of, and the projects Work's sidebar lists for them, such as their reports' projects. Curated for everyone: a business admin's list does not include every private project. */
  @get
  @route("/projects")
  listProjects(...ListProjectsRequest): ListProjectsResponse;

  /** Returns one Work project's dashboard: task progress, overdue tasks, hours logged and, for callers who may view finance invoices, the project's invoice totals. A private project is open to whoever Work lets open it (owner, creator, member, or the manager of someone involved) and to callers with Work's business-wide reporting access (business owner, admin, or work team manager); anyone else gets not found. */
  @get
  @route("/projects/{project_id}")
  getProjectDashboard(
    ...GetProjectDashboardRequest,
  ): GetProjectDashboardResponse;
}
