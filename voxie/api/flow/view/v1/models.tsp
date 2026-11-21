import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;
using common.pagination.v1;

namespace flow.view.v1;

model View {
  @encode(string)
  id: int64;

  name: string;
  icon: string;

  /**
   * visibility is "shared" (every colleague in the business can use this view)
   * or "private" (only its creator). Empty on create means "shared", which is
   * how every view behaved before this field existed, so an older client keeps
   * its current behaviour rather than silently making views private.
   */
  visibility: string;

  /** JSON-encoded filter criteria for this saved view. */
  filters: string;

  group_by: string;
  sort_by: string;
  sort_dir: string;
}

model ListViewsRequest {
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

model ListViewsResponse {
  views: View[];
  pagination: PaginationResponse;
}

model CreateViewRequest {
  workspace_id: string;

  @Onkt.required
  name: string;

  icon: string;

  /**
   * visibility is "shared" (every colleague in the business can use this view)
   * or "private" (only its creator). Empty on create means "shared", which is
   * how every view behaved before this field existed, so an older client keeps
   * its current behaviour rather than silently making views private.
   */
  visibility: string;

  /** JSON-encoded filter criteria for this saved view. */
  filters: string;

  group_by: string;
  sort_by: string;
  sort_dir: string;
}

model CreateViewResponse {
  view: View;
}

model UpdateViewRequest {
  @encode(string)
  id: int64;

  name: string;
  icon: string;

  /**
   * visibility is "shared" (every colleague in the business can use this view)
   * or "private" (only its creator). Empty on create means "shared", which is
   * how every view behaved before this field existed, so an older client keeps
   * its current behaviour rather than silently making views private.
   */
  visibility: string;

  /** JSON-encoded filter criteria for this saved view. */
  filters: string;

  group_by: string;
  sort_by: string;
  sort_dir: string;
}

model UpdateViewResponse {}

model DeleteViewRequest {
  @encode(string)
  id: int64;
}

model DeleteViewResponse {}
