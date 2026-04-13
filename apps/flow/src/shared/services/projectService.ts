import { safe } from '../lib/api';
import type { ApiProject } from './flowApi';

export const projectService = {
  getProjects: (workspaceId: string) =>
    safe.get<ApiProject[]>('/flow/projects', { params: { workspace_id: workspaceId } }),

  create: (data: { workspaceId: string; name: string; color?: string }) =>
    safe.post<ApiProject>('/flow/projects', data),

  update: (id: string, data: Partial<{ name: string; color: string }>) =>
    safe.put<ApiProject>('/flow/projects', { id: parseInt(id), ...data }),

  delete: (id: string) => safe.delete('/flow/projects', { data: { id: parseInt(id) } }),
};
