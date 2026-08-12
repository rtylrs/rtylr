import { getOneKitErrorMessage, unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { z } from 'zod';
import { safe } from '../../lib/api';
import { permissionMatrixSchema, customRoleSchema } from '../../schemas/permissions';

export type { PermissionDef, PermissionMatrix, CustomRole } from '../../schemas/permissions';
export type Role = 'owner' | 'admin' | 'manager' | 'staff' | 'viewer' | string;

type PermissionMatrix = z.infer<typeof permissionMatrixSchema>;

export const usePermissions = () => {
  return useQuery<PermissionMatrix>({
    queryKey: ['permissions'],
    queryFn: async () => {
      const res = await unwrap(safe.get('/hub/permission/v1/permissions'));
      return permissionMatrixSchema.parse(res);
    },
  });
};

export const useCustomRoles = () => {
  return useQuery({
    queryKey: ['custom-roles'],
    queryFn: async () => {
      const res = await unwrap(safe.get('/hub/permission/v1/roles'));
      const raw = res?.roles ?? res;
      return z.array(customRoleSchema).parse(Array.isArray(raw) ? raw : []);
    },
  });
};

export const useUpdatePermissions = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (payload: { role: string; key: string; enabled: boolean }) => {
      return unwrap(safe.put('/hub/permission/v1/permissions', payload));
    },
    onSuccess: () => {
      toast.success('Permission updated');
      qc.invalidateQueries({ queryKey: ['permissions'] });
    },
    onError: (error) => toast.error(getOneKitErrorMessage(error, 'Failed to update permission')),
  });
};

export const useCreateCustomRole = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (payload: { name: string; description?: string }) => {
      return unwrap(safe.post('/hub/permission/v1/roles', payload));
    },
    onSuccess: () => {
      toast.success('Role created');
      qc.invalidateQueries({ queryKey: ['custom-roles'] });
    },
    onError: (error) => toast.error(getOneKitErrorMessage(error, 'Failed to create role')),
  });
};
