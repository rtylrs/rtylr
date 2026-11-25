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

  /**
   * Cycle (sprint) this task belongs to, if any. Cycles are project-scoped in
   * v2. It is the task's upcoming or active cycle, or the completed cycle it was
   * finished in. Absent means the task is in the unplanned backlog.
   */
  @encode(string)
  cycle_id?: int64;

  /** Name of that cycle. Empty when cycle_id is absent. */
  cycle_name: string;

  /**
   * Position of the task inside its cycle, or inside the unplanned backlog when
   * cycle_id is absent. Lower comes first. Absent for a backlog task that was
   * never placed; those follow the ranked ones, oldest first.
   */
  @encode(string)
  plan_rank?: int64;

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
   * addTaskDependency / removeTaskDependency RPCs. Only "blocks" relations:
   * the tasks that block this one. Other relation types are read through
   * listTaskRelations.
   */
  @Onkt.encode("string")
  dependencies: int64[];

  /**
   * Whether at least one task blocking this one is still open, meaning its
   * status bucket is neither done nor cancelled. Read-only projection.
   */
  blocked: boolean;

  /** Number of open tasks blocking this one. Zero when blocked is false. */
  blocker_count: int32;

  /**
   * Custom field values keyed by field id (flow.field.v2). Only fields with a
   * value appear, so a task without any carries an empty map. Read-only
   * projection on list and detail reads; write through setTaskFieldValues.
   */
  field_values: Record<TaskFieldValue>;

  /**
   * What kind of work this is: task, bug, feature, chore, spike or epic. Every task that
   * predates types reads as task.
   */
  task_type: string;

  /**
   * The epic this task sits under, if any. A subtask reports its parent's
   * epic. Absent on an epic and on a task with no epic. Write it through
   * setTaskParent, parent_task_id on update, or parent_task_id on create.
   *
   * This is separate from parent_task_id on purpose: parent_task_id is set
   * only on a subtask, so a task under an epic is still a top-level task.
   */
  @encode(string)
  epic_id?: int64;

  /** Title of the epic identified by epic_id. Empty when epic_id is absent. */
  epic_title: string;

  /**
   * On an epic, its own colour. On any other task, the colour of the epic
   * identified by epic_id. Empty when there is none.
   */
  epic_color: string;
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

  /**
   * blocked narrows results to tasks that are blocked (true) or not blocked
   * (false). Omitted, both are returned.
   */
  @query("blocked")
  blocked?: boolean;

  /** The current result page. */
  @query("page")
  page: int32;

  /** The maximum number of records to return. */
  @query("limit")
  limit: int32;

  /**
   * The field to sort results by. "field:<id>" sorts by a custom field of the
   * project named by project_id; tasks without a value come last. Every type
   * sorts except multi_select and person. Anything else keeps the board order.
   */
  @query("sort_by")
  sort_by: string;

  /** The sort direction ("asc" or "desc"). */
  @query("sort_dir")
  sort_dir: string;

  /**
   * field_filter narrows the listing by custom field values. It is a JSON
   * array of clauses that must all match, each
   * `{"field_id":"12","op":"any","values":["a","b"]}`, and needs project_id.
   *
   * Operators by field type: `eq` (text, url, number, date, select,
   * checkbox), `any` (select, multi_select, person: the task holds at least
   * one of the values), `contains` (text, url), `range` (number, date: values
   * are [min, max], either may be empty for an open end). Values use the
   * encoding of `TaskFieldValue`. A checkbox `eq` "false" also matches tasks
   * that never set the field.
   */
  @query("field_filter")
  field_filter: string;

  /**
   * task_type narrows results to one type, or several separated by commas
   * ("task,bug,feature" lists everything except epics).
   */
  @query("task_type")
  task_type: string;

  /** epic_id narrows results to the tasks directly under one epic. */
  @query("epic_id")
  @encode(string)
  epic_id?: int64;

  /**
   * no_epic narrows results to tasks that sit under no epic. Epics themselves
   * are left out. Ignored when epic_id is set.
   */
  @query("no_epic")
  no_epic: boolean;
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

  /**
   * parent_task_id is the epic or the task to create this under. An epic
   * makes the new task one of the epic's tasks (it comes back with epic_id
   * set and no parent_task_id); a task makes it a subtask. The parent must be
   * in the same project, must not itself be a subtask, and an epic cannot be
   * given one.
   */
  @encode(string)
  parent_task_id?: int64;

  /** task_type is task, bug, feature, chore, spike or epic; empty means task. */
  task_type: string;

  /** epic_color is the colour of a new epic. Rejected for any other type. */
  epic_color: string;

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

  /**
   * field_values sets custom field values on the new task. A field of the
   * project that is left out receives its default, and a required field with
   * neither a value nor a default rejects the request with 400.
   */
  field_values: TaskFieldValueInput[];

  /**
   * Upcoming or active cycle of the task's project to add the task to, at the
   * end of the cycle. Absent leaves the task in the unplanned backlog.
   */
  @minValueExclusive(0)
  @encode(string)
  cycle_id?: int64;
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

  /**
   * New type: task, bug, feature, chore, spike or epic. Empty leaves it untouched. A task
   * that has a parent, an epic or subtasks cannot become an epic, and an epic
   * that still has tasks cannot stop being one.
   */
  task_type: string;

  /**
   * Colour of an epic. Absent leaves untouched; empty string clears. Rejected
   * on a task that is not an epic.
   */
  epic_color?: string;

  /**
   * Moves the task under an epic or a task in the same project. Absent leaves
   * it where it is; 0 clears both links, leaving a top-level task with no
   * epic. Same rules as setTaskParent.
   */
  @encode(string)
  parent_task_id?: int64;

  /**
   * Cycle to move the task to, at the end of that cycle. Absent leaves it
   * untouched; 0 returns the task to the unplanned backlog. Must be an upcoming
   * or active cycle of the task's project.
   */
  @encode(string)
  cycle_id?: int64;
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

  /**
   * How many tasks were detached from a deleted epic. They are kept, with no
   * epic. Zero when the deleted task was not an epic.
   */
  detached_children: int32;
}

