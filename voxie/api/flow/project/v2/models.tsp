import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;
using common.pagination.v1;

namespace flow.project.v2;

/**
 * Project is a business-scoped container of tasks. The workspace layer is
 * gone in v2: every project belongs directly to the caller's business,
 * resolved server-side from auth, and no request carries a tenant or
 * workspace identifier.
 */
model Project {
  /** id is the unique identifier for the resource. */
  @encode(string)
  id: int64;

  /**
   * owner_id is the user who created the project. Not the tenant: the
   * business is implicit from the caller's auth context.
   */
  @encode(string)
  owner_id: int64;

  /** name is the name value for this contract. */
  name: string;

  /** description is the description value for this contract. */
  description: string;

  /** icon is the icon value for this contract. */
  icon: string;

  /** color is the color value for this contract. */
  color: string;

  /** status is the current lifecycle state. */
  status: string;

  /**
   * visibility controls discoverability. "private" (default) limits the
   * project to its creator and members; "public" additionally exposes the
   * fields of this message on the owner's profile to business colleagues.
   * Public never grants task access by itself.
   */
  visibility: string;

  /**
   * is_personal marks the caller's auto-provisioned My Tasks bucket. The
   * bucket is hidden from shared lists and exempt from manager-visibility
   * policies unless its owner explicitly shares it.
   */
  is_personal: boolean;

  /**
   * developer_mode marks a software project. When true, clients show the
   * developer features (code links, Development section, dev fields); when
   * false they are hidden but any data they hold is kept.
   */
  developer_mode: boolean;

  /** start_date is the start date value for this contract. */
  start_date: string;

  /** due_date is the due date value for this contract. */
  due_date: string;

  /** created_by is the created by value for this contract. */
  @encode(string)
  created_by: int64;

  /** created_at is when the record was created. */
  created_at: string;

  /** updated_at is when the record was last updated. */
  updated_at: string;

  /** Total task count for this project. Display projection, ignored on writes. */
  task_count: int32;

  /** Number of tasks with status="done". Display projection, ignored on writes. */
  completed_count: int32;

  /** Count of project members, for the "+N" avatar stack. Display projection, ignored on writes. */
  member_count: int32;

  /** Short project code used to build Task.identifier (e.g. "FLW" -> "FLW-42"). */
  code: string;

  /** archived_at is when the project moved to the recycle bin. Empty while active. */
  archived_at?: string;

  /**
   * purge_at is when the recycle bin will permanently delete this project and
   * everything inside it. Stamped at archive time from the business's
   * retention setting; later retention changes affect future archives only.
   */
  purge_at?: string;

  /**
   * visible_via_report is true when the caller can see this project ONLY
   * because they manage someone involved in it (the dept-manager proxy), and
   * not because they created it, own it, or are a member.
   *
   * It exists so a client can group these separately without asking per
   * project why each one is visible. It is a per-caller projection, not a
   * property of the project: the same project is false for its own members and
   * true for their manager. Never persist it.
   */
  visible_via_report: boolean;
}

model ProjectColumn {
  /** id is the unique identifier for the column. */
  @encode(string)
  id: int64;

  /** project_id is the project id value for this contract. */
  @encode(string)
  project_id: int64;

  /** name is the column name. */
  name: string;

  /** color is the column color. */
  color: string;

  /** sort_order is the column sort order. */
  sort_order: int32;

  /** created_at is when the record was created. */
  created_at: string;

  /** updated_at is when the record was last updated. */
  updated_at: string;
}

model ProjectMember {
  /** id is the unique identifier for the membership record. */
  @encode(string)
  id: int64;

  /** project_id is the project id value for this contract. */
  @encode(string)
  project_id: int64;

  /** user_id identifies the user associated with this request. */
  @encode(string)
  user_id: int64;

  /** name is the member display name. */
  name: string;

  /** email is the member email address. */
  email: string;

  /** role is the role value for this contract. */
  role: string;

  /** added_by is the user id of the inviter, if any. */
  @encode(string)
  added_by: int64;

  /** added_at is when the member joined the project. */
  added_at: string;
}

model GetProjectRequest {
  /** id is the unique identifier for the resource. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model GetProjectResponse {
  /** project is the project value for this contract. */
  project: Project;
}

model SetProjectVisibilityRequest {
  /** id is the unique identifier for the resource. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  /** visibility is the target visibility: "private" or "public". */
  @Onkt.required
  visibility: string;
}

model SetProjectVisibilityResponse {
  /** project is the updated project. */
  project: Project;
}

