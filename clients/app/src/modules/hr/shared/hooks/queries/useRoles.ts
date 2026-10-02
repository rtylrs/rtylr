import { unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { z } from 'zod';
import { hrTransport } from '../../lib/clients';
import { roleSchema, permissionCatalogEntrySchema } from '../../schemas/rolesPermissions';
import type { Role, PermissionCatalogEntry } from '../../schemas/rolesPermissions';

export const useListRoles = () => {
  return useQuery<Role[]>({
    queryKey: ['hr', 'roles'],
    queryFn: async (): Promise<Role[]> => {
      const res = await unwrap(hrTransport.get<{ roles: unknown[] }>('/hr/roles'));
      return z.array(roleSchema).parse((res as { roles: unknown[] }).roles);
    },
  });
};

export const useListPermissionCatalog = () => {
  return useQuery<PermissionCatalogEntry[]>({
    queryKey: ['hr', 'roles', 'permission-catalog'],
    queryFn: async (): Promise<PermissionCatalogEntry[]> => {
      const res = await unwrap(hrTransport.get<{ permissions: unknown[] }>('/hr/permissions'));
      return z
        .array(permissionCatalogEntrySchema)
        .parse((res as { permissions: unknown[] }).permissions);
    },
  });
};

export type CreateRoleInput = {
  name: string;
  description: string;
  permissions: string;
};

export const useCreateRole = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation('hrCommon');
  return useMutation({
    mutationFn: async (data: CreateRoleInput): Promise<Role> => {
      const res = await unwrap(hrTransport.post<{ role: unknown }>('/hr/roles', data));
      return roleSchema.parse((res as { role: unknown }).role);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['hr', 'roles'] });
      toast.success(t('hrRolesPermissions.created'));
    },
    onError: () => toast.error(t('hrErrors.generic')),
  });
};

export type UpdateRoleInput = {
  id: number;
  name: string;
  description: string;
  permissions: string;
};

export const useUpdateRole = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation('hrCommon');
  return useMutation({
    mutationFn: async ({ id, ...data }: UpdateRoleInput): Promise<Role> => {
      const res = await unwrap(hrTransport.put<{ role: unknown }>(`/hr/roles/${id}`, data));
      return roleSchema.parse((res as { role: unknown }).role);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['hr', 'roles'] });
      toast.success(t('hrRolesPermissions.updated'));
    },
    onError: () => toast.error(t('hrErrors.generic')),
  });
};
