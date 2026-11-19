import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;
using common.pagination.v1;

namespace flow.task.v2;

/**
 * FlowTask belongs to exactly one project in the caller's business. The workspace
 * layer is gone in v2; "outside a project" is a UI concept only - tasks
 * created without one land in the creator's My Tasks bucket.
 */
model FlowTask {
  /** id is the unique identifier for the resource. */
  @encode(string)
  id: int64;

  /** project_id is the project id value for this contract. */
  @encode(string)
  project_id: int64;

  /** owner_id identifies the tenant that owns the resource. */
  @encode(string)
  owner_id: int64;

  /** parent_task_id is the parent task id value for this contract. */
  @encode(string)
  parent_task_id?: int64;

  /** title is the title value for this contract. */
  title: string;

  /** description is the description value for this contract. */
  description: string;

  /** status is the current lifecycle state. */
  status: string;

  /**
   * The custom status (flow.status.v2) applied to this task, if any. Zero
   * means the task is on its built-in string status alone - not a missing
   * value, since a task with no custom status is the normal case for a
   * project that has not configured any (F-A).
   */
  @encode(string)
  status_id?: int64;

  /**
   * Name of the custom status identified by status_id, denormalised for list
   * views the same way project_name is. Empty when status_id is absent.
   */
  status_name: string;

  /**
   * Color of the custom status identified by status_id, denormalised the
   * same way project_color is. Empty when status_id is absent.
   */
  status_color: string;

  /** priority is the priority value for this contract. */
  priority: string;

  /** assignee_id is the assignee id value for this contract. */
  @encode(string)
  assignee_id?: int64;

  /** reporter_id is the reporter id value for this contract. */
  @encode(string)
  reporter_id: int64;

  /** due_date is the due date value for this contract. */
  due_date?: string;

  /** start_date is the start date value for this contract. */
  start_date?: string;

  /** completed_at is the completed at value for this contract. */
  completed_at?: string;

  /** estimated_hours is the estimated hours value for this contract. */
  estimated_hours?: float64;

  /** actual_hours is the actual hours value for this contract. */
  actual_hours?: float64;

  /** sort_order is the sort order value for this contract. */
  @encode(string)
  sort_order: int64;

  /** labels is the labels value for this contract. */
  @Onkt.encode("string")
  labels: int64[];

  /** created_at is when the record was created. */
  created_at: string;

  /** updated_at is when the record was last updated. */
  updated_at: string;

  /**
   * Human-readable identifier like "FLOW-42" (project code + sequence).
   * Display projection populated by the backend on list/detail reads; safe
   * to leave at zero value on writes.
   */
  identifier: string;

  /** Project name (denormalised from flow_projects.name) for list views. */
  project_name: string;

  /** Project color hex for inline badges on cross-project lists. */
  project_color: string;

  /** Assignee display name. Empty if unassigned. */
  assignee_name: string;

  /** Two-letter uppercase initials derived from assignee_name. Empty if unassigned. */
  assignee_initials: string;

  /** Number of direct subtasks. Powers the "N subtasks" pill on task rows. */
  subtask_count: int32;

  /** Number of completed direct subtasks. */
  completed_subtasks: int32;

  /** Number of task comments. */
  comment_count: int32;

  /** Number of task attachments. */
  attachment_count: int32;

  /** Section bucket within a project (e.g. "Backlog", "Sprint 5"). Empty if unbucketed. */
  @encode(string)
  section_id?: int64;

  /** Cycle (sprint) this task belongs to, if any. Cycles are project-scoped in v2. */
  @encode(string)
  cycle_id?: int64;

  /** Milestone this task rolls up to, if any. */
  @encode(string)
  milestone_id?: int64;

  /**
   * Whether the task is archived. Archived tasks are excluded from default
   * list and board views but stay readable by id. Absent means false, as
   * with every other non-nullable scalar in this message.
   */
  archived: boolean;

  /**
   * Quick effort estimate in minutes. Null when no estimate is set. The
   * write path treats 0 as a clear, so a stored 0 does not occur.
   */
  effort_minutes?: int32;

  /** Free-text blocker label. Empty if nothing is blocking. */
  waiting_on: string;

  /** One-line decision log. Empty if none recorded. */
  decision_note: string;

  /**
   * Whether the *calling* user has pinned this task to their Today view.
   * Resolved per request from flow_task_focus_pins; never a shared task flag.
   */
  focus_pin: boolean;

  /**
   * Task ids this task depends on. Read-only projection; write through the
   * addTaskDependency / removeTaskDependency RPCs.
   */
  @Onkt.encode("string")
  dependencies: int64[];
}