model SetTaskParentBody {
  /**
   * parent_id is the epic or the task to move this task under. An epic makes
   * it one of the epic's tasks; a task makes it a subtask. Absent or zero
   * clears both, leaving a top-level task with no epic.
   */
  @encode(string)
  parent_id?: int64;
}

model SetTaskParentRequest {
  /** id is the task being moved. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  /** body carries the new parent bound from the HTTP payload. */
  @body
  body: SetTaskParentBody;
}

model SetTaskParentResponse {
  /**
   * task is the task after the move, with parent_task_id and the epic fields
   * as they now stand.
   */
  task: FlowTask;
}

/**
 * EpicRollup summarises the tasks directly under one epic. Archived tasks are
 * left out of every number.
 */
model EpicRollup {
  /** epic_id is the epic's task id. */
  @encode(string)
  epic_id: int64;

  /** Human-readable identifier of the epic, like "T-42". */
  identifier: string;

  /** title is the epic's title. */
  title: string;

  /** color is the epic's colour. Empty when none is set. */
  color: string;

  /** status is the epic's own built-in lifecycle status. */
  status: string;

  /** Number of tasks directly under the epic. Subtasks are not counted. */
  total_tasks: int32;

  /** Direct tasks whose status is in the backlog bucket. */
  backlog_tasks: int32;

  /** Direct tasks whose status is in the active bucket. */
  active_tasks: int32;

  /** Direct tasks whose status is in the done bucket. */
  done_tasks: int32;

  /** Direct tasks whose status is in the cancelled bucket. */
  cancelled_tasks: int32;

  /**
   * Sum of the direct tasks' estimates in minutes. A task's effort_minutes
   * is used when set, otherwise its estimated_hours. Subtask estimates are
   * not added, since a subtask breaks down its parent's estimate.
   */
  estimate_minutes: int32;

  /** The part of estimate_minutes that sits on tasks in the done bucket. */
  done_estimate_minutes: int32;

  /** Minutes logged on the direct tasks and on their subtasks. */
  logged_minutes: int32;
}

