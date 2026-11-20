import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;
using common.pagination.v1;

namespace flow.milestone.v2;

model MilestoneProject {
  name: string;
  color: string;
}

model Milestone {
  @encode(string)
  id: int64;

  @encode(string)
  project_id: int64;

  name: string;
  target_date: string;
  progress: int32;
  computed_progress: int32;
  done_count: int32;
  task_count: int32;
  status: string;
  description: string;
  project: MilestoneProject;
}

model ListMilestonesRequest {
  /**
   * project_id narrows milestones to one project. Omitted, all milestones of
   * projects the caller can access are returned.
   */
  @query("project_id")
  @encode(string)
  project_id?: int64;

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

model ListMilestonesResponse {
  milestones: Milestone[];
  pagination: PaginationResponse;
}

model CreateMilestoneRequest {
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  project_id?: int64;

  @Onkt.required
  name: string;

  description: string;
  target_date: string;
}

model CreateMilestoneResponse {
  milestone: Milestone;
}

model UpdateMilestoneRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  name: string;
  description: string;
  status: string;
  target_date: string;
}

model UpdateMilestoneResponse {}
