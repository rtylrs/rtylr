import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace common.pagination.v1;

/** Standard pagination request params. */
model PaginationRequest {
  /** page is the current result page. */
  @query("page")
  page: int32;

  /** limit is the maximum number of records to return. */
  @query("limit")
  limit: int32;

  /** sort_by is the sort by value for this contract. */
  @query("sort_by")
  sort_by: string;

  /**
   * sort_dir is the sort dir value for this contract.
   * "asc" or "desc"
   */
  @query("sort_dir")
  sort_dir: string;
}

/** Standard pagination response metadata. */
model PaginationResponse {
  /** page is the current result page. */
  page: int32;

  /** limit is the maximum number of records to return. */
  limit: int32;

  /** total is the total number of matching records. */
  total: int32;

  /** total_pages is the total pages value for this contract. */
  total_pages: int32;

  /** has_next is the has next value for this contract. */
  has_next: boolean;

  /** has_prev is the has prev value for this contract. */
  has_prev: boolean;
}

/** Standard timestamp wrapper for explicit RFC 3339 JSON encoding. */
model Timestamp {
  value: string;
}