model SetProjectDeveloperModeRequest {
  /** id is the unique identifier for the resource. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  /** enabled turns developer features on (true) or off (false). */
  enabled: boolean;
}

model SetProjectDeveloperModeResponse {
  /** project is the updated project. */
  project: Project;
}

model ListProjectsRequest {
  /**
   * status filters by lifecycle state ("active", "archived", "completed").
   * Empty returns active projects only, matching the sidebar's default view.
   */
  @query("status")
  status: string;

  /**
   * visibility filters by "private" or "public". Empty returns both kinds
   * the caller can reach.
   */
  @query("visibility")
  visibility: string;

  /** q narrows results by name substring. Empty disables the filter. */
  @query("q")
  q: string;

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

model ListProjectsResponse {
  /** projects is the projects value for this contract. */
  projects: Project[];

  pagination: PaginationResponse;
}

model CreateProjectRequest {
  /** name is the name value for this contract. */
  @Onkt.required
  name: string;

  /** description is the description value for this contract. */
  @Onkt.required
  description: string;

  /** icon is the icon value for this contract. */
  @Onkt.required
  icon: string;

  /** color is the color value for this contract. */
  @Onkt.required
  color: string;

  /**
   * visibility is the initial exposure. Empty means "private", which is also
   * the schema default; public makes the project metadata-discoverable via
   * profiles but grants no access by itself.
   */
  visibility: string;

  /** start_date is an optional start date; an empty value leaves it unset. */
  start_date: string;

  /** due_date is an optional due date; an empty value leaves it unset. */
  due_date: string;

  /** developer_mode creates the project as a software project. Defaults to false. */
  developer_mode: boolean;
}

model CreateProjectResponse {
  /** project is the project value for this contract. */
  project: Project;
}

model UpdateProjectBody {
  /** name is the name value for this contract. */
  @Onkt.required
  name: string;

  /** description is the description value for this contract. */
  @Onkt.required
  description: string;

  /** icon is the icon value for this contract. */
  @Onkt.required
  icon: string;

  /** color is the color value for this contract. */
  @Onkt.required
  color: string;

  /** status is the current lifecycle state. */
  @Onkt.required
  status: string;

  /**
   * start_date is an optional start date; an empty value leaves it unset,
   * matching CreateProjectRequest - a project created with no dates (a
   * documented, valid state there) could otherwise never be saved again.
   */
  start_date: string;

  /**
   * due_date is an optional due date; an empty value leaves it unset, for
   * the same reason as start_date above.
   */
  due_date: string;
}

model UpdateProjectRequest {
  /** id is the unique identifier for the resource. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  /** body carries updatable project fields bound from the HTTP payload. */
  @body
  body: UpdateProjectBody;
}

model UpdateProjectResponse {
  /** message is a human-readable result message. */
  message: string;
}

model ArchiveProjectRequest {
  /** id is the unique identifier for the resource. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model ArchiveProjectResponse {
  /**
   * project is the archived project, carrying archived_at and the purge_at
   * deadline computed from the business's retention setting. Clients should
   * surface purge_at wherever they show the archive confirmation.
   */
  project: Project;
}

model ListArchivedProjectsRequest {
  /** The current result page. */
  @query("page")
  page: int32;

  /** The maximum number of records to return. */
  @query("limit")
  limit: int32;
}

model ListArchivedProjectsResponse {
  /**
   * projects lists the caller's visible projects currently sitting in the
   * recycle bin, ordered by soonest purge_at first.
   */
  projects: Project[];

  pagination: PaginationResponse;
}

model RestoreProjectRequest {
  /** id is the unique identifier for the resource. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model RestoreProjectResponse {
  /** project is the restored project with cleared archive timestamps. */
  project: Project;
}

model PurgeProjectRequest {
  /** id is the unique identifier for the resource. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model PurgeProjectResponse {
  /** message is a human-readable result message. */
  message: string;
}

model ListProjectColumnsRequest {
  /** project_id is the project id value for this contract. */
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

model ListProjectColumnsResponse {
  /** columns is the project columns for this contract. */
  columns: ProjectColumn[];

  pagination: PaginationResponse;
}

model CreateProjectColumnRequest {
  /** project_id is the project id value for this contract. */
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  project_id?: int64;

  /** name is the column name. */
  @Onkt.required
  name: string;

  /** color is the column color. */
  color: string;

  /** sort_order is the column sort order. */
  sort_order: int32;
}

model CreateProjectColumnResponse {
  /** column is the created column. */
  column: ProjectColumn;
}

model UpdateProjectColumnRequest {
  /** id is the column identifier. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  /** name is the column name. */
  @Onkt.required
  name: string;

  /** color is the column color. */
  color: string;

  /** sort_order is the column sort order. */
  sort_order: int32;
}

model UpdateProjectColumnResponse {
  /** message is a human-readable result message. */
  message: string;
}

model DeleteProjectColumnRequest {
  /** id is the column identifier. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model DeleteProjectColumnResponse {
  /** message is a human-readable result message. */
  message: string;
}

model MoveTaskToColumnRequest {
  /** task_id is the task identifier. */
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  task_id?: int64;

  /** column_id is the destination column identifier. */
  @encode(string)
  column_id: int64;

  /**
   * ordered_task_ids is the destination column's cards in the order they
   * should appear, including the moved task. When present the server stores
   * that order (flow_tasks.sort_order) in the same transaction as the move,
   * which is how a card dropped between two others stays between them and how
   * a card is reordered inside its own column. Omitted, the move changes the
   * column only and the order is left as it was. Every id must belong to the
   * moved task's project; at most 500 are accepted.
   */
  @Onkt.encode("string")
  ordered_task_ids: int64[];
}

model MoveTaskToColumnResponse {
  /** message is a human-readable result message. */
  message: string;
}

model ListProjectMembersRequest {
  /** project_id is the project id value for this contract. */
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

model ListProjectMembersResponse {
  /** members is the project members for this contract. */
  members: ProjectMember[];

  pagination: PaginationResponse;
}

model AddProjectMemberRequest {
  /** project_id is the project id value for this contract. */
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  project_id?: int64;

  /** user_id identifies the user associated with this request. */
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  user_id?: int64;

  /** role is the role value for this contract. */
  role: string;
}

model AddProjectMemberResponse {
  /** message is a human-readable result message. */
  message: string;
}

model RemoveProjectMemberRequest {
  /** project_id is the project id value for this contract. */
  @Onkt.required
  @minValueExclusive(0)
  @query("project_id")
  @encode(string)
  project_id?: int64;

  /** user_id identifies the user associated with this request. */
  @Onkt.required
  @minValueExclusive(0)
  @query("user_id")
  @encode(string)
  user_id?: int64;
}

model RemoveProjectMemberResponse {
  /** message is a human-readable result message. */
  message: string;
}

model UpdateProjectMemberRoleRequest {
  /** project_id is the project id value for this contract. */
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  project_id?: int64;

  /** user_id identifies the user associated with this request. */
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  user_id?: int64;

  /** role is the role value for this contract. */
  @Onkt.required
  role: string;
}

model UpdateProjectMemberRoleResponse {
  /** message is a human-readable result message. */
  message: string;
}

model InviteProjectMemberRequest {
  /** project_id is the project id value for this contract. */
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  project_id?: int64;

  /** email is the invitee email address. */
  @Onkt.required
  email: string;

  /** role is the role value for this contract. */
  role: string;
}

model InviteProjectMemberResponse {
  /** message is a human-readable result message. */
  message: string;
}

model TransferProjectOwnershipRequest {
  /** project_id is the project id value for this contract. */
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  project_id?: int64;

  /** new_owner_id is the new owner user id. */
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  new_owner_id?: int64;
}

model TransferProjectOwnershipResponse {
  /** message is a human-readable result message. */
  message: string;
}

// ── Labour cost ───────────────────────────────────────────────────
model ProjectLabourCost {
  /** project_id is the project the cost was computed for. */
  @encode(string)
  project_id: int64;

  /**
   * cost_minor_units is the labour cost of billable logged time in the
   * smallest unit of the business currency, so no rounding happens before
   * the caller sees it.
   */
  @encode(string)
  cost_minor_units: int64;

  /**
   * costed_minutes is billable logged time whose contributor had a rate on
   * record; it is the time included in cost_minor_units.
   */
  @encode(string)
  costed_minutes: int64;

  /**
   * uncosted_minutes is billable logged time whose contributor has no rate
   * on record. A total with uncosted_minutes above zero understates true
   * labour cost and must be labelled as partial where it is displayed.
   */
  @encode(string)
  uncosted_minutes: int64;

  /**
   * uncosted_employee_count is how many distinct contributors have no rate
   * on record.
   */
  uncosted_employee_count: int32;
}

model GetProjectLabourCostRequest {
  /** id is the Flow project id path parameter. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model GetProjectLabourCostResponse {
  /** cost is the aggregate for the requested project. */
  cost: ProjectLabourCost;
}
