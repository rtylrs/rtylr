import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;
using common.pagination.v1;

namespace flow.insights.v2;

model VelocityPoint {
  cycle: string;
  completed: int32;
  total: int32;
  completion_rate: float64;
}

model BurndownPoint {
  day: int32;
  ideal: float64;
  actual: float64;
}

model MemberTime {
  id: string;
  name: string;
  hours: float64;
}

model ThroughputPoint {
  week: string;
  completed: int32;
}

model TeamTaskSummary {
  id: string;
  title: string;
  status: string;
  priority: string;
  due_date: string;
}

model TeamMemberWorkload {
  id: string;
  name: string;
  initials: string;
  capacity: int32;
  assigned_hours: float64;
  load: float64;
  tasks: TeamTaskSummary[];
}

/**
 * RoadmapProject is one project row on the timeline view. v2 drops the
 * workspace field; visibility filtering (which projects a caller sees named
 * here) is applied server-side per the business's policy.
 */
model RoadmapProject {
  id: string;
  name: string;
  color: string;
  start_date: string;
  target_date: string;
  progress: int32;
  done_count: int32;
  task_count: int32;
  description: string;
}

model GetInsightsRequest {}

model GetInsightsResponse {
  velocity: VelocityPoint[];
  burndown: BurndownPoint[];
  cycle_time: float64;
  time_by_member: MemberTime[];
  throughput: ThroughputPoint[];
  total_tasks: int32;
  done_tasks: int32;
}

model GetTeamWorkloadRequest {
  /** The current result page. */
  @query("page")
  page: int32;

  /** The maximum number of records to return. */
  @query("limit")
  limit: int32;

  /** The field to sort results by. */
  @query("sort_by")
  sort_by: string;

  /** The sort direction ("asc" or "desc"). */
  @query("sort_dir")
  sort_dir: string;
}

model GetTeamWorkloadResponse {
  members: TeamMemberWorkload[];
  pagination: PaginationResponse;
}

model GetRoadmapRequest {
  /** The current result page. */
  @query("page")
  page: int32;

  /** The maximum number of records to return. */
  @query("limit")
  limit: int32;

  /** The field to sort results by. */
  @query("sort_by")
  sort_by: string;

  /** The sort direction ("asc" or "desc"). */
  @query("sort_dir")
  sort_dir: string;
}

model GetRoadmapResponse {
  projects: RoadmapProject[];
  pagination: PaginationResponse;
}
