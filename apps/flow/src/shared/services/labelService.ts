import { safe } from '../lib/api';
import type { ApiLabel } from './flowApi';

export const labelService = {
  getLabels: () => safe.get<ApiLabel[]>('/flow/labels'),

  listLabels: (workspaceId: string) =>
    safe.get<{ labels: ApiLabel[] }>('/flow/labels', { params: { workspace_id: workspaceId } }),

  create: (data: { name: string; color: string }) => safe.post<ApiLabel>('/flow/labels', data),
};
