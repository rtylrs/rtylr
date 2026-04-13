import { safe } from '../lib/api';
import type { ApiCycle } from './flowApi';

export const cycleService = {
  getCycles: (workspaceId?: string) =>
    safe.get<ApiCycle[]>('/flow/cycles', {
      params: workspaceId ? { workspace_id: workspaceId } : undefined,
    }),

  create: (data: Partial<ApiCycle>) => safe.post<ApiCycle>('/flow/cycles', data),

  update: (id: string, data: Partial<ApiCycle>) => safe.put<ApiCycle>(`/flow/cycles/${id}`, data),
};
