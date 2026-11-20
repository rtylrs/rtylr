import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hr.dashboard.v1;

/** DashboardService defines HR dashboard endpoints. */
@route("/hr/dashboard/v1")
interface DashboardService {
  /** GetHRDashboard returns the HR dashboard summary. */
  @get
  @route("/dashboard")
  getHRDashboard(...GetHRDashboardRequest): GetHRDashboardResponse;

  /** GetDashboardWidgets returns widget layout and preloaded data. */
  @get
  @route("/dashboard/widgets")
  getDashboardWidgets(
    ...GetDashboardWidgetsRequest,
  ): GetDashboardWidgetsResponse;

  /** UpdateDashboardWidgets saves widget layout for the current business. */
  @put
  @route("/dashboard/widgets")
  updateDashboardWidgets(
    ...UpdateDashboardWidgetsRequest,
  ): UpdateDashboardWidgetsResponse;
}
