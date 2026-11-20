import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hub.dashboard.v1;

/** DashboardService defines hub dashboard endpoints. */
@route("/hub/dashboard/v1")
interface DashboardService {
  /** Returns key performance indicators for the active business. */
  @get
  @route("/kpis")
  getKPIs(...GetKPIsRequest): GetKPIsResponse;
}
