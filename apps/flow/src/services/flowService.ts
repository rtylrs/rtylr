import api from '../lib/api';

export const flowService = {
  getDashboard: () => api.get('/flow/dashboard'),

  getWorkspaces: () => api.get('/flow/workspaces'),
  createWorkspace: (data: any) => api.post('/flow/workspaces', data),
  updateWorkspace: (id: number, data: any) => api.put(`/flow/workspaces/${id}`, data),
  deleteWorkspace: (id: number) => api.delete(`/flow/workspaces/${id}`),

  getProjects: (params?: any) => api.get('/flow/projects', { params }),
  createProject: (data: any) => api.post('/flow/projects', data),
  updateProject: (id: number, data: any) => api.put(`/flow/projects/${id}`, data),
  deleteProject: (id: number) => api.delete(`/flow/projects/${id}`),

  getTasks: (params?: any) => api.get('/flow/tasks', { params }),
  createTask: (data: any) => api.post('/flow/tasks', data),
  updateTask: (id: number, data: any) => api.put(`/flow/tasks/${id}`, data),
  deleteTask: (id: number) => api.delete(`/flow/tasks/${id}`),

  getLabels: () => api.get('/flow/labels'),
  createLabel: (data: any) => api.post('/flow/labels', data),
};
