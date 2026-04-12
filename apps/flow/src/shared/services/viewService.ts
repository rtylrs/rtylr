import { safe } from '../lib/api';
import type { ApiView } from './flowApi';

export const viewService = {
  getViews: () => safe.get<ApiView[]>('/flow/views'),

  create: (data: Partial<ApiView>) => safe.post<ApiView>('/flow/views', data),

  update: (id: string, data: Partial<ApiView>) => safe.put<ApiView>(`/flow/views/${id}`, data),

  delete: (id: string) => safe.delete(`/flow/views/${id}`),
};
