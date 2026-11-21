import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace pos.dashboard.v1;

/** DashboardService defines POS dashboard summary endpoints. */
@route("/pos/dashboard/v1")
interface DashboardService {
  /** Returns aggregate sales metrics for the POS dashboard. */
  @get
  @route("/dashboard")
  dashboard(...DashboardRequest): DashboardResponse;
}
