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
  workspace_id: number;
  name: string;
  color: string;
  created_at: string;
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
  task_id: number;
  user_id: number;
  action: string;
  old_value: string;
  new_value: string;
  created_at: string;
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
  task_id: number;
  user_id: number;
  duration_minutes: number;
  description: string;
  logged_at: string;
  created_at: string;
}

export interface ApiTaskDetail {
  task: ApiTask;
  subtasks: ApiSubtask[];
  comments: ApiTaskComment[];
  activity: ApiTaskActivity[];
  labels?: ApiLabel[];
  estimated_hours?: number | null;
  actual_hours?: number | null;
}

// -- API ------------------------------------------------------------------

// -- Per-feature services (re-exported for backwards compat) ---------------

export { projectService } from './projectService';
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
    safe.post<ApiWorkspace>('/flow/workspaces', data),

  updateWorkspace: (id: string, data: Partial<{ name: string; color: string }>) =>
    safe.post<ApiWorkspace>(`/flow/workspaces/${id}`, data),

  deleteWorkspace: (id: string) => safe.post(`/flow/workspaces/${id}/delete`),

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
    safe.post<ApiProject>('/flow/projects', data),

  updateProject: (id: string, data: Partial<{ name: string; color: string }>) =>
    safe.post<ApiProject>(`/flow/projects/${id}`, data),

  deleteProject: (id: string) => safe.post(`/flow/projects/${id}/delete`),

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

  searchTasks: (q: string, workspaceId: string) =>
    safe.get<{ tasks: ApiTask[] }>('/flow/tasks/search', {
      params: { q, workspace_id: workspaceId },
    }),

  getTaskDetail: (id: string) => safe.get<ApiTaskDetail>('/flow/tasks/detail', { params: { id } }),

  createTask: (data: {
    projectId: string;
    workspaceId: string;
    title: string;
    status?: ApiTask['status'];
    priority?: ApiTask['priority'];
    assigneeId?: string;
    dueDate?: string;
  }) => safe.post<ApiTask>('/flow/tasks', data),

  updateTask: (id: string, data: Partial<ApiTask>) => safe.post<ApiTask>(`/flow/tasks/${id}`, data),

  deleteTask: (id: string) => safe.post(`/flow/tasks/${id}/delete`),

  updateTaskStatus: (id: string, status: ApiTask['status']) =>
    safe.post('/flow/tasks/status', { id, status }),

  addComment: (data: { taskId: string; content: string }) => safe.post('/flow/tasks/comment', data),

  deleteComment: (id: string) => safe.delete(`/flow/tasks/comment/${id}`),

  // -- Task Activity ------------------------------------------------------

  getTaskActivity: (taskId: string) =>
    safe.get<{ activity: ApiTaskActivity[] }>('/flow/tasks/activity', {
      params: { task_id: taskId },
    }),

  // -- Subtasks -----------------------------------------------------------

  getSubtasks: (id: string) => safe.get<{ subtasks: ApiSubtask[] }>(`/flow/tasks/${id}/subtasks`),

  createSubtask: (taskId: string, data: { title: string; assigneeId?: string }) =>
    safe.post<ApiSubtask>(`/flow/tasks/${taskId}/subtasks`, data),

  // -- Attachments --------------------------------------------------------

  getAttachments: (id: string) =>
    safe.get<{ attachments: ApiAttachment[] }>(`/flow/tasks/${id}/attachments`),

  // -- Time Logs ----------------------------------------------------------

  getTimeLogs: (taskId: string) =>
    safe.get<{ logs: ApiTimeLog[]; estimated_hours: number | null; actual_hours: number | null }>(
      '/flow/tasks/time',
      { params: { task_id: taskId } }
    ),

  logTime: (data: { task_id: number; duration_minutes: number; description?: string }) =>
    safe.post<ApiTimeLog>('/flow/tasks/time', data),

  deleteTimeLog: (id: number) => safe.delete('/flow/tasks/time', { params: { id } }),

  // -- Labels -------------------------------------------------------------

  listLabels: (workspaceId: string) =>
    safe.get<{ labels: ApiLabel[] }>('/flow/labels', { params: { workspace_id: workspaceId } }),

  createLabel: (data: { workspace_id: number; name: string; color: string }) =>
    safe.post<ApiLabel>('/flow/labels', data),

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
    safe.post<ApiMilestone>('/flow/milestones', data),

  updateMilestone: (id: string, data: Partial<ApiMilestone>) =>
    safe.put<ApiMilestone>(`/flow/milestones/${id}`, data),

  // -- Cycles -------------------------------------------------------------

  getCycles: (workspaceId?: string) =>
    safe.get<{ cycles: ApiCycle[] }>('/flow/cycles', {
      params: workspaceId ? { workspace_id: workspaceId } : undefined,
    }),

  createCycle: (data: Partial<ApiCycle>) => safe.post<ApiCycle>('/flow/cycles', data),

  updateCycle: (id: string, data: Partial<ApiCycle>) =>
    safe.put<ApiCycle>(`/flow/cycles/${id}`, data),

  // -- Views --------------------------------------------------------------

  getViews: () => safe.get<{ views: ApiView[] }>('/flow/views'),

  createView: (data: Partial<ApiView>) => safe.post<ApiView>('/flow/views', data),

  updateView: (id: string, data: Partial<ApiView>) => safe.put<ApiView>(`/flow/views/${id}`, data),

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

  updateProfile: (data: { name?: string; email?: string }) => safe.post('/account/profile', data),

  changePassword: (data: { currentPassword: string; newPassword: string }) =>
    safe.post('/account/password', data),

  // -- Businesses ----------------------------------------------------------

  getBusiness: (workspaceId: string) =>
    safe.get<{ business: { id: string; name: string; color: string; icon: string } | null }>(
      '/flow/workspaces/business',
      { params: { workspace_id: workspaceId } }
    ),

  listBusinesses: () =>
    safe.get<{ businesses: { id: string; name: string; color: string; icon: string }[] }>(
      '/flow/businesses'
    ),

  createBusiness: (data: { name: string; color?: string; icon?: string }) =>
    safe.post<{ id: string; name: string; color: string; icon: string }>('/flow/businesses', data),

  setWorkspaceBusiness: (workspaceId: string, businessId: string | null) =>
    safe.put('/flow/workspaces/business', {
      workspace_id: Number(workspaceId),
      business_id: businessId ? Number(businessId) : null,
    }),
};
