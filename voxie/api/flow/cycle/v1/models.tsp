import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;
using common.pagination.v1;

namespace flow.cycle.v1;

model Cycle {
  @encode(string)
  id: int64;

  @encode(string)
  workspace_id: int64;

  name: string;
  start_date: string;
  end_date: string;
  status: string;
  scope: int32;
  completed: int32;
}

model ListCyclesRequest {
  @query("workspace_id")
  workspace_id: string;

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

  workspace_id: string;
  start_date: string;
  end_date: string;
}

model CreateCycleResponse {
  cycle: Cycle;
}

model UpdateCycleRequest {
  @encode(string)
  id: int64;

  name: string;
  status: string;
  end_date: string;
}

model UpdateCycleResponse {}