model TaskComment {
  /** id is the unique identifier for the resource. */
  @encode(string)
  id: int64;

  /** task_id is the task id value for this contract. */
  @encode(string)
  task_id: int64;

  /** user_id identifies the user associated with this request. */
  @encode(string)
  user_id: int64;

  /** Display name resolved from the comment author's current profile. */
  actor_name: string;

  /** Two-letter initials resolved from the comment author's current profile. */
  actor_initials: string;

  /** content is the content value for this contract. */
  content: string;

  /** is_edited is the is edited value for this contract. */
  is_edited: boolean;

  /** created_at is when the record was created. */
  created_at: string;

  /** updated_at is when the record was last updated. */
  updated_at: string;
}

model TaskActivity {
  /** id is the unique identifier for the resource. */
  @encode(string)
  id: int64;

  /** task_id is the task id value for this contract. */
  @encode(string)
  task_id: int64;

  /** user_id identifies the user associated with this request. */
  @encode(string)
  user_id: int64;

  /** Display name resolved from the activity author's current profile. */
  actor_name: string;

  /** Two-letter initials resolved from the activity author's current profile. */
  actor_initials: string;

  /** action is the action value for this contract. */
  action: string;

  /** old_value is the old value value for this contract. */
  old_value: string;

  /** new_value is the new value value for this contract. */
  new_value: string;

  /** created_at is when the record was created. */
  created_at: string;
}

model TaskPayloadCompact {
  /** id is the unique identifier for the resource. */
  @encode(string)
  id: int64;

  /** title is the title value for this contract. */
  title: string;

  /** status is the current lifecycle state. */
  status: string;

  /** priority is the priority value for this contract. */
  priority: string;

  /** due_date is the due date value for this contract. */
  due_date: string;

  /** created_at is when the record was created. */
  created_at: string;
}

model Attachment {
  /** id is the unique identifier for the resource. */
  @encode(string)
  id: int64;

  /** task_id is the task id value for this contract. */
  @encode(string)
  task_id: int64;

  /** name is the name value for this contract. */
  name: string;

  /** url is the url value for this contract. */
  url: string;

  /** size is the size value for this contract. */
  @encode(string)
  size: int64;

  /** uploaded_by is the uploaded by value for this contract. */
  @encode(string)
  uploaded_by: int64;

  /** created_at is when the record was created. */
  created_at: string;
}

model TimeLog {
  /** id is the unique identifier for the resource. */
  @encode(string)
  id: int64;

  /** task_id is the task id value for this contract. */
  @encode(string)
  task_id: int64;

  /** user_id identifies the user associated with this request. */
  @encode(string)
  user_id: int64;

  /** duration_minutes is the duration minutes value for this contract. */
  @encode(string)
  duration_minutes: int64;

  /** description is the description value for this contract. */
  description: string;

  /** logged_at is the logged at value for this contract. */
  logged_at: string;

  /** created_at is when the record was created. */
  created_at: string;
}

