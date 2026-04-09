import { useQuery } from '@tanstack/react-query';
import api from '../../lib/api';

export interface ActivityEntry {
  id: string;
  actor: string;
  actorId?: string;
  action: string;
  target: string;
  app: 'hub' | 'pos' | 'erp' | 'hr' | 'crm' | 'finance' | 'tracker';
  ip?: string;
  createdAt: string;
  severity?: 'info' | 'warning' | 'critical';
}

const MOCK: ActivityEntry[] = [
  { id: 'a1',  actor: 'Owner Account',  action: 'signed in',                target: 'Hub',                     app: 'hub',     createdAt: new Date(Date.now() - 5 * 60e3).toISOString(),  severity: 'info' },
  { id: 'a2',  actor: 'Amelia Nguyen',  action: 'created order',            target: 'Order #1284',             app: 'pos',     createdAt: new Date(Date.now() - 20 * 60e3).toISOString(), severity: 'info' },
  { id: 'a3',  actor: 'Jorge Castillo', action: 'updated inventory',        target: 'SKU-00411 Tomatoes 5kg',  app: 'erp',     createdAt: new Date(Date.now() - 45 * 60e3).toISOString(), severity: 'info' },
  { id: 'a4',  actor: 'Priya Shah',     action: 'clocked in',               target: 'Shift at Main',           app: 'hr',      createdAt: new Date(Date.now() - 2 * 3600e3).toISOString(), severity: 'info' },
  { id: 'a5',  actor: 'Owner Account',  action: 'changed billing plan',     target: 'Pro monthly',             app: 'hub',     createdAt: new Date(Date.now() - 4 * 3600e3).toISOString(), severity: 'warning' },
  { id: 'a6',  actor: 'System',         action: 'rotated API key',          target: 'key_********ab12',        app: 'hub',     createdAt: new Date(Date.now() - 8 * 3600e3).toISOString(), severity: 'warning' },
  { id: 'a7',  actor: 'Amelia Nguyen',  action: 'invited member',           target: 'liam@example.com',        app: 'hub',     createdAt: new Date(Date.now() - 24 * 3600e3).toISOString(), severity: 'info' },
  { id: 'a8',  actor: 'System',         action: 'failed login attempt',     target: 'unknown@acme.com',        app: 'hub',     createdAt: new Date(Date.now() - 30 * 3600e3).toISOString(), severity: 'critical' },
  { id: 'a9',  actor: 'Mika Tanaka',    action: 'exported customers',       target: 'CRM CSV',                 app: 'crm',     createdAt: new Date(Date.now() - 48 * 3600e3).toISOString(), severity: 'info' },
  { id: 'a10', actor: 'Owner Account',  action: 'approved invoice',         target: 'INV-2026-0317',           app: 'finance', createdAt: new Date(Date.now() - 72 * 3600e3).toISOString(), severity: 'info' },
];

export const useActivity = (limit = 100) => {
  return useQuery<ActivityEntry[]>({
    queryKey: ['activity', limit],
    queryFn: async () => {
      try {
        const res = await api.get(`/activity?limit=${limit}`);
        return (res.data?.entries ?? res.data) as ActivityEntry[];
      } catch {
        return MOCK.slice(0, limit);
      }
    },
    staleTime: 20_000,
  });
};

export const useAuditLog = (limit = 200) => {
  return useQuery<ActivityEntry[]>({
    queryKey: ['audit', limit],
    queryFn: async () => {
      try {
        const res = await api.get(`/audit?limit=${limit}`);
        return res.data as ActivityEntry[];
      } catch {
        return MOCK.map((e) => ({ ...e, ip: '127.0.0.1' }));
      }
    },
    staleTime: 20_000,
  });
};
