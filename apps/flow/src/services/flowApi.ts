import api from '../lib/api';

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
  status: 'planning' | 'active' | 'at_risk' | 'completed';
  description?: string;
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
  filters: Record<string, any>;
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
}

// -- API ------------------------------------------------------------------

export const flowApi = {
  getDashboard: (workspaceId?: string) =>
    api.get<DashboardStats>('/flow/dashboard', { params: workspaceId ? { workspaceId } : undefined })
      .then((r) => r.data),

  // -- Workspaces ---------------------------------------------------------

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

  // -- Projects -----------------------------------------------------------

  getProjects: (workspaceId: string) =>
    api.get<ApiProject[]>('/flow/projects', { params: { workspaceId } }).then((r) => r.data),

  createProject: (data: { workspaceId: string; name: string; color?: string }) =>
    api.post<ApiProject>('/flow/projects', data).then((r) => r.data),

  updateProject: (id: string, data: Partial<{ name: string; color: string }>) =>
    api.post<ApiProject>(`/flow/projects/${id}`, data).then((r) => r.data),

  deleteProject: (id: string) =>
    api.post(`/flow/projects/${id}/delete`).then((r) => r.data),

  // -- Tasks --------------------------------------------------------------

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

  deleteComment: (id: string) =>
    api.delete(`/flow/tasks/comment/${id}`).then((r) => r.data),

  // -- Labels -------------------------------------------------------------

  getLabels: () =>
    api.get<ApiLabel[]>('/flow/labels').then((r) => r.data),

  createLabel: (data: { name: string; color: string }) =>
    api.post<ApiLabel>('/flow/labels', data).then((r) => r.data),

  // -- Milestones ---------------------------------------------------------

  getMilestones: (projectId?: string) =>
    api.get<ApiMilestone[]>('/flow/milestones', { params: projectId ? { projectId } : undefined }).then((r) => r.data),

  createMilestone: (data: Partial<ApiMilestone>) =>
    api.post<ApiMilestone>('/flow/milestones', data).then((r) => r.data),

  updateMilestone: (id: string, data: Partial<ApiMilestone>) =>
    api.put<ApiMilestone>(`/flow/milestones/${id}`, data).then((r) => r.data),

  // -- Cycles -------------------------------------------------------------

  getCycles: (workspaceId?: string) =>
    api.get<ApiCycle[]>('/flow/cycles', { params: workspaceId ? { workspaceId } : undefined }).then((r) => r.data),

  createCycle: (data: Partial<ApiCycle>) =>
    api.post<ApiCycle>('/flow/cycles', data).then((r) => r.data),

  updateCycle: (id: string, data: Partial<ApiCycle>) =>
    api.put<ApiCycle>(`/flow/cycles/${id}`, data).then((r) => r.data),

  // -- Views --------------------------------------------------------------

  getViews: () =>
    api.get<ApiView[]>('/flow/views').then((r) => r.data),

  createView: (data: Partial<ApiView>) =>
    api.post<ApiView>('/flow/views', data).then((r) => r.data),

  updateView: (id: string, data: Partial<ApiView>) =>
    api.put<ApiView>(`/flow/views/${id}`, data).then((r) => r.data),

  deleteView: (id: string) =>
    api.delete(`/flow/views/${id}`).then((r) => r.data),

  // -- Insights -----------------------------------------------------------

  getInsights: (workspaceId?: string) =>
    api.get<ApiInsights>('/flow/insights', { params: workspaceId ? { workspaceId } : undefined }).then((r) => r.data),

  // -- Team workload ------------------------------------------------------

  getTeamWorkload: () =>
    api.get<ApiTeamWorkload[]>('/flow/team/workload').then((r) => r.data),

  // -- Roadmap ------------------------------------------------------------

  getRoadmap: () =>
    api.get<ApiRoadmap[]>('/flow/roadmap').then((r) => r.data),
};
