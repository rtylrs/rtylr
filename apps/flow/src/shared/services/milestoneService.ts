import { safe } from '../lib/api';
import type { ApiMilestone } from './flowApi';

export const milestoneService = {
  getMilestones: (projectId?: string) =>
    safe.get<ApiMilestone[]>('/flow/milestones', {
      params: projectId ? { project_id: projectId } : undefined,
    }),

  create: (data: Partial<ApiMilestone>) => safe.post<ApiMilestone>('/flow/milestones', data),

  update: (id: string, data: Partial<ApiMilestone>) =>
    safe.put<ApiMilestone>(`/flow/milestones/${id}`, data),
};
