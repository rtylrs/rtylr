import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;
using common.pagination.v1;

namespace flow.project.v1;

model Project {
  /** id is the unique identifier for the resource. */
  @encode(string)
  id: int64;

  /** workspace_id is the workspace id value for this contract. */
  @encode(string)
  workspace_id: int64;

  /** owner_id identifies the tenant that owns the resource. */
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
}

model ProjectColumn {
  /** id is the unique identifier for the column. */
  @encode(string)
  id: int64;

  /** project_id is the project id value for this contract. */
  @encode(string)
  project_id: int64;

  /** workspace_id is the workspace id value for this contract. */
  @encode(string)
  workspace_id: int64;

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

model ListProjectsRequest {
  /** workspace_id is the workspace id value for this contract. */
  @query("workspace_id")
  @encode(string)
  workspace_id: int64;

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
  /** workspace_id is the workspace id value for this contract. */
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  workspace_id?: int64;

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

  /** start_date is an optional start date; an empty value leaves it unset. */
  start_date: string;

  /** due_date is an optional due date; an empty value leaves it unset. */
  due_date: string;
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

  /** start_date is the start date value for this contract. */
  @Onkt.required
  start_date: string;

  /** due_date is the due date value for this contract. */
  @Onkt.required
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

model DeleteProjectRequest {
  /** id is the unique identifier for the resource. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model DeleteProjectResponse {
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

  /** workspace_id is the workspace id value for this contract. */
  @encode(string)
  workspace_id: int64;

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

model MoveProjectRequest {
  /** project_id is the project id value for this contract. */
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  project_id?: int64;

  /** workspace_id is the destination workspace id. */
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  workspace_id?: int64;
}

model MoveProjectResponse {
  /** message is a human-readable result message. */
  message: string;
}

/**
 * ProjectLabourCost is the live labour cost of one project over all logged
 * time. It reports aggregate money only and deliberately carries no
 * per-person rows: hourly rates stay inside HR. Cost covers billable time;
 * time whose contributor has no rate on record is reported separately as
 * uncosted rather than silently folded in as zero-cost work.
 */
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

/**
 * GetProjectLabourCostRequest asks for one project's labour cost. The id is
 * the Flow project id taken from the path.
 */
model GetProjectLabourCostRequest {
  /** id is the Flow project id path parameter. */
  @encode(string)
  id: int64;
}

/** GetProjectLabourCostResponse returns the project's live cost aggregate. */
model GetProjectLabourCostResponse {
  /** cost is the aggregate for the requested project. */
  cost: ProjectLabourCost;
}
