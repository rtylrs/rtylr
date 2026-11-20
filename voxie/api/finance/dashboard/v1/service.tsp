import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace finance.dashboard.v1;

/** DashboardService defines the finance dashboard endpoint. */
@route("/finance/dashboard/v1")
interface DashboardService {
  @get
  @route("/dashboard")
  getFinanceDashboard(
    ...GetFinanceDashboardRequest,
  ): GetFinanceDashboardResponse;
}
