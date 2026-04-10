import { useQuery } from '@tanstack/react-query';
import api from '../../lib/api';

export interface ActivityEntry {
  id: string;
  actor: string;
  actorId?: string;
  action: string;
  target: string;
  app: 'hub' | 'pos' | 'erp' | 'hr' | 'crm' | 'finance' | 'flow';
  ip?: string;
  createdAt: string;
  severity?: 'info' | 'warning' | 'critical';
}

export const useActivity = (limit = 100) => {
  return useQuery<ActivityEntry[]>({
    queryKey: ['activity', limit],
    queryFn: async () => {
      const res = await api.get(`/activity?limit=${limit}`);
      const raw = res.data?.activities ?? res.data?.entries ?? res.data;
      return Array.isArray(raw) ? raw : [];
    },
    staleTime: 20_000,
  });
};

export const useAuditLog = (limit = 200) => {
  return useQuery<ActivityEntry[]>({
    queryKey: ['audit', limit],
    queryFn: async () => {
      const res = await api.get(`/audit?limit=${limit}`);
      return res.data as ActivityEntry[];
    },
    staleTime: 20_000,
  });
};
