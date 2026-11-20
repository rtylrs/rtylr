import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace flow.status.v2;

/**
 * Status is one entry in a project's own status set.
 *
 * Versioned v2 rather than v1 despite being a new capability. Every live flow
 * capability is on v2 after the workspace removal, and v1 contracts are still
 * being deleted, so a lone `flow/status/v1` would read as the older of two
 * generations when it is the only one that exists. Nothing here has a v1
 * history to be compatible with.
 *
 * A project's statuses are its board columns. Every project has a set: it is
 * copied from the business's status template when the project is created.
 */
model Status {
  /** id is the unique identifier for the resource. */
  @encode(string)
  id: int64;

  /**
   * project_id is the project this status belongs to. Status sets are per
   * project by design: two projects in one business may run entirely different
   * pipelines.
   */
  @encode(string)
  project_id: int64;

  /** name is what the user sees on the board. Unique within a project. */
  name: string;

  /**
   * bucket is the semantic meaning, one of "backlog", "active", "done" or
   * "cancelled".
   *
   * It exists so a rollup can count across projects whose pipelines have
   * nothing in common: a project calling its in-flight state "Building" and
   * one calling it "Legal review" both report `active`.
   *
   * The contract cannot express a closed set, so the server validates it and
   * rejects anything else with 400. Do not treat an unrecognised value as a
   * default; that is how a cancelled task would get counted as open.
   */
  bucket: string;

  /**
   * legacy_status is which of the six built-in values `Task.status` carries
   * while this status is applied: "backlog", "todo", "in_progress",
   * "in_review", "done" or "cancelled".
   *
   * The string status column is not going away, so every existing query,
   * filter and index keeps working and no consumer has to learn about this
   * capability to stay correct. A custom status reports its nearest built-in.
   */
  legacy_status: string;

  /** color is the swatch shown on the board, as a hex string. */
  color: string;

  /** sort_order is the position within the project's set, ascending. */
  sort_order: int32;

  /**
   * is_default marks the status a newly created task receives. At most one per
   * project, enforced by the server rather than by the schema, because MySQL
   * cannot express a partial unique index on it.
   */
  is_default: boolean;

  /** created_at is when the record was created. */
  created_at: string;
}

model ListStatusesRequest {
  /**
   * project_id scopes the listing to one project's set. Required: there is no
   * business-wide set to fall back on, and a listing across projects would
   * mix pipelines that have nothing to do with each other.
   */
  @query("project_id")
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  project_id?: int64;
}

model ListStatusesResponse {
  /**
   * statuses is the project's set in `sort_order`, which is also its board
   * column order. A project without statuses is seeded on first read.
   */
  statuses: Status[];
}

model CreateStatusRequest {
  /** project_id is the project to add the status to. */
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  project_id?: int64;

  /** name must be unique within the project. A duplicate is rejected with 409. */
  @Onkt.required
  name: string;

  /** bucket is one of "backlog", "active", "done", "cancelled". */
  @Onkt.required
  bucket: string;

  /**
   * legacy_status is one of the six built-in values. Optional: when absent the
   * server derives it from the bucket, so a caller that does not care about
   * the compatibility column can leave it alone.
   */
  legacy_status: string;

  /** color is a hex string. Optional, server default applies. */
  color: string;

  /** sort_order places the status in the set. Optional, appended when absent. */
  sort_order: int32;

  /**
   * is_default makes this the status new tasks receive, clearing the flag on
   * whichever status held it.
   */
  is_default: boolean;
}

model CreateStatusResponse {
  /** status is the created status, including the derived legacy_status. */
  status: Status;
}

model UpdateStatusRequest {
  /** id is the status identifier. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  @Onkt.required
  @body
  body: UpdateStatusBody;
}

/**
 * UpdateStatusBody is a replacement, so send the whole status.
 *
 * Only `name`, `bucket` and `color` are required. `legacy_status` stays
 * optional for the same reason as on create, and nothing here is required
 * merely because the row has a column for it: end_date on the HR staff body
 * was \@required for exactly that reason and it made every update to an active
 * employee silently impossible.
 */
model UpdateStatusBody {
  /** name must remain unique within the project. */
  @Onkt.required
  name: string;

  /** bucket is one of the four semantic buckets. */
  @Onkt.required
  bucket: string;

  /** color is a hex string. */
  @Onkt.required
  color: string;

  /**
   * legacy_status is one of the six built-in values, derived from the bucket
   * when absent.
   */
  legacy_status: string;

  /** sort_order is the position within the set. */
  sort_order: int32;

  /** is_default makes this the status new tasks receive. */
  is_default: boolean;
}

model UpdateStatusResponse {
  /** status is the updated status. */
  status: Status;
}

model ReorderStatusesRequest {
  /** project_id is the project whose set is reordered. */
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  project_id?: int64;

  /**
   * status_ids is every status in the project, in the new order. It must be
   * exactly the project's current set: a missing, extra, repeated or foreign
   * id is rejected with 400 rather than guessed at, because a partial list
   * has no single correct place for the statuses it leaves out.
   */
  @Onkt.required
  @Onkt.encode("string")
  status_ids: int64[];
}

model ReorderStatusesResponse {
  /**
   * statuses is the project's set in its new order, with the rewritten
   * sort_order values.
   */
  statuses: Status[];
}

model DeleteStatusRequest {
  /** id is the status identifier. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  /**
   * move_to_status_id is the status in the same project that receives the
   * deleted status's tasks. Required when the status still has tasks.
   */
  @query("move_to_status_id")
  @encode(string)
  move_to_status_id: int64;
}

model DeleteStatusResponse {
  /** message is a human-readable result message. */
  message: string;

  /**
   * reassigned_task_count is how many tasks moved to move_to_status_id because
   * they were on the deleted status. Tasks are never deleted with a status.
   */
  reassigned_task_count: int32;
}

model SetTaskStatusRequest {
  /** task_id is the task to move. */
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  task_id?: int64;

  /**
   * status_id is the custom status to apply. Zero clears it, putting the task
   * back on its built-in string status.
   *
   * Not \@required \@gt(0), because zero is the meaningful "no custom status"
   * value rather than a missing argument.
   */
  @encode(string)
  status_id: int64;
}

model SetTaskStatusResponse {
  /** message is a human-readable result message. */
  message: string;

  /**
   * status is the resulting built-in string status the task now carries, so a
   * caller does not need a second request to keep its cached task correct.
   */
  status: string;
}

model StatusTemplateItem {
  /** name is the status name. */
  @Onkt.required
  name: string;

  /** bucket is one of backlog, active, done, cancelled. */
  @Onkt.required
  bucket: string;

  /** legacy_status is the built-in status the entry maps to; derived from the bucket when empty. */
  legacy_status: string;

  /** color is a hex colour. */
  color: string;

  /** sort_order is the position in the template. */
  sort_order: int32;
}

model GetStatusTemplateRequest {}

model GetStatusTemplateResponse {
  /** statuses is the template in sort order; the built-in default set when never saved. */
  statuses: StatusTemplateItem[];
}

model SaveStatusTemplateRequest {
  /** statuses is the full replacement template. */
  @Onkt.required
  statuses: StatusTemplateItem[];
}

model SaveStatusTemplateResponse {
  /** statuses is the saved template in sort order. */
  statuses: StatusTemplateItem[];
}
