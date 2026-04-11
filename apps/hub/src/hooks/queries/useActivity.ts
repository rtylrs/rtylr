import { unwrap } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { safe } from '../../lib/api';

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
      const res = await unwrap(safe.get(`/activity?limit=${limit}`));
      const raw = res?.activities ?? res?.entries ?? res;
      return Array.isArray(raw) ? raw : [];
    },
    staleTime: 20_000,
  });
};

export const useAuditLog = (limit = 200) => {
  return useQuery<ActivityEntry[]>({
    queryKey: ['audit', limit],
    queryFn: async () => {
      const res = await unwrap(safe.get(`/audit?limit=${limit}`));
      return res as ActivityEntry[];
    },
    staleTime: 20_000,
  });
};
