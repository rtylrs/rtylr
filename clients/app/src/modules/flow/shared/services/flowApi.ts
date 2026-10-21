import type { FlowTask as ContractTask } from '@voxire/voxie/flow/task/v2/types';

import { oneKitResult } from '@voxire/api';
import {
  businessClient,
  notificationClient,
  settingClient,
  userClient,
  projectV2,
  taskV2,
  labelV2,
  viewV2,
  insightsV2,
  milestoneV2,
  cycleV2,
  statusClient,
} from '../lib/clients';

// -- Types ----------------------------------------------------------------

export interface ApiNotification {
  id: string;
  type: 'assigned' | 'mention' | 'comment' | 'status_change' | 'due_soon';
  title: string;
  description: string;
  taskId?: string;
  taskTitle: string;
  actorInitials: string;
  project: string;
  projectColor: string;
  read: boolean;
  timestamp: string;
}

export interface ApiUserSettings {
  taskAssigned: boolean;
  taskComments: boolean;
  mentions: boolean;
  statusChanges: boolean;
  weeklyDigest: boolean;
  emailNotifications: boolean;
  sidebarPreferences?: string;
}

export interface ApiBusinessSummary {
  id: string | number;
  name: string;
  isActive?: boolean;
}

export interface ApiProject {
  id: string;
  name: string;
  color: string;
  description?: string;
  icon?: string;
  status?: string;
  startDate?: string;
  dueDate?: string;
  taskCount: number;
  completedCount: number;
  memberAvatars: string[];
  // Per-caller projection from flow.project.v2: set only when the caller sees
  // the project through the dept-manager proxy rather than as creator, owner or
  // member. Optional because the v1 shape has no such field.
  visibleViaReport?: boolean;
  // Stamped when the project is archived. purgeAt is the scheduled permanent
  // deletion, computed server-side from the business's retention setting, and
  // ruling P3 requires it to be stated explicitly rather than described as
  // "when retention expires".
  archivedAt?: string;
  purgeAt?: string;
}

export type ProjectUpdate = Partial<{
  name: string;
  color: string;
  description: string;
  icon: string;
  status: string;
  startDate: string;
  dueDate: string;
}>;

export interface ApiPagination {
  page?: number;
  limit?: number;
  total?: number;
  totalPages?: number;
  hasNext?: boolean;
  hasPrev?: boolean;
}

export interface ApiTask {
  id: string;
  projectId: string;
  title: string;
  description?: string;
  status: 'todo' | 'in_progress' | 'in_review' | 'done' | 'cancelled' | 'backlog';
  priority: 'urgent' | 'high' | 'medium' | 'low' | 'no_priority';
  assignee: string;
  assigneeName?: ContractTask['assigneeName'];
  assigneeId?: string;
  assigneeInitials: string;
  dueDate: string;
  // Write path only for now. The wire carries start_date on both Task and
  // UpdateTaskBody, but the response mapping never populates it, so reading
  // this off a fetched task still yields undefined (see ProjectTimeline).
  // It exists here so updateTask can forward a caller-supplied value instead
  // of hardcoding a placeholder.
  startDate?: string;
  subtaskCount: number;
  completedSubtasks?: number;
  commentCount?: number;
  attachmentCount?: number;
  projectName: string;
  estimate?: number;
  estimatedHours?: number | null;
  actualHours?: number | null;
  // Pinned to the generated contract rather than re-declared. ApiTask is
  // hand-rolled and drifts from the wire (see apiTaskToTask on the numeric
  // ids); swapping the whole interface for the generated Task is the Phase 3
  // migration, but `archived` is the field this archive work turns on, so it
  // tracks voxie directly - if the contract drops it or makes it required,
  // this file stops compiling instead of silently disagreeing.
  archived?: ContractTask['archived'];
  // F-A: the custom status (flow.status.v2) applied to this task, if any.
  // Absent means the task is on its built-in string status alone. Pinned to
  // the generated contract for the same reason as archived above.
  statusId?: ContractTask['statusId'];
  statusName?: ContractTask['statusName'];
  statusColor?: ContractTask['statusColor'];
  identifier?: string;
  createdAt?: string;
  updatedAt?: string;
  labels?: string[];
  dependencies?: string[];
  sectionId?: string;
  cycleId?: string;
  milestoneId?: string;
  // UI-only fields stored in local store
  effortMinutes?: number;
  waitingOn?: string;
  decisionNote?: string;
  focusPin?: boolean;
  customFieldValues?: Record<string, string | number>;
}

export interface ApiLabel {
  id: number;
  name: string;
  color: string;
  createdAt: string;
}

export interface ApiStatus {
  id: string;
  projectId: string;
  name: string;
  bucket: 'backlog' | 'active' | 'done' | 'cancelled';
  legacyStatus: string;
  color: string;
  sortOrder: number;
  isDefault: boolean;
  createdAt: string;
}

