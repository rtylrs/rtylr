import { getOneKitErrorMessage } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { z } from 'zod';
import { permissionClient } from '../../lib/clients';
import { permissionMatrixSchema, customRoleSchema } from '../../schemas/permissions';
import i18next from 'i18next';

export type { PermissionDef, PermissionMatrix, CustomRole } from '../../schemas/permissions';
export type Role = 'owner' | 'admin' | 'manager' | 'staff' | 'viewer' | string;

type PermissionMatrix = z.infer<typeof permissionMatrixSchema>;

export const usePermissions = () => {
  return useQuery<PermissionMatrix>({
    queryKey: ['permissions'],
    queryFn: async () => {
      const res = await permissionClient.getPermissions({});
      return permissionMatrixSchema.parse(res.matrix ?? res);
    },
  });
};

export const useCustomRoles = () => {
  return useQuery({
    queryKey: ['custom-roles'],
    queryFn: async () => {
      const res = await permissionClient.listRoles({});
      return z.array(customRoleSchema).parse(res.roles ?? []);
    },
  });
};

export const useUpdatePermissions = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (payload: { role: string; key: string; enabled: boolean }) => {
      return permissionClient.updatePermissionMatrix({
        role: payload.role,
        permissionKey: payload.key,
        enabled: payload.enabled,
      });
    },
    onSuccess: () => {
      toast.success(i18next.t('toasts.permissionUpdated', 'Permission updated'));
      qc.invalidateQueries({ queryKey: ['permissions'] });
    },
    onError: (error) =>
      toast.error(
        getOneKitErrorMessage(
          error,
          i18next.t('toasts.failedUpdatePermission', 'Failed to update permission')
        )
      ),
  });
};

export const useCreateCustomRole = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (payload: { name: string; description?: string }) => {
      return permissionClient.createRole(payload);
    },
    onSuccess: () => {
      toast.success(i18next.t('toasts.roleCreated', 'Role created'));
      qc.invalidateQueries({ queryKey: ['custom-roles'] });
    },
    onError: (error) =>
      toast.error(
        getOneKitErrorMessage(error, i18next.t('toasts.failedCreateRole', 'Failed to create role'))
      ),
  });
};
