import type { Task as ContractTask } from '@voxire/voxie/flow/task/v1/types';

import { safe } from '../lib/api';

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
}

export interface DashboardStats {
  openTasks: number;
  inProgress: number;
  completedThisWeek: number;
  overdue: number;
  myTasks: ApiTask[];
  recentProjects: ApiProject[];
}

export interface ApiWorkspace {
  id: string;
  name: string;
  color: string;
  memberCount?: number;
}

export interface ApiBusinessSummary {
  id: string | number;
  name: string;
  isActive?: boolean;
}

export interface ApiProject {
  id: string;
  workspaceId: string;
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

export interface ApiTask {
  id: string;
  projectId: string;
  workspaceId: string;
  title: string;
  description?: string;
  status: 'todo' | 'in_progress' | 'in_review' | 'done' | 'cancelled' | 'backlog';
  priority: 'urgent' | 'high' | 'medium' | 'low' | 'no_priority';
  assignee: string;
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
  workspaceId: number;
  name: string;
  color: string;
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
  workspaceId: string;
  name: string;
  startDate: string;
  endDate: string;
  status: 'upcoming' | 'active' | 'completed';
  scope: number;
  completed: number;
}

export interface ApiProjectColumn {
  id: string;
  projectId: string;
  workspaceId: string;
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
  workspaceId?: string;
  description?: string;
  memberAvatars?: string[];
  taskCount?: number;
}

const toTaskCreatePayload = (data: {
  projectId?: string;
  workspaceId: string;
  title: string;
  description?: string;
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
  workspace_id: String(Number(data.workspaceId) || 0),
  title: data.title,
  description: data.description ?? '-',
  priority: data.priority ?? 'medium',
  assignee_id: data.assigneeId ? String(Number(data.assigneeId)) : undefined,
  due_date: data.dueDate ?? '-',
  start_date: '-',
  estimated_hours:
    data.estimatedHours ?? (data.effortMinutes !== undefined ? data.effortMinutes / 60 : undefined),
  labels: data.labels?.map((label) => String(Number(label))),
});

const toMilestoneCreatePayload = (data: Partial<ApiMilestone>) => ({
  projectId: data.projectId,
  workspaceId:
    'workspaceId' in data
      ? (data as Partial<ApiMilestone> & { workspaceId?: string }).workspaceId
      : undefined,
  name: data.name,
  description: data.description,
  targetDate: data.targetDate,
});

const toMilestoneUpdatePayload = (data: Partial<ApiMilestone>) => ({
  name: data.name,
  description: data.description,
  status: data.status,
  targetDate: data.targetDate,
});

const toCycleCreatePayload = (data: Partial<ApiCycle>) => ({
  name: data.name,
  workspaceId: data.workspaceId,
  startDate: data.startDate,
  endDate: data.endDate,
});

const toCycleUpdatePayload = (data: Partial<ApiCycle>) => ({
  name: data.name,
  status: data.status,
  endDate: data.endDate,
});

const toViewPayload = (data: Partial<ApiView> & { workspaceId?: string }) => ({
  workspaceId: data.workspaceId,
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

const toViewUpdatePayload = (data: Partial<ApiView>) => ({
  name: data.name,
  icon: data.icon,
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
  };
  return {
    taskAssigned: data.taskAssigned ?? source.task_assigned,
    taskComments: data.taskComments ?? source.task_comments,
    mentions: data.mentions,
    statusChanges: data.statusChanges ?? source.status_changes,
    weeklyDigest: data.weeklyDigest ?? source.weekly_digest,
    emailNotifications: data.emailNotifications ?? source.email_notifications,
  };
};

export interface ApiWorkspaceMember {
  id: string;
  workspaceId: string;
  userId: string;
  name: string;
  email: string;
  role: 'admin' | 'member' | 'viewer';
  invitedBy?: string;
  joinedAt: string;
}

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

export const flowApi = {
  getDashboard: (workspaceId?: string) =>
    safe.get<DashboardStats>('/flow/dashboard/v1/dashboard', {
      params: workspaceId ? { workspace_id: workspaceId } : undefined,
    }),
  // -- Workspaces ---------------------------------------------------------

  getWorkspaces: () => safe.get<{ workspaces: ApiWorkspace[] }>('/flow/workspace/v1/workspaces'),

  getMyBusinesses: () =>
    safe.get<{ businesses: ApiBusinessSummary[] }>('/hub/business/v1/businesses/mine'),

  createWorkspace: (data: { name: string; color: string; description?: string; icon?: string }) =>
    safe.post<{ workspace: ApiWorkspace }>('/flow/workspace/v1/workspaces', {
      name: data.name,
      description: data.description ?? '-',
      icon: data.icon ?? 'folder',
      color: data.color,
    }),

  updateWorkspace: (
    id: string,
    data: Partial<{ name: string; color: string; description: string; icon: string }>
  ) =>
    safe.put<{ message: string }>(`/flow/workspace/v1/workspaces/${id}`, {
      body: {
        name: data.name ?? '-',
        description: data.description ?? '-',
        icon: data.icon ?? 'folder',
        color: data.color ?? '#6366f1',
      },
    }),

  deleteWorkspace: (id: string) =>
    safe.delete<{ message: string }>(`/flow/workspace/v1/workspaces/${id}`),

  getWorkspaceMembers: (workspaceId: string) =>
    safe.get<{ members: ApiWorkspaceMember[] }>('/flow/workspace/v1/workspaces/members', {
      params: { workspace_id: workspaceId },
    }),

  addWorkspaceMember: (data: { workspace_id: number; user_id: number; role?: string }) =>
    safe.post('/flow/workspace/v1/workspaces/members', {
      workspace_id: data.workspace_id,
      user_id: data.user_id,
      role: data.role ?? 'member',
    }),

  removeWorkspaceMember: (data: { workspace_id: number; user_id: number }) =>
    safe.delete('/flow/workspace/v1/workspaces/members', {
      params: { workspace_id: data.workspace_id, user_id: data.user_id },
    }),

  updateWorkspaceMemberRole: (data: { workspace_id: number; user_id: number; role: string }) =>
    safe.put('/flow/workspace/v1/workspaces/members/role', data),

  // -- Projects -----------------------------------------------------------

  getProjects: (workspaceId: string) =>
    safe.get<{ projects: ApiProject[] }>('/flow/project/v1/projects', {
      params: { workspace_id: workspaceId },
    }),

  // -- Project Columns ----------------------------------------------------

  getProjectColumns: (projectId: string) =>
    safe.get<{ columns: ApiProjectColumn[] }>('/flow/project/v1/projects/columns', {
      params: { project_id: projectId },
    }),

  createProjectColumn: (data: {
    project_id: string;
    workspace_id: string;
    name: string;
    color?: string;
    sort_order?: number;
  }) => safe.post<{ column: ApiProjectColumn }>('/flow/project/v1/projects/columns', data),

  updateProjectColumn: (data: { id: string; name: string; color?: string; sort_order?: number }) =>
    safe.put<{ message: string }>(`/flow/project/v1/projects/columns/${data.id}`, {
      name: data.name,
      color: data.color,
      sort_order: data.sort_order,
    }),

  deleteProjectColumn: (id: string) =>
    safe.delete<{ message: string }>(`/flow/project/v1/projects/columns/${id}`),

  moveTaskToColumn: (taskId: string, columnId: string) =>
    safe.put<{ message: string }>('/flow/project/v1/projects/columns/move', {
      task_id: taskId,
      column_id: columnId,
    }),

  createProject: (data: {
    workspaceId: string;
    name: string;
    color?: string;
    description?: string;
    startDate?: string;
    dueDate?: string;
  }) =>
    safe.post<{ project: ApiProject }>('/flow/project/v1/projects', {
      // int64 fields are wire-encoded as JSON strings, matching every other
      // numeric id in this API - sending a bare number 400s with "cannot
      // unmarshal number into Go struct field of type string".
      //
      // The `|| 0` fallback is NOT a "my personal workspace" sentinel, despite
      // what this comment used to claim: CreateProjectRequest declares
      // `workspace_id: int64 @required @gt(0)`, so "0" is rejected with
      // "workspace_id must be greater than 0". Callers must gate on
      // isApiWorkspaceId() before reaching this.
      workspace_id: String(Number(data.workspaceId) || 0),
      name: data.name,
      description: data.description ?? '-',
      icon: 'folder',
      color: data.color ?? '#6366f1',
      // An omitted date is an unset date. Never manufacture the Unix epoch,
      // which becomes a real project date once the backend persists it.
      start_date: data.startDate,
      due_date: data.dueDate,
    }),

  updateProject: (id: string, data: ProjectUpdate) =>
    // Project settings sends the current project snapshot together with the
    // edited fields because the backend PUT is a replacement update. Keeping
    // omitted values omitted here prevents a partial caller from inventing
    // placeholders or sentinel dates.
    safe.put<{ message: string }>(`/flow/project/v1/projects/${id}`, {
      name: data.name,
      description: data.description,
      icon: data.icon,
      color: data.color,
      status: data.status,
      start_date: data.startDate,
      due_date: data.dueDate,
    }),

  deleteProject: (id: string) =>
    safe.delete<{ message: string }>(`/flow/project/v1/projects/${id}`),

  getProjectMembers: (projectId: string) =>
    safe.get<{ members: ApiProjectMember[] }>('/flow/project/v1/projects/members', {
      params: { project_id: projectId },
    }),

  addProjectMember: (data: { project_id: number; user_id: number; role?: string }) =>
    safe.post('/flow/project/v1/projects/members', data),

  inviteProjectMember: (data: { project_id: number; email: string; role: string }) =>
    safe.post('/flow/project/v1/projects/members/invite', data),

  removeProjectMember: (data: { project_id: number; user_id: number }) =>
    safe.delete('/flow/project/v1/projects/members', {
      params: { project_id: data.project_id, user_id: data.user_id },
    }),

  updateProjectMemberRole: (data: { project_id: number; user_id: number; role: string }) =>
    safe.put('/flow/project/v1/projects/members/role', data),

  transferProjectOwnership: (data: { project_id: number; new_owner_id: number }) =>
    safe.put('/flow/project/v1/projects/transfer', data),

  moveProject: (data: { project_id: number; workspace_id: number }) =>
    safe.put('/flow/project/v1/projects/move', data),

  // -- Tasks --------------------------------------------------------------

  getTasks: (params: {
    projectId?: string;
    workspaceId?: string;
    status?: ApiTask['status'];
    assigneeId?: string;
    activeOnly?: boolean;
    compact?: boolean;
  }) => {
    const q: Record<string, string> = {};
    if (params.projectId) q.project_id = params.projectId;
    if (params.workspaceId) q.workspace_id = params.workspaceId;
    if (params.status) q.status = params.status;
    if (params.assigneeId) q.assignee_id = params.assigneeId;
    if (params.activeOnly) q.active_only = 'true';
    if (params.compact) q.compact = 'true';
    return safe.get<{ tasks: ApiTask[] }>('/flow/task/v1/tasks', { params: q });
  },

  getMyTasks: () => safe.get<{ tasks: ApiTask[] }>('/flow/task/v1/tasks/my'),

  // Get-or-create the caller's personal "My Tasks" project board and return its
  // ids. Backs the My Tasks landing — we redirect to its project board.
  getMyTasksProject: () =>
    safe.get<{ projectId: string; workspaceId: string }>('/flow/task/v1/tasks/my-project'),

  searchTasks: (q: string, workspaceId: string) =>
    safe.get<{ tasks: ApiTask[] }>('/flow/task/v1/tasks/search', {
      params: { q, workspace_id: workspaceId },
    }),

  getTaskDetail: (id: string) => safe.get<ApiTaskDetail>(`/flow/task/v1/tasks/${id}`),

  // Backend: project_id is optional. When omitted, the handler resolves the
  // caller's auto-provisioned "Inbox" project in the target workspace
  // (task_handlers.go CreateTaskHandler.Serve). This powers quick-create from
  // My Tasks / the command bar where no project is selected.
  createTask: (data: {
    projectId?: string;
    workspaceId: string;
    title: string;
    description?: string;
    status?: ApiTask['status'];
    priority?: ApiTask['priority'];
    assigneeId?: string;
    dueDate?: string;
    effortMinutes?: number;
    estimatedHours?: number;
    labels?: string[];
  }) => safe.post<{ task: ApiTask }>('/flow/task/v1/tasks', toTaskCreatePayload(data)),

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
    safe.put<{ message: string }>(`/flow/task/v1/tasks/${id}`, {
      body: {
        title: data.title,
        description: data.description,
        status: data.status,
        priority: data.priority,
        assignee_id:
          data.assigneeId !== undefined
            ? data.assigneeId
              ? Number(data.assigneeId)
              : null
            : undefined,
        due_date: data.dueDate,
        start_date: data.startDate,
        labels: data.labels?.map((l) => Number(l)),
        archived: data.archived,
      },
    }),

  deleteTask: (id: string) => safe.delete<{ message: string }>(`/flow/task/v1/tasks/${id}`),

  updateTaskStatus: (id: string, status: ApiTask['status']) =>
    safe.patch<{ message: string }>(`/flow/task/v1/tasks/${id}/status`, {
      body: { status },
    }),

  addComment: (data: { taskId: string; content: string }) =>
    safe.post<{ comment: ApiTaskComment }>(`/flow/task/v1/tasks/${data.taskId}/comments`, {
      content: data.content,
    }),

  deleteComment: (id: string) =>
    safe.delete<{ message: string }>(`/flow/task/v1/tasks/comments/${id}`),

  // -- Task Activity ------------------------------------------------------

  getTaskActivity: (taskId: string) =>
    safe.get<{ activity: ApiTaskActivity[] }>(`/flow/task/v1/tasks/${taskId}/activity`),

  getSubtasks: (id: string) =>
    safe.get<{ subtasks: ApiSubtask[] }>(`/flow/task/v1/tasks/${id}/subtasks`),

  // Backend returns { task_id, message } - not the created subtask itself.
  // Caller must refetch subtasks to see it appear.
  createSubtask: (
    taskId: string,
    data: { title: string; assigneeId?: string; projectId?: string; workspaceId?: string }
  ) =>
    safe.post<{ task_id: number; message: string }>('/flow/task/v1/tasks/subtasks', {
      parent_task_id: Number(taskId),
      project_id: Number(data.projectId ?? 0),
      workspace_id: Number(data.workspaceId ?? 0),
      title: data.title,
      description: '-',
      priority: 'medium',
      due_date: '-',
      start_date: '-',
      assignee_id: data.assigneeId ? Number(data.assigneeId) : undefined,
    }),

  getAttachments: (id: string) =>
    safe.get<{ attachments: ApiAttachment[] }>(`/flow/task/v1/tasks/${id}/attachments`),

  addAttachment: (
    taskId: string,
    data: { name: string; size: number; mime?: string; url?: string }
  ) =>
    safe.post<{ attachment_id: number; message: string }>('/flow/task/v1/tasks/attachments', {
      task_id: Number(taskId),
      name: data.name,
      size: data.size,
      url: data.url ?? `attachment://${taskId}/${encodeURIComponent(data.name)}`,
    }),

  removeAttachment: (taskId: string, attachmentId: string) =>
    safe.delete<{ message: string }>(`/flow/task/v1/tasks/attachments/${attachmentId}`),

  // -- Time Logs ----------------------------------------------------------

  getTimeLogs: (taskId: string) =>
    safe.get<{ timeLogs: ApiTimeLog[] }>(`/flow/task/v1/tasks/${taskId}/time-logs`),

  logTime: (data: { task_id: number; duration_minutes: number; description?: string }) =>
    safe.post<{ time_log_id: number; message: string }>(
      `/flow/task/v1/tasks/${data.task_id}/time-logs`,
      {
        task_id: data.task_id,
        duration_minutes: data.duration_minutes,
        description: data.description ?? '-',
      }
    ),

  deleteTimeLog: (id: number) =>
    safe.delete<{ message: string }>(`/flow/task/v1/tasks/time-logs/${id}`),

  // -- Labels -------------------------------------------------------------

  listLabels: (workspaceId: string) =>
    safe.get<{ labels: ApiLabel[] }>('/flow/label/v1/labels', {
      params: workspaceId ? { workspace_id: workspaceId } : undefined,
    }),

  // workspace_id is already a string end-to-end from the workspace store (see
  // ApiWorkspace.id / activeWorkspaceId), so this passes it straight through -
  // no Number()/String() round trip. int64 fields are wire-encoded as JSON
  // strings, matching every other numeric id in this API; a bare number 400s
  // with "cannot unmarshal number into Go struct field of type string".
  createLabel: (data: { workspace_id: string; name: string; color: string }) =>
    safe.post<{ label: ApiLabel }>('/flow/label/v1/labels', data),

  updateLabel: (data: { id: number; name: string; color: string }) =>
    safe.put<{ message: string }>(`/flow/label/v1/labels/${data.id}`, {
      name: data.name,
      color: data.color,
    }),

  deleteLabel: (id: number) => safe.delete<{ message: string }>(`/flow/label/v1/labels/${id}`),

  // task_id/label_ids stay numbers up through the hook and UI layer (TaskDetail
  // compares them against numeric ApiLabel.id values), so the string-on-the-wire
  // conversion happens only here, at the boundary - same int64-as-string
  // convention as createLabel above.
  setTaskLabels: (data: { task_id: number; label_ids: number[] }) =>
    safe.put('/flow/label/v1/tasks/labels', {
      task_id: String(data.task_id),
      label_ids: data.label_ids.map((id) => String(id)),
    }),

  /** @deprecated use listLabels(workspaceId) */
  getLabels: () => safe.get<{ labels: ApiLabel[] }>('/flow/label/v1/labels'),

  // -- Milestones ---------------------------------------------------------

  getMilestones: (query?: string | { projectId?: string; workspaceId?: string }) => {
    const params =
      typeof query === 'string'
        ? query
          ? { project_id: query }
          : undefined
        : query
          ? {
              ...(query.projectId ? { project_id: query.projectId } : {}),
              ...(query.workspaceId ? { workspace_id: query.workspaceId } : {}),
            }
          : undefined;

    return safe.get<{ milestones: ApiMilestone[] }>('/flow/milestone/v1/milestones', {
      params,
    });
  },

  createMilestone: (data: Partial<ApiMilestone>) =>
    safe.post<{ milestone: ApiMilestone }>(
      '/flow/milestone/v1/milestones',
      toMilestoneCreatePayload(data)
    ),

  updateMilestone: (id: string, data: Partial<ApiMilestone>) =>
    safe.put<{ message: string }>(
      `/flow/milestone/v1/milestones/${id}`,
      toMilestoneUpdatePayload(data)
    ),

  // -- Cycles -------------------------------------------------------------

  getCycles: (workspaceId?: string) =>
    safe.get<{ cycles: ApiCycle[] }>('/flow/cycle/v1/cycles', {
      params: workspaceId ? { workspace_id: workspaceId } : undefined,
    }),

  createCycle: (data: Partial<ApiCycle>) =>
    safe.post<{ cycle: ApiCycle }>('/flow/cycle/v1/cycles', toCycleCreatePayload(data)),

  updateCycle: (id: string, data: Partial<ApiCycle>) =>
    safe.put<{ message: string }>(`/flow/cycle/v1/cycles/${id}`, toCycleUpdatePayload(data)),

  // -- Views --------------------------------------------------------------

  getViews: (workspaceId?: string) =>
    safe
      .get<{ views: ApiView[] }>('/flow/view/v1/views', {
        params: workspaceId ? { workspace_id: workspaceId } : undefined,
      })
      .map((res) => ({
        ...res,
        views: (res.views ?? []).map(decodeApiView),
      })),

  createView: (data: Partial<ApiView>) =>
    safe
      .post<{ view: ApiView }>('/flow/view/v1/views', toViewPayload(data))
      .map((res) => ({ ...res, view: res.view ? decodeApiView(res.view) : res.view })),

  updateView: (id: string, data: Partial<ApiView>) =>
    safe.put<{ message: string }>(`/flow/view/v1/views/${id}`, toViewUpdatePayload(data)),

  deleteView: (id: string) => safe.delete(`/flow/view/v1/views/${id}`),

  // -- Insights -----------------------------------------------------------

  getInsights: (workspaceId?: string) =>
    safe.get<ApiInsights>('/flow/insights/v1/overview', {
      params: workspaceId ? { workspace_id: workspaceId } : undefined,
    }),

  // -- Team workload ------------------------------------------------------

  getTeamWorkload: (workspaceId?: string) =>
    safe.get<{ members: ApiTeamWorkload[] }>('/flow/insights/v1/team/workload', {
      params: workspaceId ? { workspace_id: workspaceId } : undefined,
    }),

  // -- Roadmap ------------------------------------------------------------

  getRoadmap: (workspaceId?: string) =>
    safe.get<{ projects: ApiRoadmap[] }>('/flow/insights/v1/roadmap', {
      params: workspaceId ? { workspace_id: workspaceId } : undefined,
    }),

  // -- Notifications ------------------------------------------------------

  getNotifications: () =>
    safe.get<{ notifications: ApiNotification[] }>('/flow/notification/v1/notifications'),

  markNotificationRead: (id: string) =>
    safe.put(`/flow/notification/v1/notifications/${id}/read`, {}),

  markAllNotificationsRead: () => safe.put('/flow/notification/v1/notifications/read-all', {}),

  // -- User Settings ------------------------------------------------------

  getUserSettings: () => safe.get<ApiUserSettings>('/flow/setting/v1/settings'),

  updateUserSettings: (data: ApiUserSettings) =>
    safe.put('/flow/setting/v1/settings', toUserSettingsPayload(data)),

  // -- Account / Profile --------------------------------------------------

  updateProfile: (data: { name?: string; email?: string }) =>
    safe.patch('/hub/user/v1/users/me', {
      name: data.name ?? '',
      ...(data.email ? { email: data.email } : {}),
    }),

  changePassword: (data: { currentPassword: string; newPassword: string }) =>
    safe.post('/hub/user/v1/users/password', {
      current_password: data.currentPassword,
      new_password: data.newPassword,
    }),
};
