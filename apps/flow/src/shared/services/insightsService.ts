import { safe } from '../lib/api';
import type { ApiInsights, ApiTeamWorkload, ApiRoadmap, DashboardStats } from './flowApi';

export const insightsService = {
  getDashboard: (workspaceId?: string) =>
    safe.get<DashboardStats>('/flow/dashboard', {
      params: workspaceId ? { workspace_id: workspaceId } : undefined,
    }),

  getInsights: (workspaceId?: string) =>
    safe.get<ApiInsights>('/flow/insights', {
      params: workspaceId ? { workspace_id: workspaceId } : undefined,
    }),

  getTeamWorkload: () => safe.get<ApiTeamWorkload[]>('/flow/team/workload'),

  getRoadmap: () => safe.get<ApiRoadmap[]>('/flow/roadmap'),
};
