import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;
using common.pagination.v1;

namespace flow.view.v2;

/**
 * View is a saved filter/sort preset. Business-shared in v2: every member of
 * the caller's business can list and use saved views, matching v1's
 * workspace-sharing semantics, with a per-view visibility toggle: see
 * `visibility` below.
 */
model View {
  @encode(string)
  id: int64;

  name: string;
  icon: string;

  /** JSON-encoded filter criteria for this saved view. */
  filters: string;

  group_by: string;
  sort_by: string;
  sort_dir: string;

  /**
   * visibility is "shared" (every colleague in the business can use this view)
   * or "private" (only its creator). Empty means "shared", which is how every
   * view behaved before this field existed, so an older client keeps its
   * current reach rather than silently making views private.
   */
  visibility: string;
}

model ListViewsRequest {
  /** The current result page. */
  @query("page")
  page: int32;

  /** The maximum number of records to return. */
  @query("limit")
  limit: int32;
}

model ListViewsResponse {
  views: View[];
  pagination: PaginationResponse;
}

model CreateViewRequest {
  @Onkt.required
  name: string;

  icon: string;

  /** JSON-encoded filter criteria for this saved view. */
  filters: string;

  group_by: string;
  sort_by: string;
  sort_dir: string;

  /**
   * visibility is "shared" (every colleague in the business can use this view)
   * or "private" (only its creator). Empty means "shared", which is how every
   * view behaved before this field existed, so an older client keeps its
   * current reach rather than silently making views private.
   */
  visibility: string;
}

model CreateViewResponse {
  view: View;
}

model UpdateViewRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  name: string;
  icon: string;

  /** JSON-encoded filter criteria for this saved view. */
  filters: string;

  group_by: string;
  sort_by: string;
  sort_dir: string;

  /**
   * visibility is "shared" (every colleague in the business can use this view)
   * or "private" (only its creator). Empty means "shared", which is how every
   * view behaved before this field existed, so an older client keeps its
   * current reach rather than silently making views private.
   */
  visibility: string;
}

model UpdateViewResponse {}

model DeleteViewRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model DeleteViewResponse {}
