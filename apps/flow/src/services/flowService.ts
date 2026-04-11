import { safe } from '../lib/api';

export const flowService = {
  getDashboard: () => safe.get('/flow/dashboard'),

  getWorkspaces: () => safe.get('/flow/workspaces'),
  createWorkspace: (data: any) => safe.post('/flow/workspaces', data),
  updateWorkspace: (id: number, data: any) => safe.put(`/flow/workspaces/${id}`, data),
  deleteWorkspace: (id: number) => safe.delete(`/flow/workspaces/${id}`),

  getProjects: (params?: any) => safe.get('/flow/projects', { params }),
  createProject: (data: any) => safe.post('/flow/projects', data),
  updateProject: (id: number, data: any) => safe.put(`/flow/projects/${id}`, data),
  deleteProject: (id: number) => safe.delete(`/flow/projects/${id}`),

  getTasks: (params?: any) => safe.get('/flow/tasks', { params }),
  createTask: (data: any) => safe.post('/flow/tasks', data),
  updateTask: (id: number, data: any) => safe.put(`/flow/tasks/${id}`, data),
  deleteTask: (id: number) => safe.delete(`/flow/tasks/${id}`),

  getLabels: () => safe.get('/flow/labels'),
  createLabel: (data: any) => safe.post('/flow/labels', data),
};
