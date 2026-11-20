import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;
using flow.task.v1;

namespace flow.dashboard.v1;

model GetFlowDashboardRequest {
  /** workspace_id is the workspace id value for this contract. */
  @Onkt.required
  @minValueExclusive(0)
  @query("workspace_id")
  @encode(string)
  workspace_id?: int64;
}

model GetFlowDashboardResponse {
  /** open_tasks is the open tasks value for this contract. */
  @encode(string)
  open_tasks: int64;

  /** in_progress_tasks is the in progress tasks value for this contract. */
  @encode(string)
  in_progress_tasks: int64;

  /** completed_this_week is the completed this week value for this contract. */
  @encode(string)
  completed_this_week: int64;

  /** overdue_tasks is the overdue tasks value for this contract. */
  @encode(string)
  overdue_tasks: int64;

  /** my_tasks is the my tasks value for this contract. */
  my_tasks: Task[];
}
