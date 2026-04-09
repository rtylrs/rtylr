import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import api from '../../lib/api';
import { toast } from 'sonner';

export type Role = 'owner' | 'admin' | 'manager' | 'staff' | 'viewer' | string;

export interface PermissionDef {
  key: string;
  label: string;
  group: string;
  description?: string;
}

export interface PermissionMatrix {
  permissions: PermissionDef[];
  roles: Role[];
  // role key -> permission key -> boolean
  values: Record<Role, Record<string, boolean>>;
}

const PERMS: PermissionDef[] = [
  { key: 'team.view',            label: 'View team',           group: 'Team' },
  { key: 'team.invite',          label: 'Invite members',      group: 'Team' },
  { key: 'team.edit',            label: 'Edit members',        group: 'Team' },
  { key: 'team.deactivate',      label: 'Deactivate members',  group: 'Team' },
  { key: 'billing.view',         label: 'View billing',        group: 'Billing' },
  { key: 'billing.manage',       label: 'Manage billing',      group: 'Billing' },
  { key: 'apps.launch',          label: 'Launch apps',         group: 'Apps' },
  { key: 'apps.manage',          label: 'Manage app access',   group: 'Apps' },
  { key: 'settings.view',        label: 'View settings',       group: 'Settings' },
  { key: 'settings.edit',        label: 'Edit workspace',      group: 'Settings' },
  { key: 'api_keys.manage',      label: 'Manage API keys',     group: 'Developer' },
  { key: 'webhooks.manage',      label: 'Manage webhooks',     group: 'Developer' },
  { key: 'audit.view',           label: 'View audit log',      group: 'Security' },
  { key: 'security.manage',      label: 'Manage security',     group: 'Security' },
];

const defaultMatrix: PermissionMatrix = {
  permissions: PERMS,
  roles: ['owner', 'admin', 'manager', 'staff', 'viewer'],
  values: {
    owner:   Object.fromEntries(PERMS.map((p) => [p.key, true])),
    admin:   Object.fromEntries(PERMS.map((p) => [p.key, !p.key.startsWith('billing.manage') && p.key !== 'settings.edit' ? true : p.key === 'settings.edit'])),
    manager: Object.fromEntries(PERMS.map((p) => [p.key, ['team.view','team.invite','apps.launch','settings.view','audit.view'].includes(p.key)])),
    staff:   Object.fromEntries(PERMS.map((p) => [p.key, ['apps.launch','settings.view','team.view'].includes(p.key)])),
    viewer:  Object.fromEntries(PERMS.map((p) => [p.key, ['team.view','apps.launch','settings.view'].includes(p.key)])),
  },
};

export const usePermissions = () => {
  return useQuery<PermissionMatrix>({
    queryKey: ['permissions'],
    queryFn: async () => {
      try {
        const res = await api.get('/permissions');
        return res.data as PermissionMatrix;
      } catch {
        return defaultMatrix;
      }
    },
  });
};

export const useUpdatePermission = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (payload: { role: Role; permission: string; value: boolean }) => {
      try {
        await api.post('/permissions', payload);
      } catch {
        // ignore
      }
      return payload;
    },
    onMutate: async ({ role, permission, value }) => {
      await qc.cancelQueries({ queryKey: ['permissions'] });
      const prev = qc.getQueryData<PermissionMatrix>(['permissions']);
      if (prev) {
        qc.setQueryData<PermissionMatrix>(['permissions'], {
          ...prev,
          values: {
            ...prev.values,
            [role]: { ...(prev.values[role] ?? {}), [permission]: value },
          },
        });
      }
      return { prev };
    },
    onError: (_e, _v, ctx) => {
      if (ctx?.prev) qc.setQueryData(['permissions'], ctx.prev);
      toast.error('Failed to update permission');
    },
  });
};

export interface CustomRole {
  id: string;
  name: string;
  color: string;
  permissions: string[];
  members: number;
}

const MOCK_ROLES: CustomRole[] = [
  { id: 'r_floor',    name: 'Floor Lead',      color: '#4583C4', permissions: ['team.view','apps.launch','settings.view'], members: 3 },
  { id: 'r_kitchen',  name: 'Kitchen',         color: '#f59e0b', permissions: ['apps.launch'],                              members: 5 },
  { id: 'r_cashier',  name: 'Cashier',         color: '#10b981', permissions: ['apps.launch'],                              members: 8 },
  { id: 'r_analyst',  name: 'Analyst',         color: '#8b5cf6', permissions: ['apps.launch','team.view','audit.view'],     members: 2 },
];

export const useCustomRoles = () => {
  return useQuery<CustomRole[]>({
    queryKey: ['custom-roles'],
    queryFn: async () => {
      try {
        const res = await api.get('/roles');
        return (res.data?.roles ?? res.data) as CustomRole[];
      } catch {
        return MOCK_ROLES;
      }
    },
  });
};