model ListTasksRequest {
  /**
   * project_id narrows results to one project. Omitted, the call returns
   * every task in the caller's business the visibility policy lets them see,
   * which replaces v1's workspace-wide listing mode.
   */
  @query("project_id")
  @encode(string)
  project_id?: int64;

  /** status is the current lifecycle state. */
  @query("status")
  status: string;

  /** assignee_id is the assignee id value for this contract. */
  @query("assignee_id")
  @encode(string)
  assignee_id?: int64;

  /** priority is the priority value for this contract. */
  @query("priority")
  priority: string;

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

model ListTasksResponse {
  /** tasks is the tasks value for this contract. */
  tasks: FlowTask[];

  pagination: PaginationResponse;
}

model GetMyTasksRequest {
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

model GetMyTasksResponse {
  /** tasks is the tasks value for this contract. */
  tasks: FlowTask[];

  pagination: PaginationResponse;
}

model GetMyTasksProjectRequest {}

model GetMyTasksProjectResponse {
  @encode(string)
  project_id: int64;
}

model SearchTasksRequest {
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

model SearchTasksResponse {
  tasks: FlowTask[];
  pagination: PaginationResponse;
}

model CreateTaskRequest {
  /**
   * project_id files the task into a specific project. OMITTED, the backend
   * resolves (and lazily provisions) the caller's My Tasks bucket - quick-add
   * with nothing but a title must keep working, so this field deliberately
   * carries no \@required. A response always echoes where the task landed.
   */
  @encode(string)
  project_id: int64;

  /** parent_task_id is the parent task id value for this contract. */
  @encode(string)
  parent_task_id?: int64;

  /** title is the title value for this contract. */
  @Onkt.required
  title: string;

  /**
   * description is the description value for this contract. Optional in v2:
   * quick-created tasks start bare and are fleshed out later.
   */
  description: string;

  /** Optional board column within the selected project. */
  @minValueExclusive(0)
  @encode(string)
  section_id?: int64;

  /** Initial built-in lifecycle status; empty uses todo. */
  status: string;

  /** Effort estimate in minutes; omitted means no estimate. */
  @minValue(0)
  effort_minutes?: int32;

  /**
   * priority is the priority value for this contract. Optional in v2; empty
   * means the backend default rather than a client-mandated value.
   */
  priority: string;

  /**
   * assignee_id is the assignee id value for this contract. Assigning grants
   * the assignee project membership when they lack it.
   */
  @encode(string)
  assignee_id?: int64;

  /** due_date is the due date value for this contract. */
  due_date: string;

  /** start_date is the start date value for this contract. */
  start_date: string;

  /** estimated_hours is the estimated hours value for this contract. */
  estimated_hours?: float64;

  /**
   * labels is the labels value for this contract. Optional in v2: tasks may
   * start unlabelled and be tagged later from detail views.
   */
  @Onkt.encode("string")
  labels: int64[];

  /** Milestone to file the task under; must belong to the task's project. */
  @minValueExclusive(0)
  @encode(string)
  milestone_id?: int64;
}

model CreateTaskResponse {
  /** task is the task value for this contract. */
  task: FlowTask;
}

model GetTaskDetailRequest {
  /** id is the unique identifier for the resource. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model GetTaskDetailResponse {
  /** task is the task value for this contract. */
  task: FlowTask;

  /** subtasks is the subtasks value for this contract. */
  subtasks: FlowTask[];

  /** comments is the comments value for this contract. */
  comments: TaskComment[];

  /** activity is the activity value for this contract. */
  activity: TaskActivity[];
}

model UpdateTaskBody {
  /**
   * Every field here is optional on purpose. This is a partial update: the
   * client sends only what changed, and the repository preserves omitted
   * columns when absent. Empty descriptions/dates and assignee_id = 0 clear
   * their fields. Marking a field \@required here rejects a
   * single-field patch with a 400 before the handler ever runs, which is
   * exactly the defect this contract previously had.
   *
   * title is the title value for this contract.
   */
  title: string;

  /** description is the description value for this contract. */
  description?: string;

  /** status is the current lifecycle state. */
  status: string;

  /** priority is the priority value for this contract. */
  priority: string;

  /** assignee_id is the assignee id value for this contract. */
  @encode(string)
  assignee_id?: int64;

  /** due_date is the due date value for this contract. */
  due_date?: string;

  /** start_date is the start date value for this contract. */
  start_date?: string;

  /**
   * labels is the labels value for this contract. Absent leaves labels
   * untouched; \@min_items is deliberately absent because requiring one label
   * made every task without labels permanently unupdatable.
   */
  @Onkt.encode("string")
  labels: int64[];

  /** Board column; omitted keeps it, zero clears it. */
  @encode(string)
  section_id?: int64;

  /**
   * Updated archived status; absent to keep current value. True archives the
   * task, false restores it.
   */
  archived?: boolean;

  /**
   * Quick effort estimate in minutes for the effort chips. Absent leaves the
   * value untouched; 0 clears it. Zero minutes is meaningless as a stored
   * value, so it is safe as the clear sentinel, and it keeps this field out
   * of ambiguity between an omitted value and an explicit clear.
   */
  effort_minutes?: int32;

  /**
   * Estimated hours. Absent leaves untouched; 0 clears. Same sentinel rule as
   * effort_minutes.
   */
  estimated_hours?: float64;

  /**
   * Free-text blocker label: person, team or approval. Absent leaves
   * untouched; empty string clears.
   */
  waiting_on?: string;

  /** One-line decision log. Absent leaves untouched; empty string clears. */
  decision_note?: string;

  /**
   * Pin this task into the caller's own Today view. Absent leaves untouched;
   * true pins, false unpins. Per-caller, not per-task: Today is a personal
   * view, so this writes flow_task_focus_pins rather than a task column.
   */
  focus_pin?: boolean;

  /**
   * Milestone the task rolls up to. Absent leaves it untouched; 0 removes the
   * task from its milestone. Must belong to the task's project.
   */
  @encode(string)
  milestone_id?: int64;
}

model UpdateTaskRequest {
  /** id is the unique identifier for the resource. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  /** body carries updatable task fields bound from the HTTP payload. */
  @body
  body: UpdateTaskBody;
}

model UpdateTaskResponse {
  /** message is a human-readable result message. */
  message: string;
}

model UpdateTaskStatusBody {
  /** status is the current lifecycle state. */
  @Onkt.required
  status: string;
}

model UpdateTaskStatusRequest {
  /** id is the unique identifier for the resource. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  /** body carries the new task status bound from the HTTP payload. */
  @body
  body: UpdateTaskStatusBody;
}

model UpdateTaskStatusResponse {
  /** message is a human-readable result message. */
  message: string;
}

model DeleteTaskRequest {
  /** id is the unique identifier for the resource. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model DeleteTaskResponse {
  /** message is a human-readable result message. */
  message: string;
}

model CreateSubtaskRequest {
  /**
   * parent_task_id is the parent task id value for this contract. The new
   * subtask inherits the parent's project; v1 made clients pass it again,
   * which only ever invited mismatches.
   */
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  parent_task_id?: int64;

  /** title is the title value for this contract. */
  @Onkt.required
  title: string;

  /** description is the description value for this contract. */
  description: string;

  /** priority is the priority value for this contract. */
  priority: string;

  /** assignee_id is the assignee id value for this contract. */
  @encode(string)
  assignee_id?: int64;

  /** due_date is the due date value for this contract. */
  due_date: string;

  /** start_date is the start date value for this contract. */
  start_date: string;

  /** estimated_hours is the estimated hours value for this contract. */
  estimated_hours?: float64;
}

model CreateSubtaskResponse {
  /** task_id is the task id value for this contract. */
  @encode(string)
  task_id: int64;

  /** message is a human-readable result message. */
  message: string;
}

model GetSubtasksRequest {
  /** task_id is the task id value for this contract. */
  @minValueExclusive(0)
  @encode(string)
  task_id: int64;

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

model GetSubtasksResponse {
  /** subtasks is the subtasks value for this contract. */
  subtasks: TaskPayloadCompact[];

  pagination: PaginationResponse;
}

model AddCommentRequest {
  /** task_id is the task id value for this contract. */
  @minValueExclusive(0)
  @encode(string)
  task_id: int64;

  /** content is the content value for this contract. */
  @Onkt.required
  content: string;
}

model AddCommentResponse {
  /** comment is the comment value for this contract. */
  comment: TaskComment;
}

model DeleteCommentRequest {
  /** id is the unique identifier for the resource. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model DeleteCommentResponse {
  /** message is a human-readable result message. */
  message: string;
}

model UpdateCommentRequest {
  /** id is the comment being reworded; only its author may change it. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  /** content is the comment's new text. */
  @Onkt.required
  content: string;
}

model UpdateCommentResponse {
  /** comment is the stored comment after the edit. */
  comment: TaskComment;
}

model GetTaskActivityRequest {
  /** task_id is the task id value for this contract. */
  @minValueExclusive(0)
  @encode(string)
  task_id: int64;
}

model GetTaskActivityResponse {
  /** activity is the activity value for this contract. */
  activity: TaskActivity[];
}

model CreateAttachmentRequest {
  /** task_id identifies the task the attachment belongs to. */
  @minValueExclusive(0)
  @encode(string)
  task_id: int64;

  /** name is the name value for this contract. */
  @Onkt.required
  name: string;

  /** url is the url value for this contract. */
  @Onkt.required
  url: string;

  /** size is the size value for this contract. */
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  size?: int64;
}

model CreateAttachmentResponse {
  /** attachment_id is the attachment id value for this contract. */
  @encode(string)
  attachment_id: int64;

  /** message is a human-readable result message. */
  message: string;
}

model GetAttachmentsRequest {
  /** task_id is the task id value for this contract. */
  @minValueExclusive(0)
  @encode(string)
  task_id: int64;
}

model GetAttachmentsResponse {
  /** attachments is the attachments value for this contract. */
  attachments: Attachment[];
}

model DeleteAttachmentRequest {
  /** id is the unique identifier for the resource. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model DeleteAttachmentResponse {
  /** message is a human-readable result message. */
  message: string;
}

model CreateTimeLogRequest {
  /** duration_minutes is the duration minutes value for this contract. */
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  duration_minutes?: int64;

  /** description is the description value for this contract. */
  @Onkt.required
  description: string;

  /** task_id identifies the task the time log belongs to. */
  @minValueExclusive(0)
  @encode(string)
  task_id: int64;
}

model CreateTimeLogResponse {
  /** time_log_id is the time log id value for this contract. */
  @encode(string)
  time_log_id: int64;

  /** message is a human-readable result message. */
  message: string;
}

model GetTimeLogsRequest {
  /** task_id is the task id value for this contract. */
  @minValueExclusive(0)
  @encode(string)
  task_id: int64;
}

model GetTimeLogsResponse {
  /** time_logs is the time logs value for this contract. */
  time_logs: TimeLog[];
}

model DeleteTimeLogRequest {
  /** id is the time log identifier. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model DeleteTimeLogResponse {
  /** message is a human-readable result message. */
  message: string;
}

model AddTaskDependencyBody {
  /** depends_on_task_id is the task this task is blocked by. */
  @minValueExclusive(0)
  @encode(string)
  depends_on_task_id: int64;
}

model AddTaskDependencyRequest {
  /** task_id is the task gaining the dependency. */
  @minValueExclusive(0)
  @encode(string)
  task_id: int64;

  /** body carries the dependency target bound from the HTTP payload. */
  @body
  body: AddTaskDependencyBody;
}

model AddTaskDependencyResponse {
  /** message is a human-readable result message. */
  message: string;
}

model RemoveTaskDependencyRequest {
  /** task_id is the task losing the dependency. */
  @minValueExclusive(0)
  @encode(string)
  task_id: int64;

  /** depends_on_task_id is the dependency being removed. */
  @minValueExclusive(0)
  @encode(string)
  depends_on_task_id: int64;
}

model RemoveTaskDependencyResponse {
  /** message is a human-readable result message. */
  message: string;
}

model TaskLink {
  /** id is the unique identifier for the resource. */
  @encode(string)
  id: int64;

  /** task_id is the task the link belongs to. */
  @encode(string)
  task_id: int64;

  /** url is the link as added: always http or https. */
  url: string;

  /**
   * kind is the recognised link type: github_pr, github_issue, gitlab_mr,
   * gitlab_issue, figma, vercel, netlify, sentry, google_docs or generic.
   */
  kind: string;

  /**
   * title is the fetched page or PR title. Empty when it could not be read,
   * for example a private repository.
   */
  title: string;

  /**
   * ref is a short label derived from the URL, such as "rtylrs/rtylr#1615"
   * or the host name.
   */
  ref: string;

  /** state is the PR or issue state when known: open, closed, merged or draft. */
  state: string;

  /** added_by is the user who added the link. */
  @encode(string)
  added_by: int64;

  /** created_at is when the link was added. */
  created_at: string;
}

model CreateTaskLinkRequest {
  /** task_id is the task to add the link to. */
  @minValueExclusive(0)
  @encode(string)
  task_id: int64;

  /** url is the http or https URL to add. */
  @Onkt.required
  url: string;
}

model CreateTaskLinkResponse {
  /** link is the added link, or the existing one when the task already had it. */
  link: TaskLink;
}

model ListTaskLinksRequest {
  /** task_id is the task whose links to list. */
  @minValueExclusive(0)
  @encode(string)
  task_id: int64;
}

model ListTaskLinksResponse {
  /** links are the task's links, newest first. */
  links: TaskLink[];
}

model DeleteTaskLinkRequest {
  /** id is the link to remove. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model DeleteTaskLinkResponse {
  /** message is a human-readable result message. */
  message: string;
}
