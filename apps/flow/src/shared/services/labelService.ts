import { safe } from '../lib/api';
import type { ApiLabel } from './flowApi';

export const labelService = {
  getLabels: () => safe.get<ApiLabel[]>('/flow/labels'),

  create: (data: { name: string; color: string }) => safe.post<ApiLabel>('/flow/labels', data),
};
