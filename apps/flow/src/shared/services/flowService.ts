import { safe } from '../lib/api';

export interface WorkspacePayload {
  name: string;
  description?: string;
  [key: string]: unknown;
}

export interface ProjectPayload {
  name: string;
  workspaceId?: number;
  description?: string;
  status?: string;
  dueDate?: string;
  [key: string]: unknown;
}

export interface ProjectFilterParams {
  workspaceId?: number;
  status?: string;
  [key: string]: unknown;
}

export interface TaskPayload {
  title: string;
  projectId?: number;
  assigneeId?: number;
  status?: string;
  priority?: string;
  dueDate?: string;
  description?: string;
  [key: string]: unknown;
}

export interface TaskFilterParams {
  projectId?: number;
  assigneeId?: number;
  status?: string;
  priority?: string;
  [key: string]: unknown;
}

export interface LabelPayload {
  name: string;
  color?: string;
  [key: string]: unknown;
}

export const flowService = {
  getDashboard: () => safe.get('/flow/dashboard'),

  getWorkspaces: () => safe.get('/flow/workspaces'),
  createWorkspace: (data: WorkspacePayload) => safe.post('/flow/workspaces', data),
  updateWorkspace: (id: number, data: Partial<WorkspacePayload>) =>
    safe.put('/flow/workspaces', { id, ...data }),
  deleteWorkspace: (id: number) => safe.delete('/flow/workspaces', { data: { id } }),

  getProjects: (params?: ProjectFilterParams) => safe.get('/flow/projects', { params }),
  createProject: (data: ProjectPayload) => safe.post('/flow/projects', data),
  updateProject: (id: number, data: Partial<ProjectPayload>) =>
    safe.put('/flow/projects', { id, ...data }),
  deleteProject: (id: number) => safe.delete('/flow/projects', { data: { id } }),

  getTasks: (params?: TaskFilterParams) => safe.get('/flow/tasks', { params }),
  createTask: (data: TaskPayload) => safe.post('/flow/tasks', data),
  updateTask: (id: number, data: Partial<TaskPayload>) => safe.put('/flow/tasks', { id, ...data }),
  deleteTask: (id: number) => safe.delete('/flow/tasks', { data: { id } }),

  getLabels: () => safe.get('/flow/labels'),
  createLabel: (data: LabelPayload) => safe.post('/flow/labels', data),
};
