import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;
using common.pagination.v1;

namespace flow.label.v2;

/**
 * Label is business-scoped in v2: one pool shared by every project in the
 * caller's business. Names are not unique by contract; duplicate names from
 * the migration are tolerated and may be deduped later without a breaking
 * change.
 */
model Label {
  /** id is the unique identifier for the resource. */
  @encode(string)
  id: int64;

  /** name is the name value for this contract. */
  name: string;

  /** color is the color value for this contract. */
  color: string;

  /** created_at is when the record was created. */
  created_at: string;
}

model ListLabelsRequest {
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

model ListLabelsResponse {
  /** labels is the labels value for this contract. */
  labels: Label[];

  pagination: PaginationResponse;
}

model CreateLabelRequest {
  /** name is the name value for this contract. */
  @Onkt.required
  name: string;

  /** color is the color value for this contract. */
  @Onkt.required
  color: string;
}

model CreateLabelResponse {
  /** label is the label value for this contract. */
  label: Label;
}

model UpdateLabelRequest {
  /** id is the label identifier. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  /** name is the label name. */
  @Onkt.required
  name: string;

  /** color is the label color. */
  @Onkt.required
  color: string;
}

model UpdateLabelResponse {
  /** message is a human-readable result message. */
  message: string;
}

model DeleteLabelRequest {
  /** id is the label identifier. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model DeleteLabelResponse {
  /** message is a human-readable result message. */
  message: string;
}

model SetTaskLabelsRequest {
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  task_id?: int64;

  @Onkt.encode("string")
  label_ids: int64[];
}

model SetTaskLabelsResponse {
  message: string;
}
