import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace booking.dashboard.v1;

@route("/booking/dashboard/v1")
interface DashboardService {
  @get
  @route("/dashboard")
  getDashboard(...GetDashboardRequest): GetDashboardResponse;
}
