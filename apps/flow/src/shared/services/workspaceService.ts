import { safe } from '../lib/api';
import type { ApiWorkspace } from './flowApi';

export const workspaceService = {
  getWorkspaces: () => safe.get<ApiWorkspace[]>('/flow/workspaces'),

  create: (data: { name: string; color: string }) =>
    safe.post<ApiWorkspace>('/flow/workspaces', data),

  update: (id: string, data: Partial<{ name: string; color: string }>) =>
    safe.put<ApiWorkspace>('/flow/workspaces', { id: parseInt(id), ...data }),

  delete: (id: string) => safe.delete('/flow/workspaces', { data: { id: parseInt(id) } }),

  addMember: (data: { workspaceId: string; userId: string; role?: string }) =>
    safe.post('/flow/workspaces/member', data),

  removeMember: (userId: string, workspaceId: string) =>
    safe.delete('/flow/workspaces/member', {
      data: { workspaceId: parseInt(workspaceId), userId: parseInt(userId) },
    }),
};
