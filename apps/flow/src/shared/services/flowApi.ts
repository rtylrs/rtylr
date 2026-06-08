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

export interface ApiProject {
  id: string;
  workspaceId: string;
  name: string;
  color: string;
  taskCount: number;
  completedCount: number;
  memberAvatars: string[];
}

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
  subtaskCount: number;
  projectName: string;
  estimate?: number;
  estimatedHours?: number | null;
  actualHours?: number | null;
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

export interface ApiWorkspaceMember {
  id: string;
  workspaceId: string;
  userId: string;
  name: string;
  email: string;
  role: 'manager' | 'member' | 'viewer';
  invitedBy?: string;
  joinedAt: string;
}

export interface ApiProjectMember {
  id: string;
  projectId: string;
  userId: string;
  name: string;
  email: string;
  role: 'manager' | 'member' | 'viewer';
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

// -- Per-feature services (re-exported for backwards compat) ---------------

export { taskService } from './taskService';
export { workspaceService } from './workspaceService';
export { cycleService } from './cycleService';
export { milestoneService } from './milestoneService';
export { viewService } from './viewService';
export { labelService } from './labelService';
export { insightsService } from './insightsService';

// -- Legacy barrel (delegates to per-feature services) ---------------------

export const flowApi = {
  getDashboard: (workspaceId?: string) =>
    safe.get<DashboardStats>('/flow/dashboard', {
      params: workspaceId ? { workspace_id: workspaceId } : undefined,
    }),
  // -- Workspaces ---------------------------------------------------------

  getWorkspaces: () => safe.get<{ workspaces: ApiWorkspace[] }>('/flow/workspaces'),

  createWorkspace: (data: { name: string; color: string }) =>
    safe.post<{ workspace: ApiWorkspace }>('/flow/workspaces', data),

  // Backend: PUT /flow/workspaces with id in body (not path-param).
  updateWorkspace: (id: string, data: Partial<{ name: string; color: string }>) =>
    safe.put<{ message: string }>('/flow/workspaces', { id: Number(id), ...data }),

  // Backend: DELETE /flow/workspaces with id in body (not /{id}/delete).
  deleteWorkspace: (id: string) =>
    safe.delete<{ message: string }>('/flow/workspaces', { data: { id: Number(id) } }),

  getWorkspaceMembers: (workspaceId: string) =>
    safe.get<{ members: ApiWorkspaceMember[] }>('/flow/workspaces/members', {
      params: { workspace_id: workspaceId },
    }),

  addWorkspaceMember: (data: { workspace_id: number; user_id: number; role?: string }) =>
    safe.post('/flow/workspaces/member', data),

  removeWorkspaceMember: (data: { workspace_id: number; user_id: number }) =>
    safe.delete('/flow/workspaces/member', { data }),

  updateWorkspaceMemberRole: (data: { workspace_id: number; user_id: number; role: string }) =>
    safe.put('/flow/workspaces/member/role', data),

  // -- Projects -----------------------------------------------------------

  getProjects: (workspaceId: string) =>
    safe.get<{ projects: ApiProject[] }>('/flow/projects', {
      params: { workspace_id: workspaceId },
    }),

  // -- Project Columns ----------------------------------------------------

  getProjectColumns: (projectId: string) =>
    safe.get<{ columns: ApiProjectColumn[] }>('/flow/projects/columns', {
      params: { project_id: projectId },
    }),

  createProjectColumn: (data: {
    project_id: string;
    workspace_id: string;
    name: string;
    color?: string;
    sort_order?: number;
  }) => safe.post<{ column: ApiProjectColumn }>('/flow/projects/columns', data),

  updateProjectColumn: (data: { id: string; name: string; color?: string; sort_order?: number }) =>
    safe.put<{ message: string }>('/flow/projects/columns', data),

  deleteProjectColumn: (id: string) =>
    safe.delete<{ message: string }>('/flow/projects/columns', { data: { id } }),

  moveTaskToColumn: (taskId: string, columnId: string) =>
    safe.put<{ message: string }>('/flow/projects/columns/move', {
      task_id: taskId,
      column_id: columnId,
    }),

  createProject: (data: { workspaceId: string; name: string; color?: string }) =>
    safe.post<{ project: ApiProject }>('/flow/projects', data),

  // Backend: PUT /flow/projects with id in body.
  updateProject: (id: string, data: Partial<{ name: string; color: string }>) =>
    safe.put<{ message: string }>('/flow/projects', { id: Number(id), ...data }),

  // Backend: DELETE /flow/projects with id in body.
  deleteProject: (id: string) =>
    safe.delete<{ message: string }>('/flow/projects', { data: { id: Number(id) } }),

  getProjectMembers: (projectId: string) =>
    safe.get<{ members: ApiProjectMember[] }>('/flow/projects/members', {
      params: { project_id: projectId },
    }),

  addProjectMember: (data: { project_id: number; user_id: number; role?: string }) =>
    safe.post('/flow/projects/member', data),

  inviteProjectMember: (data: { project_id: number; email: string; role: string }) =>
    safe.post('/flow/projects/member/invite', data),

  removeProjectMember: (data: { project_id: number; user_id: number }) =>
    safe.delete('/flow/projects/member', { data }),

  updateProjectMemberRole: (data: { project_id: number; user_id: number; role: string }) =>
    safe.put('/flow/projects/member/role', data),

  transferProjectOwnership: (data: { project_id: number; new_owner_id: number }) =>
    safe.put('/flow/projects/transfer', data),

  moveProject: (data: { project_id: number; workspace_id: number }) =>
    safe.put('/flow/projects/move', data),

  // -- Tasks --------------------------------------------------------------

  getTasks: (params: {
    projectId?: string;
    workspaceId?: string;
    status?: ApiTask['status'];
    assigneeId?: string;
  }) => {
    const q: Record<string, string> = {};
    if (params.projectId) q.project_id = params.projectId;
    if (params.workspaceId) q.workspace_id = params.workspaceId;
    if (params.status) q.status = params.status;
    if (params.assigneeId) q.assignee_id = params.assigneeId;
    return safe.get<{ tasks: ApiTask[] }>('/flow/tasks', { params: q });
  },

  getMyTasks: () => safe.get<{ tasks: ApiTask[] }>('/flow/tasks/my'),

  // Get-or-create the caller's personal "My Tasks" project board and return its
  // ids. Backs the My Tasks landing — we redirect to its project board.
  getMyTasksProject: () =>
    safe.get<{ projectId: number; workspaceId: number }>('/flow/my-tasks-project'),

  searchTasks: (q: string, workspaceId: string) =>
    safe.get<{ tasks: ApiTask[] }>('/flow/tasks/search', {
      params: { q, workspace_id: workspaceId },
    }),

  getTaskDetail: (id: string) => safe.get<ApiTaskDetail>('/flow/tasks/detail', { params: { id } }),

  // Backend: project_id is optional. When omitted, the handler resolves the
  // caller's auto-provisioned "Inbox" project in the target workspace
  // (task_handlers.go CreateTaskHandler.Serve). This powers quick-create from
  // My Tasks / the command bar where no project is selected.
  createTask: (data: {
    projectId?: string;
    workspaceId: string;
    title: string;
    status?: ApiTask['status'];
    priority?: ApiTask['priority'];
    assigneeId?: string;
    dueDate?: string;
  }) => safe.post<{ task: ApiTask }>('/flow/tasks', data),

  // Backend: PUT /flow/tasks with id in body.
  updateTask: (id: string, data: Partial<ApiTask>) =>
    safe.put<{ message: string }>('/flow/tasks', { id: Number(id), ...data }),

  // Backend: DELETE /flow/tasks with id in body.
  deleteTask: (id: string) =>
    safe.delete<{ message: string }>('/flow/tasks', { data: { id: Number(id) } }),

  // Backend: PUT /flow/tasks/status.
  updateTaskStatus: (id: string, status: ApiTask['status']) =>
    safe.put<{ message: string }>('/flow/tasks/status', { id: Number(id), status }),

  addComment: (data: { taskId: string; content: string }) =>
    safe.post<{ comment: ApiTaskComment }>('/flow/tasks/comment', data),

  // Backend: DELETE /flow/tasks/comment with id in body.
  deleteComment: (id: string) =>
    safe.delete<{ message: string }>('/flow/tasks/comment', { data: { id: Number(id) } }),

  // -- Task Activity ------------------------------------------------------

  getTaskActivity: (taskId: string) =>
    safe.get<{ activity: ApiTaskActivity[] }>('/flow/tasks/activity', {
      params: { task_id: taskId },
    }),

  // -- Subtasks -----------------------------------------------------------

  getSubtasks: (id: string) => safe.get<{ subtasks: ApiSubtask[] }>(`/flow/tasks/${id}/subtasks`),

  // Backend returns { task_id, message } - not the created subtask itself.
  // Caller must refetch subtasks to see it appear.
  createSubtask: (taskId: string, data: { title: string; assigneeId?: string }) =>
    safe.post<{ task_id: number; message: string }>(`/flow/tasks/${taskId}/subtasks`, data),

  // -- Attachments --------------------------------------------------------

  getAttachments: (id: string) =>
    safe.get<{ attachments: ApiAttachment[] }>(`/flow/tasks/${id}/attachments`),

  // -- Time Logs ----------------------------------------------------------

  getTimeLogs: (taskId: string) =>
    safe.get<{ timeLogs: ApiTimeLog[] }>('/flow/tasks/time', { params: { task_id: taskId } }),

  logTime: (data: { task_id: number; duration_minutes: number; description?: string }) =>
    safe.post<ApiTimeLog>('/flow/tasks/time', data),

  deleteTimeLog: (id: number) => safe.delete('/flow/tasks/time', { data: { id } }),

  // -- Labels -------------------------------------------------------------

  listLabels: (workspaceId: string) =>
    safe.get<{ labels: ApiLabel[] }>('/flow/labels', { params: { workspace_id: workspaceId } }),

  createLabel: (data: { workspace_id: number; name: string; color: string }) =>
    safe.post<{ label: ApiLabel }>('/flow/labels', data),

  updateLabel: (data: { id: number; name: string; color: string }) =>
    safe.put('/flow/labels', data),

  deleteLabel: (id: number) => safe.delete('/flow/labels', { params: { id } }),

  setTaskLabels: (data: { task_id: number; label_ids: number[] }) =>
    safe.put('/flow/tasks/labels', data),

  /** @deprecated use listLabels(workspaceId) */
  getLabels: () => safe.get<{ labels: ApiLabel[] }>('/flow/labels'),

  // -- Milestones ---------------------------------------------------------

  getMilestones: (projectId?: string) =>
    safe.get<{ milestones: ApiMilestone[] }>('/flow/milestones', {
      params: projectId ? { project_id: projectId } : undefined,
    }),

  createMilestone: (data: Partial<ApiMilestone>) =>
    safe.post<{ milestone: ApiMilestone }>('/flow/milestones', data),

  updateMilestone: (id: string, data: Partial<ApiMilestone>) =>
    safe.put<{ message: string }>(`/flow/milestones/${id}`, data),

  // -- Cycles -------------------------------------------------------------

  getCycles: (workspaceId?: string) =>
    safe.get<{ cycles: ApiCycle[] }>('/flow/cycles', {
      params: workspaceId ? { workspace_id: workspaceId } : undefined,
    }),

  createCycle: (data: Partial<ApiCycle>) => safe.post<{ cycle: ApiCycle }>('/flow/cycles', data),

  updateCycle: (id: string, data: Partial<ApiCycle>) =>
    safe.put<{ message: string }>(`/flow/cycles/${id}`, data),

  // -- Views --------------------------------------------------------------

  getViews: () => safe.get<{ views: ApiView[] }>('/flow/views'),

  createView: (data: Partial<ApiView>) => safe.post<{ view: ApiView }>('/flow/views', data),

  updateView: (id: string, data: Partial<ApiView>) =>
    safe.put<{ message: string }>(`/flow/views/${id}`, data),

  deleteView: (id: string) => safe.delete(`/flow/views/${id}`),

  // -- Insights -----------------------------------------------------------

  getInsights: (workspaceId?: string) =>
    safe.get<ApiInsights>('/flow/insights', {
      params: workspaceId ? { workspace_id: workspaceId } : undefined,
    }),

  // -- Team workload ------------------------------------------------------

  getTeamWorkload: (workspaceId?: string) =>
    safe.get<{ members: ApiTeamWorkload[] }>('/flow/team/workload', {
      params: workspaceId ? { workspace_id: workspaceId } : undefined,
    }),

  // -- Roadmap ------------------------------------------------------------

  getRoadmap: (workspaceId?: string) =>
    safe.get<{ projects: ApiRoadmap[] }>('/flow/roadmap', {
      params: workspaceId ? { workspace_id: workspaceId } : undefined,
    }),

  // -- Notifications ------------------------------------------------------

  getNotifications: () => safe.get<{ notifications: ApiNotification[] }>('/flow/notifications'),

  markNotificationRead: (id: string) => safe.put(`/flow/notifications/${id}/read`, {}),

  markAllNotificationsRead: () => safe.put('/flow/notifications/read-all', {}),

  // -- User Settings ------------------------------------------------------

  getUserSettings: () => safe.get<ApiUserSettings>('/flow/settings'),

  updateUserSettings: (data: ApiUserSettings) => safe.put('/flow/settings', data),

  // -- Account / Profile --------------------------------------------------

  updateProfile: (data: { name?: string; email?: string }) => safe.patch('/account/profile', data),

  changePassword: (data: { currentPassword: string; newPassword: string }) =>
    safe.post('/account/password', data),
};
