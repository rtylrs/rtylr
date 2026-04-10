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

export const usePermissions = () => {
  return useQuery<PermissionMatrix>({
    queryKey: ['permissions'],
    queryFn: async () => {
      const res = await api.get('/permissions');
      return res.data as PermissionMatrix;
    },
  });
};

export const useUpdatePermission = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (payload: { role: Role; permission: string; value: boolean }) => {
      await api.post('/permissions', payload);
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

export const useCustomRoles = () => {
  return useQuery<CustomRole[]>({
    queryKey: ['custom-roles'],
    queryFn: async () => {
      const res = await api.get('/roles');
      return (res.data?.roles ?? res.data) as CustomRole[];
    },
  });
};