export interface ApiMilestone {
  id: string;
  projectId: string;
  name: string;
  targetDate: string;
  progress: number;
  computedProgress: number;
  doneCount: number;
  taskCount: number;
  status: 'planning' | 'active' | 'at_risk' | 'completed';
  description?: string;
  project?: { name: string; color: string };
}

export interface ApiCycle {
  id: string;
  name: string;
  startDate: string;
  endDate: string;
  status: 'upcoming' | 'active' | 'completed';
  scope: number;
  completed: number;
}

export interface ApiProjectColumn {
  /**
   * An int64 on the wire (flow/project/v2 `id: int64`), so a JSON number.
   *
   * This was annotated `string`, which the contract and the Go handler's
   * `Id int64` both contradict, and which hid a real defect: comparing it to a
   * task's stringified `sectionId` matched `"5"` against `5` - always false -
   * so every task fell out of its column. Callers that need a string key for
   * the board's grouping or for the request payload coerce explicitly.
   */
  id: number;
  projectId: string;
  name: string;
  color: string;
  sortOrder: number;
  createdAt: string;
  updatedAt: string;
}

export interface ApiView {
  id: string;
  name: string;
  icon: string;
  // App-facing shape. The wire contract (flow/view/v1) encodes this as a JSON
  // string - see toViewPayload/toViewUpdatePayload for the encode side and
  // decodeApiView for the decode side.
  filters: Record<string, unknown>;
  groupBy?: string;
  sortBy?: string;
  sortDir?: string;
  visibility?: 'shared' | 'private';
}

export interface ApiInsights {
  velocity: { cycle: string; completed: number; total: number; completionRate: number }[];
  burndown: { day: number; ideal: number; actual: number }[];
  cycleTime: number;
  timeByMember: { id: string; name: string; hours: number }[];
  throughput: { week: string; completed: number }[];
  totalTasks: number;
  doneTasks: number;
}

export interface ApiTeamWorkload {
  id: string;
  name: string;
  initials: string;
  capacity: number;
  assignedHours: number;
  load: number;
  tasks: ApiTask[];
}

export interface ApiRoadmap {
  id: string;
  name: string;
  color: string;
  startDate: string;
  targetDate: string;
  progress: number;
  tasks: ApiTask[];
  doneCount: number;
  description?: string;
  memberAvatars?: string[];
  taskCount?: number;
}

/**
 * The UI models "no priority" as `no_priority`, but `flow_tasks.priority` is
 * `enum('none','low','medium','high','urgent')` and the contract types the
 * field as a plain string, so `no_priority` reaches SQL unvalidated and the
 * write fails with a 500 rather than a 400. Map it at the boundary on the way
 * out; `apiTaskToTask` already normalises `none` back to `no_priority` on the
 * way in.
 */
export type WirePriority = 'none' | 'low' | 'medium' | 'high' | 'urgent';

export const toWirePriority = (priority: ApiTask['priority']): WirePriority =>
  priority === 'no_priority' ? 'none' : priority;

/**
 * The value `flow_projects.icon` actually holds.
 *
 * Every reader in Flow treats this column as a glyph and falls back to this
 * same character, and the backend seeds personal projects with a glyph too. In
 * the dev database 2322 of ~2354 rows hold this exact value and 18 hold '\u2705';
 * the only rows holding anything else are the ten Flow itself corrupted by
 * writing lucide component names here ('folder', 'rocket', 'globe', 'flask',
 * 'building-2'), which then render as those words.
 *
 * `icon` is `@required` on both the create and the update body, so a writer
 * cannot decline to send one, which is why this is a constant rather than an
 * omission. #1085 proposes moving the whole convention to lucide names; doing
 * that needs those 2340 rows migrated, so it is not settled here.
 */
export const DEFAULT_PROJECT_ICON = '\u{1F4CB}';

const toTaskCreatePayload = (data: {
  projectId?: string;
  sectionId?: string;
  title: string;
  description?: string;
  status?: ApiTask['status'];
  priority?: ApiTask['priority'];
  assigneeId?: string;
  dueDate?: string;
  effortMinutes?: number;
  estimatedHours?: number;
  labels?: string[];
}) => ({
  // OneKit-generated contracts wire-encode int64 fields as JSON strings, even
  // the personal-workspace/project sentinel value 0 - never send a bare number.
  project_id: String(data.projectId ? Number(data.projectId) : 0),
  section_id: data.sectionId,
  title: data.title,
  description: data.description ?? '',
  priority: toWirePriority(data.priority ?? 'medium'),
  assignee_id: data.assigneeId ? String(Number(data.assigneeId)) : undefined,
  due_date: data.dueDate ?? '',
  start_date: '',
  status: data.status,
  effort_minutes: data.effortMinutes,
  estimated_hours:
    data.estimatedHours ?? (data.effortMinutes !== undefined ? data.effortMinutes / 60 : undefined),
  labels: data.labels?.map((label) => String(Number(label))),
});

