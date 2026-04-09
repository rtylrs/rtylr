import api from '../lib/api';

// ── Types ──────────────────────────────────────────────────────

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
}

export interface ApiLabel {
  id: string;
  name: string;
  color: string;
}

// ── Dashboard ──────────────────────────────────────────────────

export const flowApi = {
  getDashboard: (workspaceId?: string) =>
    api.get<DashboardStats>('/flow/dashboard', { params: workspaceId ? { workspaceId } : undefined })
      .then((r) => r.data),

  // ── Workspaces ───────────────────────────────────────────────

  getWorkspaces: () =>
    api.get<ApiWorkspace[]>('/flow/workspaces').then((r) => r.data),

  createWorkspace: (data: { name: string; color: string }) =>
    api.post<ApiWorkspace>('/flow/workspaces', data).then((r) => r.data),

  updateWorkspace: (id: string, data: Partial<{ name: string; color: string }>) =>
    api.post<ApiWorkspace>(`/flow/workspaces/${id}`, data).then((r) => r.data),

  deleteWorkspace: (id: string) =>
    api.post(`/flow/workspaces/${id}/delete`).then((r) => r.data),

  addWorkspaceMember: (data: { workspaceId: string; userId: string; role?: string }) =>
    api.post('/flow/workspaces/member', data).then((r) => r.data),

  removeWorkspaceMember: (userId: string, workspaceId: string) =>
    api.post(`/flow/workspaces/member/${userId}`, { workspaceId }).then((r) => r.data),

  // ── Projects ─────────────────────────────────────────────────

  getProjects: (workspaceId: string) =>
    api.get<ApiProject[]>('/flow/projects', { params: { workspaceId } }).then((r) => r.data),

  createProject: (data: { workspaceId: string; name: string; color?: string }) =>
    api.post<ApiProject>('/flow/projects', data).then((r) => r.data),

  updateProject: (id: string, data: Partial<{ name: string; color: string }>) =>
    api.post<ApiProject>(`/flow/projects/${id}`, data).then((r) => r.data),

  deleteProject: (id: string) =>
    api.post(`/flow/projects/${id}/delete`).then((r) => r.data),

  // ── Tasks ─────────────────────────────────────────────────────

  getTasks: (params: { projectId?: string; status?: ApiTask['status']; assigneeId?: string }) =>
    api.get<ApiTask[]>('/flow/tasks', { params }).then((r) => r.data),

  getMyTasks: () =>
    api.get<ApiTask[]>('/flow/tasks/my').then((r) => r.data),

  getTaskDetail: (id: string) =>
    api.get<ApiTask>('/flow/tasks/detail', { params: { id } }).then((r) => r.data),

  createTask: (data: {
    projectId: string;
    workspaceId: string;
    title: string;
    status?: ApiTask['status'];
    priority?: ApiTask['priority'];
    assigneeId?: string;
    dueDate?: string;
  }) =>
    api.post<ApiTask>('/flow/tasks', data).then((r) => r.data),

  updateTask: (id: string, data: Partial<ApiTask>) =>
    api.post<ApiTask>(`/flow/tasks/${id}`, data).then((r) => r.data),

  deleteTask: (id: string) =>
    api.post(`/flow/tasks/${id}/delete`).then((r) => r.data),

  updateTaskStatus: (id: string, status: ApiTask['status']) =>
    api.post('/flow/tasks/status', { id, status }).then((r) => r.data),

  addComment: (data: { taskId: string; content: string }) =>
    api.post('/flow/tasks/comment', data).then((r) => r.data),

  // ── Labels ────────────────────────────────────────────────────

  getLabels: () =>
    api.get<ApiLabel[]>('/flow/labels').then((r) => r.data),

  createLabel: (data: { name: string; color: string }) =>
    api.post<ApiLabel>('/flow/labels', data).then((r) => r.data),
};
