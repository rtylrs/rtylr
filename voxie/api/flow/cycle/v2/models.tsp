import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;
using common.pagination.v1;

namespace flow.cycle.v2;

/**
 * Cycle is a sprint-style time box scoped to ONE project in v2. v1 attached
 * cycles to workspaces with an optional project filter; the removal of
 * workspaces makes the project the only sensible owner.
 */
model Cycle {
  /** id is the unique identifier for the resource. */
  @encode(string)
  id: int64;

  /**
   * project_id is the project this cycle belongs to. Always set in v2;
   * workspace-era cycles were migrated onto their filtered project or archived.
   */
  @encode(string)
  project_id: int64;

  /** name is the name value for this contract. */
  name: string;

  /** start_date is the start date value for this contract. */
  start_date: string;

  /** end_date is the end date value for this contract. */
  end_date: string;

  /** status is the current lifecycle state. */
  status: string;

  /** scope is the total task count in the cycle. Display projection. */
  scope: int32;

  /** completed is how many of those tasks are done. Display projection. */
  completed: int32;
}

model ListCyclesRequest {
  /** project_id lists the cycles of one project. */
  @Onkt.required
  @minValueExclusive(0)
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

model ListCyclesResponse {
  cycles: Cycle[];
  pagination: PaginationResponse;
}

model CreateCycleRequest {
  @Onkt.required
  name: string;

  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  project_id?: int64;

  start_date: string;
  end_date: string;
}

model CreateCycleResponse {
  cycle: Cycle;
}

model UpdateCycleRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  name: string;
  status: string;
  end_date: string;
}

model UpdateCycleResponse {}