const toMilestoneUpdatePayload = (data: Partial<ApiMilestone>) => ({
  name: data.name,
  description: data.description,
  status: data.status,
  targetDate: data.targetDate,
});

const toCycleUpdatePayload = (data: Partial<ApiCycle>) => ({
  name: data.name,
  status: data.status,
  endDate: data.endDate,
});

const toViewPayload = (data: Partial<ApiView>) => ({
  name: data.name,
  icon: data.icon,
  // The flow/view/v1 contract wire-encodes `filters` as a JSON string (the
  // .onk model documents it as "JSON-encoded filter criteria"; the generated
  // Go struct field is `Filters string`) - sending the raw object fails
  // json.Unmarshal server-side and 400s on every request.
  filters: data.filters !== undefined ? JSON.stringify(data.filters) : undefined,
  groupBy: data.groupBy,
  sortBy: data.sortBy,
  sortDir: data.sortDir,
});

// Reverses toViewPayload's filters encoding. The backend always returns
// `filters` as the same JSON string it stores; parse it back into an object
// so the rest of the app can keep treating SavedView.filters as structured
// data. Tolerates an already-decoded object (defensive) and falls back to {}
// on missing/invalid data rather than throwing.
const decodeApiView = (view: ApiView): ApiView => {
  const raw = view.filters as unknown;
  let filters: Record<string, unknown> = {};
  if (raw && typeof raw === 'object') {
    filters = raw as Record<string, unknown>;
  } else if (typeof raw === 'string' && raw.length > 0) {
    try {
      filters = JSON.parse(raw) as Record<string, unknown>;
    } catch {
      filters = {};
    }
  }
  return { ...view, filters };
};

const toUserSettingsPayload = (data: ApiUserSettings) => {
  const source = data as ApiUserSettings & {
    task_assigned?: boolean;
    task_comments?: boolean;
    status_changes?: boolean;
    weekly_digest?: boolean;
    email_notifications?: boolean;
    sidebar_preferences?: string;
  };
  return {
    taskAssigned: data.taskAssigned ?? source.task_assigned,
    taskComments: data.taskComments ?? source.task_comments,
    mentions: data.mentions,
    statusChanges: data.statusChanges ?? source.status_changes,
    weeklyDigest: data.weeklyDigest ?? source.weekly_digest,
    emailNotifications: data.emailNotifications ?? source.email_notifications,
    sidebarPreferences: data.sidebarPreferences ?? source.sidebar_preferences,
  };
};

/**
 * A member's role on a single project, backed by the `flow_project_members.role`
 * enum. Distinct from the business-wide team role (`Member['role']` in flowStore),
 * which is company scoped and says nothing about a given project.
 *
 * The enum is admin|member|viewer. The API accepts the legacy "manager" alias
 * at its boundary but normalizes it to admin before persistence.
 */
export type FlowProjectRole = 'admin' | 'member' | 'viewer';

export interface ApiProjectMember {
  id: string;
  projectId: string;
  userId: string;
  name: string;
  email: string;
  role: FlowProjectRole;
  addedBy?: string;
  addedAt: string;
}

export interface ApiTaskActivity {
  id: number;
  taskId: number;
  userId: number;
  action: string;
  oldValue: string;
  newValue: string;
  createdAt: string;
}

