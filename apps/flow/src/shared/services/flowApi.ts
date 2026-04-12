import { safe } from '../lib/api';

// -- Types ----------------------------------------------------------------

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
  status: 'todo' | 'in_progress' | 'in_review' | 'done';
  priority: 'urgent' | 'high' | 'medium' | 'low';
  assignee: string;
  assigneeInitials: string;
  dueDate: string;
  subtaskCount: number;
  projectName: string;
  estimate?: number;
}

export interface ApiLabel {
  id: string;
  name: string;
  color: string;
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

export interface ApiTaskActivity {
  id: string;
  taskId: string;
  userId: string;
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
  id: string;
  taskId: string;
  userId: string;
  durationMinutes: number;
  description: string;
  loggedAt: string;
}

export interface ApiTaskDetail {
  task: ApiTask;
  subtasks: ApiSubtask[];
  comments: ApiTaskComment[];
  activity: ApiTaskActivity[];
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
      params: workspaceId ? { workspaceId } : undefined,
    }),
  // -- Workspaces ---------------------------------------------------------

  getWorkspaces: () => safe.get<ApiWorkspace[]>('/flow/workspaces'),

  createWorkspace: (data: { name: string; color: string }) =>
    safe.post<ApiWorkspace>('/flow/workspaces', data),

  updateWorkspace: (id: string, data: Partial<{ name: string; color: string }>) =>
    safe.post<ApiWorkspace>(`/flow/workspaces/${id}`, data),

  deleteWorkspace: (id: string) => safe.post(`/flow/workspaces/${id}/delete`),

  addWorkspaceMember: (data: { workspaceId: string; userId: string; role?: string }) =>
    safe.post('/flow/workspaces/member', data),

  removeWorkspaceMember: (userId: string, workspaceId: string) =>
    safe.post(`/flow/workspaces/member/${userId}`, { workspaceId }),

  // -- Projects -----------------------------------------------------------

  getProjects: (workspaceId: string) =>
    safe.get<ApiProject[]>('/flow/projects', { params: { workspaceId } }),

  createProject: (data: { workspaceId: string; name: string; color?: string }) =>
    safe.post<ApiProject>('/flow/projects', data),

  updateProject: (id: string, data: Partial<{ name: string; color: string }>) =>
    safe.post<ApiProject>(`/flow/projects/${id}`, data),

  deleteProject: (id: string) => safe.post(`/flow/projects/${id}/delete`),

  // -- Tasks --------------------------------------------------------------

  getTasks: (params: { projectId?: string; status?: ApiTask['status']; assigneeId?: string }) =>
    safe.get<ApiTask[]>('/flow/tasks', { params }),

  getMyTasks: () => safe.get<ApiTask[]>('/flow/tasks/my'),

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

  getTaskActivity: (id: string) => safe.get<ApiTaskActivity[]>(`/flow/tasks/${id}/activity`),

  // -- Subtasks -----------------------------------------------------------

  getSubtasks: (id: string) => safe.get<ApiSubtask[]>(`/flow/tasks/${id}/subtasks`),

  createSubtask: (taskId: string, data: { title: string; assigneeId?: string }) =>
    safe.post<ApiSubtask>(`/flow/tasks/${taskId}/subtasks`, data),

  // -- Attachments --------------------------------------------------------

  getAttachments: (id: string) => safe.get<ApiAttachment[]>(`/flow/tasks/${id}/attachments`),

  // -- Time Logs ----------------------------------------------------------

  getTimeLogs: (id: string) => safe.get<ApiTimeLog[]>(`/flow/tasks/${id}/time`),

  logTime: (id: string, data: { durationMinutes: number; description?: string }) =>
    safe.post<ApiTimeLog>(`/flow/tasks/${id}/time`, data),

  // -- Labels -------------------------------------------------------------

  getLabels: () => safe.get<ApiLabel[]>('/flow/labels'),

  createLabel: (data: { name: string; color: string }) => safe.post<ApiLabel>('/flow/labels', data),

  // -- Milestones ---------------------------------------------------------

  getMilestones: (projectId?: string) =>
    safe.get<ApiMilestone[]>('/flow/milestones', { params: projectId ? { projectId } : undefined }),

  createMilestone: (data: Partial<ApiMilestone>) =>
    safe.post<ApiMilestone>('/flow/milestones', data),

  updateMilestone: (id: string, data: Partial<ApiMilestone>) =>
    safe.put<ApiMilestone>(`/flow/milestones/${id}`, data),

  // -- Cycles -------------------------------------------------------------

  getCycles: (workspaceId?: string) =>
    safe.get<ApiCycle[]>('/flow/cycles', { params: workspaceId ? { workspaceId } : undefined }),

  createCycle: (data: Partial<ApiCycle>) => safe.post<ApiCycle>('/flow/cycles', data),

  updateCycle: (id: string, data: Partial<ApiCycle>) =>
    safe.put<ApiCycle>(`/flow/cycles/${id}`, data),

  // -- Views --------------------------------------------------------------

  getViews: () => safe.get<ApiView[]>('/flow/views'),

  createView: (data: Partial<ApiView>) => safe.post<ApiView>('/flow/views', data),

  updateView: (id: string, data: Partial<ApiView>) => safe.put<ApiView>(`/flow/views/${id}`, data),

  deleteView: (id: string) => safe.delete(`/flow/views/${id}`),

  // -- Insights -----------------------------------------------------------

  getInsights: (workspaceId?: string) =>
    safe.get<ApiInsights>('/flow/insights', { params: workspaceId ? { workspaceId } : undefined }),

  // -- Team workload ------------------------------------------------------

  getTeamWorkload: () => safe.get<ApiTeamWorkload[]>('/flow/team/workload'),

  // -- Roadmap ------------------------------------------------------------

  getRoadmap: () => safe.get<ApiRoadmap[]>('/flow/roadmap'),
};