model ListEpicRollupsRequest {
  /** project_id is the project whose epics to summarise. */
  @Onkt.required
  @minValueExclusive(0)
  @query("project_id")
  @encode(string)
  project_id?: int64;
}

model ListEpicRollupsResponse {
  /**
   * epics lists every unarchived epic in the project, in board order,
   * including epics with no tasks yet.
   */
  epics: EpicRollup[];
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

/**
 * RelatedTask is the task at the other end of a relation, with what a row in
 * the relations panel shows.
 */
model RelatedTask {
  /** relation_id identifies the relation; pass it to removeTaskRelation. */
  @encode(string)
  relation_id: int64;

  /** task_id is the other task. */
  @encode(string)
  task_id: int64;

  /** Human-readable identifier of the other task, like "T-42". */
  identifier: string;

  /** title is the other task's title. */
  title: string;

  /**
   * project_id is the other task's project. Relations may cross projects
   * inside one business.
   */
  @encode(string)
  project_id: int64;

  /** project_name is the other task's project name. */
  project_name: string;

  /** status is the other task's built-in string status. */
  status: string;

  /**
   * status_name is the other task's project status name. Empty when the task
   * has no project status.
   */
  status_name: string;

  /** status_color is the colour of that project status. Empty when absent. */
  status_color: string;

  /** status_bucket is backlog, active, done or cancelled. */
  status_bucket: string;

  /** assignee_id is the other task's assignee, if any. */
  @encode(string)
  assignee_id?: int64;

  /** Assignee display name. Empty if unassigned. */
  assignee_name: string;

  /** Two-letter initials derived from assignee_name. Empty if unassigned. */
  assignee_initials: string;
}

/**
 * TaskRelationGroup is every relation of one type and direction, read from
 * the listed task's side.
 */
model TaskRelationGroup {
  /** type is blocks, relates, duplicates or clones. */
  type: string;

  /**
   * direction is outward when the listed task is the subject ("this task
   * blocks", "duplicates", "clones") and inward when it is the object ("is
   * blocked by", "is duplicated by", "is cloned by"). relates is symmetric
   * and always outward.
   */
  direction: string;

  /**
   * relation is type and direction as one key for labels: blocks,
   * blocked_by, relates_to, duplicates, duplicated_by, clones or cloned_by.
   */
  relation: string;

  /** tasks are the related tasks the caller may see, oldest relation first. */
  tasks: RelatedTask[];
}

model AddTaskRelationBody {
  /**
   * other_task_id is the task at the other end. It must be in the caller's
   * business and visible to the caller; it may be in another project.
   */
  @minValueExclusive(0)
  @encode(string)
  other_task_id: int64;

  /** type is blocks, relates, duplicates or clones. */
  @Onkt.required
  type: string;

  /**
   * direction is outward (default) or inward, read from task_id's side:
   * outward with blocks means task_id blocks other_task_id, inward means
   * task_id is blocked by other_task_id. Ignored for relates.
   */
  direction: string;
}

model AddTaskRelationRequest {
  /** task_id is the task the relation is added from. */
  @minValueExclusive(0)
  @encode(string)
  task_id: int64;

  /** body carries the relation bound from the HTTP payload. */
  @body
  body: AddTaskRelationBody;
}

model AddTaskRelationResponse {
  /** relation_id identifies the stored relation. */
  @encode(string)
  relation_id: int64;

  /** message is a human-readable result message. */
  message: string;
}

model ListTaskRelationsRequest {
  /** task_id is the task whose relations to list. */
  @minValueExclusive(0)
  @encode(string)
  task_id: int64;
}

model ListTaskRelationsResponse {
  /**
   * groups holds one entry per type and direction that has at least one
   * relation, in a fixed order: blocks, blocked_by, relates_to, duplicates,
   * duplicated_by, clones, cloned_by.
   */
  groups: TaskRelationGroup[];
}

model RemoveTaskRelationRequest {
  /** task_id is either end of the relation. */
  @minValueExclusive(0)
  @encode(string)
  task_id: int64;

  /** relation_id is the relation to remove. */
  @minValueExclusive(0)
  @encode(string)
  relation_id: int64;
}

model RemoveTaskRelationResponse {
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

/**
 * TaskFieldValue is a task's value for one custom field.
 *
 * Every value travels as a string so one shape serves all eight field types:
 * text and url as written, number as a decimal ("12.5"), date as an ISO date
 * ("2026-11-19"), checkbox as "true" or "false", select as the option id.
 */
model TaskFieldValue {
  /** value is the value of a single-value field. Empty for multi_select and person. */
  value: string;

  /**
   * values is the value of a multi_select (option ids) or person (user ids)
   * field. Empty for every other type.
   */
  values: string[];
}

model TaskFieldValueInput {
  /** field_id is the custom field to write. It must belong to the task's project. */
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  field_id?: int64;

  /** value is the new value for a single-value field. */
  value: string;

  /** values is the new value for a multi_select or person field. */
  values: string[];
}

model SetTaskFieldValuesRequest {
  /** task_id is the task to write to. */
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  task_id?: int64;

  /**
   * values lists the fields to write, at most 50. An entry with an empty value
   * and no values clears that field. Fields left out keep their value. The
   * whole batch is applied or none of it.
   */
  @Onkt.required
  values: TaskFieldValueInput[];
}

model SetTaskFieldValuesResponse {
  /**
   * field_values is the task's full set of values after the write, keyed by
   * field id.
   */
  field_values: Record<TaskFieldValue>;
}

model DuplicateTaskRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  /** title for the copy; empty keeps the original title. */
  @maxLength(500)
  title?: string;
}

model DuplicateTaskResponse {
  task: FlowTask;
}

/** ImportTaskRow is one task to create. Ids are resolved by the client from names and emails. */
model ImportTaskRow {
  @Onkt.required
  @maxLength(500)
  title: string;

  @maxLength(20000)
  description?: string;

  @encode(string)
  assignee_id?: int64;

  /** due_date is YYYY-MM-DD. */
  due_date?: string;

  /** priority is none, low, medium, high or urgent. */
  priority?: string;

  @minValue(0)
  effort_minutes?: int32;

  /** section_id is the status (board column) of the project to put the task in. */
  @encode(string)
  section_id?: int64;

  /** subtasks are titles of subtasks to create under the task. */
  @maxItems(50)
  subtasks?: string[];
}

model ImportTasksRequest {
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  project_id: int64;

  @Onkt.required
  @maxItems(500)
  rows: ImportTaskRow[];
}

/** ImportRowError says why one row was refused. row is 1-based, in the order sent. */
model ImportRowError {
  row: int32;
  message: string;
}

model ImportTasksResponse {
  /** created is how many tasks were created, not counting subtasks. */
  created: int32;

  /** task_ids are the new tasks, in row order. */
  @Onkt.encode("string")
  task_ids: int64[];

  /** errors lists the invalid rows. When it is not empty nothing was created. */
  errors: ImportRowError[];
}

/** TaskTemplate is the description a project fills in for new tasks of one type (T-380). */
model TaskTemplate {
  @encode(string)
  id: int64;

  @encode(string)
  project_id: int64;

  /** task_type is task, bug, feature, chore, spike or epic. */
  task_type: string;

  /** body is the Markdown description a new task of this type starts with. */
  body: string;

  /** updated_at is when the template last changed, RFC 3339. */
  updated_at: string;
}

model ListTaskTemplatesRequest {
  /** project_id is the project whose templates to list. */
  @Onkt.required
  @minValueExclusive(0)
  @query("project_id")
  @encode(string)
  project_id?: int64;
}

model ListTaskTemplatesResponse {
  templates: TaskTemplate[];
}

model SetTaskTemplateRequest {
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  project_id: int64;

  @Onkt.required
  task_type: string;

  /** body is the template; empty removes the project's template for the type. */
  @maxLength(20000)
  body: string;
}

model SetTaskTemplateResponse {
  /** template is the saved template; absent when the body was empty and it was removed. */
  template?: TaskTemplate;
}