export interface ApiTaskComment {
  id: string;
  taskId: string;
  userId: string;
  content: string;
  isEdited: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface ApiSubtask {
  id: string;
  title: string;
  status: string;
  assigneeId?: string;
}

export interface ApiAttachment {
  id: string;
  taskId: string;
  name: string;
  url: string;
  size: number;
  uploadedBy: string;
  createdAt: string;
}

export interface ApiTimeLog {
  id: number;
  taskId: number;
  userId: number;
  durationMinutes: number;
  description: string;
  loggedAt: string;
  createdAt: string;
}

export interface ApiTaskDetail {
  task: ApiTask;
  subtasks: ApiSubtask[];
  comments: ApiTaskComment[];
  activity: ApiTaskActivity[];
}

// -- API ------------------------------------------------------------------

function flowResult<T>(operation: () => Promise<unknown>) {
  return oneKitResult(operation).map((response) => response as T);
}

export const flowApi = {
  getMyBusinesses: () =>
    flowResult<{ businesses: ApiBusinessSummary[] }>(() => businessClient.listMyBusinesses({})),

  // -- Projects -----------------------------------------------------------

  // The sidebar's project list is business-scoped now. It has to be v2: a
  // project created through v2 is stored with workspace_id 0, and v1's
  // listProjects filters on `workspace_id = ?`, so a v1 list would never show a
  // newly created project. limit is the page cap the sidebar renders.
  getProjects: () =>
    flowResult<{ projects: ApiProject[] }>(() =>
      projectV2.listProjects({ status: '', visibility: '', q: '', page: 1, limit: 100 })
    ),

  // -- Project Columns ----------------------------------------------------

  getProjectColumns: (projectId: string, params: { page?: number; limit?: number } = {}) =>
    flowResult<{ columns: ApiProjectColumn[]; pagination?: ApiPagination }>(() =>
      projectV2.listProjectColumns({ projectId, ...params })
    ),

  // Columns move to v2 because the v2 requests carry no workspace_id at all:
  // the server derives it from the project rather than trusting a client echo,
  // which is what let the board keep a workspace id in its hand.
  createProjectColumn: (data: {
    project_id: string;
    name: string;
    color?: string;
    sort_order?: number;
  }) =>
    flowResult<{ column: ApiProjectColumn }>(() =>
      projectV2.createProjectColumn({
        projectId: data.project_id,
        name: data.name,
        color: data.color,
        sortOrder: data.sort_order,
      })
    ),

  updateProjectColumn: (data: { id: string; name: string; color?: string; sort_order?: number }) =>
    flowResult<{ message: string }>(() =>
      projectV2.updateProjectColumn({
        id: data.id,
        name: data.name,
        color: data.color,
        sortOrder: data.sort_order,
      })
    ),

  deleteProjectColumn: (id: string) =>
    flowResult<{ message: string }>(() => projectV2.deleteProjectColumn({ id })),

  /**
   * `orderedTaskIds` is the destination column's cards in their new order, the
   * moved one included. Sent, the server stores that order with the move;
   * omitted, only the column changes.
   */
  moveTaskToColumn: (taskId: string, columnId: string, orderedTaskIds?: readonly string[]) =>
    flowResult<{ message: string }>(() =>
      projectV2.moveTaskToColumn({
        taskId,
        columnId,
        ...(orderedTaskIds?.length ? { orderedTaskIds: [...orderedTaskIds] } : {}),
      })
    ),

  createProject: (data: {
    name: string;
    color?: string;
    description?: string;
    startDate?: string;
    dueDate?: string;
    visibility?: string;
  }) =>
    flowResult<{ project: ApiProject }>(() =>
      projectV2.createProject({
        name: data.name,
        // CreateProjectRequest still declares description as @required, and the
        // generated validator rejects "" for a required string, so an omitted
        // description has to travel as the '-' sentinel the rest of this file
        // uses. Sending '' 400s with "description is required".
        description: data.description ?? '-',
        icon: DEFAULT_PROJECT_ICON,
        // A literal, deliberately, and it must stay one: this value is
        // PERSISTED, so a CSS var() would be stored as the string
        // "var(--flow-...)" and read back as a colour no browser can paint.
        // #6366f1 is the last entry of PROJECT_COLORS, the palette the picker
        // offers, so the default is a colour the user could have chosen.
        color: data.color ?? '#6366f1',
        // visibility is optional in the contract and empty already means
        // private, but sending it explicitly keeps the default visible at the
        // call site rather than implied by omission.
        visibility: data.visibility ?? 'private',
        // An omitted date is an unset date. Never manufacture the Unix epoch,
        // which becomes a real project date once the backend persists it.
        startDate: data.startDate ?? '',
        dueDate: data.dueDate ?? '',
      })
    ),

  updateProject: (id: string, data: ProjectUpdate) =>
    // Project settings sends the current project snapshot together with the
    // edited fields because the backend PUT is a replacement update. Keeping
    // omitted values omitted here prevents a partial caller from inventing
    // placeholders or sentinel dates.
    flowResult<{ message: string }>(() =>
      projectV2.updateProject({
        id,
        body: {
          name: data.name ?? '-',
          description: data.description ?? '-',
          icon: data.icon ?? DEFAULT_PROJECT_ICON,
          // A literal, deliberately, and it must stay one: this value is
          // PERSISTED, so a CSS var() would be stored as the string
          // "var(--flow-...)" and read back as a colour no browser can paint.
          // #6366f1 is the last entry of PROJECT_COLORS, the palette the picker
          // offers, so the default is a colour the user could have chosen.
          color: data.color ?? '#6366f1',
          status: data.status ?? 'active',
          startDate: data.startDate ?? '',
          dueDate: data.dueDate ?? '',
        },
      })
    ),

  // Archives rather than deletes: v2 has no deleteProject. archiveProject moves
  // the project to the recycle bin with a purge_at stamp, so it stays restorable
  // until retention expires. The UI copy says so - it used to promise the
  // opposite ("cannot be restored").
  // The contract returns the archived project carrying purge_at, and says
  // clients should surface it wherever they show the archive confirmation.
  // Typing this as { message } discarded that, which is why the confirmation
  // could only speak vaguely about retention.
  archiveProject: (id: string) =>
    flowResult<{ project: ApiProject }>(() => projectV2.archiveProject({ id })),

  getProjectMembers: (projectId: string) =>
    flowResult<{ members: ApiProjectMember[] }>(() => projectV2.listProjectMembers({ projectId })),

  addProjectMember: (data: { project_id: number; user_id: number; role?: string }) =>
    flowResult(() =>
      projectV2.addProjectMember({
        projectId: String(data.project_id),
        userId: String(data.user_id),
        role: data.role ?? 'member',
      })
    ),

  inviteProjectMember: (data: { project_id: number; email: string; role: string }) =>
    flowResult(() =>
      projectV2.inviteProjectMember({
        projectId: String(data.project_id),
        email: data.email,
        role: data.role,
      })
    ),

  removeProjectMember: (data: { project_id: number; user_id: number }) =>
    flowResult(() =>
      projectV2.removeProjectMember({
        projectId: String(data.project_id),
        userId: String(data.user_id),
      })
    ),

  updateProjectMemberRole: (data: { project_id: number; user_id: number; role: string }) =>
    flowResult(() =>
      projectV2.updateProjectMemberRole({
        projectId: String(data.project_id),
        userId: String(data.user_id),
        role: data.role,
      })
    ),

  transferProjectOwnership: (data: { project_id: number; new_owner_id: number }) =>
    flowResult(() =>
      projectV2.transferProjectOwnership({
        projectId: String(data.project_id),
        newOwnerId: String(data.new_owner_id),
      })
    ),

  // -- Tasks --------------------------------------------------------------

  // v2 lists business-wide with project_id, status and assignee_id as optional
  // filters. activeOnly and compact were only ever collected into a local
  // object and dropped before the request; neither exists on either contract,
  // so they are gone rather than silently ignored.
  getTasks: (params: {
    projectId?: string;
    status?: ApiTask['status'];
    assigneeId?: string;
    page?: number;
    limit?: number;
  }) =>
    flowResult<{ tasks: ApiTask[]; pagination?: ApiPagination }>(() =>
      taskV2.listTasks({
        projectId: params.projectId,
        status: params.status ?? '',
        assigneeId: params.assigneeId,
        page: params.page ?? 1,
        limit: params.limit ?? 500,
      })
    ),

  getMyTasks: (params: { page?: number; limit?: number } = {}) =>
    flowResult<{ tasks: ApiTask[]; pagination?: ApiPagination }>(() => taskV2.getMyTasks(params)),

  // Get-or-create the caller's personal "My Tasks" project board and return its
  // ids. Backs the My Tasks landing — we redirect to its project board.
  getMyTasksProject: () => flowResult<{ projectId: string }>(() => taskV2.getMyTasksProject({})),

  searchTasks: (q: string) =>
    flowResult<{ tasks: ApiTask[] }>(() =>
      taskV2.searchTasks({ q, page: 1, limit: 50, sortBy: '', sortDir: '' })
    ),

  getTaskDetail: (id: string) => flowResult<ApiTaskDetail>(() => taskV2.getTaskDetail({ id })),

  // Backend: project_id is optional. When omitted, the handler resolves the
  // caller's auto-provisioned "Inbox" project in the target workspace
  // (task_handlers.go CreateTaskHandler.Serve). This powers quick-create from
  // My Tasks / the command bar where no project is selected.
  // v2 create: no workspace_id, and title is the only required field, which is
  // what lets quick-add post a task with nothing else filled in.
  createTask: (data: {
    projectId?: string;
    sectionId?: string;
    title: string;
    description?: string;
    status?: ApiTask['status'];
    priority?: ApiTask['priority'];
    assigneeId?: string;
    dueDate?: string;
    effortMinutes?: number;
    estimatedHours?: number;
    labels?: string[];
  }) =>
    flowResult<{ task: ApiTask }>(() => {
      const payload = toTaskCreatePayload(data);
      return taskV2.createTask({
        projectId: payload.project_id,
        sectionId: payload.section_id,
        title: payload.title,
        description: payload.description,
        status: payload.status,
        effortMinutes: payload.effort_minutes,
        priority: payload.priority,
        assigneeId: payload.assignee_id,
        dueDate: payload.due_date,
        startDate: payload.start_date,
        estimatedHours: payload.estimated_hours,
        labels: payload.labels,
      });
    }),

  // Partial update: a key the caller did not supply must stay omitted. Undefined
  // keys are dropped before the request goes out, and onion leaves a column
  // untouched when its param arrives NULL (COALESCE(?, col)). Substituting a
  // placeholder is what previously overwrote every field the caller never
  // touched, rewriting a task's title to "-" on a priority-only edit, so do not
  // reintroduce `?? '-'` style defaults here.
  //
  // `archived` is load bearing in both directions: `false` is what un-archives a
  // task so it must be sent as `false`, while `undefined` must stay omitted.
  // Never collapse it to a truthiness check - flowApi.test pins both directions.
  updateTask: (id: string, data: Partial<ApiTask>) =>
    flowResult<{ message: string }>(() =>
      taskV2.updateTask({
        id,
        body: {
          sectionId: data.sectionId !== undefined ? data.sectionId || '0' : undefined,
          title: data.title,
          description: data.description,
          status: data.status,
          priority: data.priority !== undefined ? toWirePriority(data.priority) : undefined,
          assigneeId:
            data.assigneeId !== undefined
              ? data.assigneeId
                ? String(Number(data.assigneeId))
                : '0'
              : undefined,
          dueDate: data.dueDate,
          startDate: data.startDate,
          labels: data.labels?.map((l) => String(Number(l))),
          archived: data.archived,
          effortMinutes: data.effortMinutes,
          estimatedHours:
            data.estimatedHours ?? (data.estimate !== undefined ? data.estimate : undefined),
          focusPin: data.focusPin,
          waitingOn: data.waitingOn,
          decisionNote: data.decisionNote,
        },
      })
    ),

  deleteTask: (id: string) => flowResult<{ message: string }>(() => taskV2.deleteTask({ id })),

  addTaskDependency: (taskId: string, dependsOnTaskId: string) =>
    flowResult<{ message: string }>(() =>
      taskV2.addTaskDependency({ taskId, body: { dependsOnTaskId } })
    ),

  removeTaskDependency: (taskId: string, dependsOnTaskId: string) =>
    flowResult<{ message: string }>(() => taskV2.removeTaskDependency({ taskId, dependsOnTaskId })),

  updateTaskStatus: (id: string, status: ApiTask['status']) =>
    flowResult<{ message: string }>(() => taskV2.updateTaskStatus({ id, body: { status } })),

  addComment: (data: { taskId: string; content: string }) =>
    flowResult<{ comment: ApiTaskComment }>(() =>
      taskV2.addComment({ taskId: data.taskId, content: data.content })
    ),

  deleteComment: (id: string) =>
    flowResult<{ message: string }>(() => taskV2.deleteComment({ id })),

  // -- Task Activity ------------------------------------------------------

  getTaskActivity: (taskId: string) =>
    flowResult<{ activity: ApiTaskActivity[] }>(() => taskV2.getTaskActivity({ taskId })),

  getSubtasks: (id: string) =>
    flowResult<{ subtasks: ApiSubtask[] }>(() => taskV2.getSubtasks({ taskId: id })),

  // Backend returns { task_id, message } - not the created subtask itself.
  // Caller must refetch subtasks to see it appear.
  // A subtask inherits its parent's placement, so v2 asks only for the parent
  // and a title. v1 required project_id AND workspace_id, both @gt(0), and this
  // call site sent "0" for each: creating a subtask 400d on every attempt.
  createSubtask: (taskId: string, data: { title: string; assigneeId?: string }) =>
    flowResult<{ task_id: number; message: string }>(() =>
      taskV2.createSubtask({
        parentTaskId: String(Number(taskId)),
        title: data.title,
        assigneeId: data.assigneeId ? String(Number(data.assigneeId)) : undefined,
      })
    ),

  getAttachments: (id: string) =>
    flowResult<{ attachments: ApiAttachment[] }>(() => taskV2.getAttachments({ taskId: id })),

  addAttachment: (
    taskId: string,
    data: { name: string; size: number; mime?: string; url?: string }
  ) =>
    flowResult<{ attachment_id: number; message: string }>(() =>
      taskV2.createAttachment({
        taskId,
        name: data.name,
        size: String(data.size),
        url: data.url ?? `attachment://${taskId}/${encodeURIComponent(data.name)}`,
      })
    ),

  removeAttachment: (taskId: string, attachmentId: string) =>
    flowResult<{ message: string }>(() => taskV2.deleteAttachment({ id: attachmentId })),

  // -- Time Logs ----------------------------------------------------------

  // durationMinutes is an int64, and OneKit puts an int64 on the wire as a
  // STRING. The interface above declares `number`, so TypeScript believed the
  // lie and `logs.reduce((sum, l) => sum + l.durationMinutes, 0)` concatenated
  // instead of adding: four entries of 2:00, 1:00, 2:15 and 1:30 rendered as
  // "20100226:30 of 40:00" and "50250566% of estimate", with the bar full and
  // an "Over estimate" badge (#1079).
  //
  // Coerced here rather than at the one call site that summed it, so the type
  // becomes true for every consumer instead of for one of them.
  getTimeLogs: (taskId: string) =>
    flowResult<{ timeLogs: ApiTimeLog[] }>(() => taskV2.getTimeLogs({ taskId })).map(
      (response) => ({
        timeLogs: (response.timeLogs ?? []).map((log) => ({
          ...log,
          durationMinutes: Number(log.durationMinutes),
        })),
      })
    ),

  logTime: (data: { task_id: number; duration_minutes: number; description?: string }) =>
    flowResult<{ time_log_id: number; message: string }>(() =>
      taskV2.createTimeLog({
        taskId: String(data.task_id),
        durationMinutes: String(data.duration_minutes),
        description: data.description ?? '-',
      })
    ),

  deleteTimeLog: (id: number) =>
    flowResult<{ message: string }>(() => taskV2.deleteTimeLog({ id: String(id) })),

  // -- Labels -------------------------------------------------------------

  // Labels are one business-wide pool now, so there is no workspace to scope
  // them by and no second "deprecated" variant to keep around.
  listLabels: () =>
    flowResult<{ labels: ApiLabel[] }>(() =>
      labelV2.listLabels({ page: 1, limit: 200, sortBy: '', sortDir: '' })
    ),

  createLabel: (data: { name: string; color: string }) =>
    flowResult<{ label: ApiLabel }>(() =>
      labelV2.createLabel({ name: data.name, color: data.color })
    ),

  updateLabel: (data: { id: number; name: string; color: string }) =>
    flowResult<{ message: string }>(() =>
      labelV2.updateLabel({ id: String(data.id), name: data.name, color: data.color })
    ),

  deleteLabel: (id: number) =>
    flowResult<{ message: string }>(() => labelV2.deleteLabel({ id: String(id) })),

  // task_id/label_ids stay numbers up through the hook and UI layer (TaskDetail
  // compares them against numeric ApiLabel.id values), so the string-on-the-wire
  // conversion happens only here, at the boundary - same int64-as-string
  // convention as createLabel above.
  setTaskLabels: (data: { task_id: number; label_ids: number[] }) =>
    flowResult(() =>
      labelV2.setTaskLabels({
        taskId: String(data.task_id),
        labelIds: data.label_ids.map((id) => String(id)),
      })
    ),

  // -- Statuses (F-A) -------------------------------------------------------

  // Empty means "use the six built-in string statuses", never "not configured
  // yet" - the panel and board both read it that way, so nothing here needs a
  // separate has-custom-statuses flag.
  listStatuses: (projectId: string) =>
    flowResult<{ statuses: ApiStatus[] }>(() => statusClient.listStatuses({ projectId })),

  createStatus: (data: {
    projectId: string;
    name: string;
    bucket: string;
    legacyStatus?: string;
    color?: string;
    sortOrder?: number;
    isDefault?: boolean;
  }) =>
    flowResult<{ status: ApiStatus }>(() =>
      statusClient.createStatus({
        projectId: data.projectId,
        name: data.name,
        bucket: data.bucket,
        legacyStatus: data.legacyStatus,
        color: data.color,
        sortOrder: data.sortOrder,
        isDefault: data.isDefault,
      })
    ),

  updateStatus: (data: {
    id: string;
    name: string;
    bucket: string;
    color: string;
    legacyStatus?: string;
    sortOrder?: number;
    isDefault?: boolean;
  }) =>
    flowResult<{ status: ApiStatus }>(() =>
      statusClient.updateStatus({
        id: data.id,
        body: {
          name: data.name,
          bucket: data.bucket,
          color: data.color,
          legacyStatus: data.legacyStatus,
          sortOrder: data.sortOrder,
          isDefault: data.isDefault,
        },
      })
    ),

  // Never blocks on the status's tasks: they fall back to the built-in string
  // status they already carry rather than being deleted or held hostage.
  deleteStatus: (id: string) =>
    flowResult<{ message: string; reassignedTaskCount: number }>(() =>
      statusClient.deleteStatus({ id })
    ),

  // status_id 0 clears a task's custom status, putting it back on its
  // built-in string status - not a missing argument, so this always sends the
  // field rather than omitting it on falsy.
  setTaskStatus: (data: { taskId: string; statusId: string }) =>
    flowResult<{ message: string; status: string }>(() =>
      statusClient.setTaskStatus({ taskId: data.taskId, statusId: data.statusId })
    ),

  // -- Milestones ---------------------------------------------------------

  // v2 lists business-wide with project_id as an optional narrowing filter, so
  // the old string-or-object overload collapses to one optional project id.
  getMilestones: (projectId?: string) =>
    flowResult<{ milestones: ApiMilestone[] }>(() =>
      milestoneV2.listMilestones({
        projectId,
        page: 1,
        limit: 200,
        sortBy: '',
        sortDir: '',
      })
    ),

  // projectId is required rather than optional: v2 declares project_id
  // @required @gt(0), so every milestone belongs to a project and the caller
  // has to say which. The old signature let it be omitted and 400d.
  createMilestone: (data: Partial<ApiMilestone> & { projectId: string }) =>
    flowResult<{ milestone: ApiMilestone }>(() =>
      milestoneV2.createMilestone({
        projectId: data.projectId,
        name: data.name ?? '',
        description: data.description,
        targetDate: data.targetDate,
      })
    ),

  updateMilestone: (id: string, data: Partial<ApiMilestone>) =>
    flowResult<{ message: string }>(() =>
      milestoneV2.updateMilestone({ id, ...toMilestoneUpdatePayload(data) })
    ),

  // -- Cycles -------------------------------------------------------------

  // Cycles are project-scoped on v2: listCycles declares
  // project_id @required @gt(0), so unlike the other domains there is no
  // business-wide listing to fall back on and the caller must name a project.
  getCycles: (projectId: string) =>
    flowResult<{ cycles: ApiCycle[] }>(() =>
      cycleV2.listCycles({ projectId, page: 1, limit: 200, sortBy: '', sortDir: '' })
    ),

  createCycle: (data: Partial<ApiCycle> & { projectId: string }) =>
    flowResult<{ cycle: ApiCycle }>(() =>
      cycleV2.createCycle({
        projectId: data.projectId,
        name: data.name ?? '',
        startDate: data.startDate,
        endDate: data.endDate,
      })
    ),

  updateCycle: (id: string, data: Partial<ApiCycle>) =>
    flowResult<{ message: string }>(() =>
      cycleV2.updateCycle({ id, ...toCycleUpdatePayload(data) })
    ),

  // -- Views --------------------------------------------------------------

  // Saved views stay business-shared per the ledger, so there is no workspace
  // to scope the list by.
  getViews: () =>
    flowResult<{ views: ApiView[] }>(async () => {
      const res = await viewV2.listViews({ page: 1, limit: 200 });
      return { views: ((res as { views?: ApiView[] }).views ?? []).map(decodeApiView) };
    }),

  createView: (data: Partial<ApiView>) =>
    flowResult<{ view: ApiView }>(async () => {
      const res = (await viewV2.createView(
        toViewPayload(data) as Parameters<typeof viewV2.createView>[0]
      )) as { view?: ApiView };
      return { view: res.view ? decodeApiView(res.view) : res.view };
    }),

  updateView: (id: string, data: Partial<ApiView>) =>
    flowResult<{ message: string }>(() =>
      // Spelled out rather than spread behind a cast: the v2 request carries
      // `id` itself, so spreading a casted payload over an explicit `id`
      // silently overwrote it.
      viewV2.updateView({
        id,
        name: data.name,
        icon: data.icon,
        filters: data.filters !== undefined ? JSON.stringify(data.filters) : undefined,
        groupBy: data.groupBy,
        sortBy: data.sortBy,
        sortDir: data.sortDir,
        visibility: data.visibility,
      })
    ),

  deleteView: (id: string) => flowResult(() => viewV2.deleteView({ id })),

  // -- Insights -----------------------------------------------------------

  // Insights, team workload and the roadmap are all business-wide on v2 and
  // take no parameters at all.
  getInsights: () => flowResult<ApiInsights>(() => insightsV2.getInsights({})),

  // -- Team workload ------------------------------------------------------

  getTeamWorkload: () =>
    flowResult<{ members: ApiTeamWorkload[] }>(() => insightsV2.getTeamWorkload({})),

  // -- Roadmap ------------------------------------------------------------

  getRoadmap: () => flowResult<{ projects: ApiRoadmap[] }>(() => insightsV2.getRoadmap({})),

  // -- Notifications ------------------------------------------------------

  getNotifications: (params: { page?: number; limit?: number } = {}) =>
    flowResult<{ notifications: ApiNotification[]; pagination?: ApiPagination }>(() =>
      notificationClient.listNotifications(params)
    ),

  markNotificationRead: (id: string) =>
    flowResult(() => notificationClient.markNotificationRead({ id })),

  markAllNotificationsRead: () => flowResult(() => notificationClient.markAllNotificationsRead({})),

  // -- User Settings ------------------------------------------------------

  getUserSettings: () => flowResult<ApiUserSettings>(() => settingClient.getUserSettings({})),

  updateUserSettings: (data: ApiUserSettings) =>
    flowResult(() => settingClient.updateUserSettings(toUserSettingsPayload(data))),

  // -- Account / Profile --------------------------------------------------

  updateProfile: (data: { name?: string; email?: string }) =>
    flowResult(() =>
      userClient.updateProfile({
        name: data.name ?? '',
        email: data.email,
      })
    ),

  changePassword: (data: { currentPassword: string; newPassword: string }) =>
    flowResult(() =>
      userClient.changePassword({
        currentPassword: data.currentPassword,
        newPassword: data.newPassword,
      })
    ),
};
