import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace flow.status.v2;

/**
 * StatusService manages per-project statuses.
 *
 * A project's statuses are its board columns: the project columns API reads
 * and writes these same rows, and moving a card sets the task's status.
 */
@route("/flow/status/v2")
interface StatusService {
  /** Returns one project's status set in sort order, empty when unconfigured. */
  @get
  @route("/statuses")
  listStatuses(...ListStatusesRequest): ListStatusesResponse;

  /** Creates a status in a project's set. */
  @post
  @route("/statuses")
  createStatus(...CreateStatusRequest): CreateStatusResponse;

  /** Replaces a status. */
  @put
  @route("/statuses/{id}")
  updateStatus(...UpdateStatusRequest): UpdateStatusResponse;

  /**
   * Sets the order of a project's whole status set in one write.
   *
   * Reordering through updateStatus took one request per status, so a failure
   * part way through left the board half reordered. This either applies the
   * full order or nothing.
   */
  @put
  @route("/statuses/order")
  reorderStatuses(...ReorderStatusesRequest): ReorderStatusesResponse;

  /**
   * Deletes a status. A status that still has tasks needs move_to_status_id,
   * and a project's last done status cannot be deleted.
   */
  @delete
  @route("/statuses/{id}")
  deleteStatus(...DeleteStatusRequest): DeleteStatusResponse;

  /** Returns the business's status template, the set new projects copy. */
  @get
  @route("/status-template")
  getStatusTemplate(...GetStatusTemplateRequest): GetStatusTemplateResponse;

  /** Replaces the business's status template. Existing projects keep their own set. */
  @put
  @route("/status-template")
  saveStatusTemplate(...SaveStatusTemplateRequest): SaveStatusTemplateResponse;

  /**
   * Applies a custom status to a task, or clears it with status_id zero.
   *
   * Lives here rather than on the task service, mirroring setTaskLabels in
   * flow.label.v2: the capability that owns the concept owns the setter.
   */
  @put
  @route("/tasks/status")
  setTaskStatus(...SetTaskStatusRequest): SetTaskStatusResponse;
}
