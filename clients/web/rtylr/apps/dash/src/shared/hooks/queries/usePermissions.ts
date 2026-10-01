import { getOneKitErrorMessage } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { z } from 'zod';
import { permissionClient } from '../../lib/clients';
import { permissionCatalogSchema, permissionProfileSchema } from '../../schemas/permissions';
import i18next from 'i18next';

export type {
  PermissionDef,
  PermissionCatalog,
  PermissionProfile,
} from '../../schemas/permissions';

export const usePermissionCatalog = () => {
  return useQuery({
    queryKey: ['permission-catalog'],
    queryFn: async () => {
      const response = await permissionClient.listPermissionCatalog({});
      return permissionCatalogSchema.parse(response);
    },
  });
};

export const usePermissionProfiles = () => {
  return useQuery({
    queryKey: ['permission-profiles'],
    queryFn: async () => {
      const roles: z.infer<typeof permissionProfileSchema>[] = [];
      let page = 1;
      for (let requestCount = 0; requestCount < 100; requestCount += 1) {
        const res = await permissionClient.listRoles({ page, pageSize: 100 });
        roles.push(...z.array(permissionProfileSchema).parse(res.roles ?? []));
        if (!res.nextPage) break;
        page = res.nextPage;
      }
      return roles;
    },
  });
};

export const useCreatePermissionProfile = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (payload: {
      name: string;
      description?: string;
      appId?: string;
      permissions?: string[];
    }) => {
      return permissionClient.createRole(payload);
    },
    onSuccess: () => {
      toast.success(i18next.t('common.toasts.roleCreated', 'Role created'));
      qc.invalidateQueries({ queryKey: ['permission-profiles'] });
    },
    onError: (error) =>
      toast.error(
        getOneKitErrorMessage(
          error,
          i18next.t('common.toasts.failedCreateRole', 'Failed to create role')
        )
      ),
  });
};

export const useUpdatePermissionProfile = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (payload: {
      id: string;
      name: string;
      description?: string;
      permissions?: string[];
    }) =>
      permissionClient.updateRole({
        id: payload.id,
        body: {
          name: payload.name,
          description: payload.description,
          permissions: payload.permissions ?? [],
        },
      }),
    onSuccess: () => {
      toast.success(i18next.t('common.toasts.roleUpdated', 'Role updated'));
      qc.invalidateQueries({ queryKey: ['permission-profiles'] });
    },
    onError: (error) =>
      toast.error(
        getOneKitErrorMessage(
          error,
          i18next.t('common.toasts.failedUpdateRole', 'Failed to update role')
        )
      ),
  });
};

export const useArchivePermissionProfile = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (id: string) => permissionClient.deleteRole({ id }),
    onSuccess: () => {
      toast.success(i18next.t('common.toasts.roleArchived', 'Role archived'));
      qc.invalidateQueries({ queryKey: ['permission-profiles'] });
    },
    onError: (error) =>
      toast.error(
        getOneKitErrorMessage(
          error,
          i18next.t('common.toasts.failedArchiveRole', 'Failed to archive role')
        )
      ),
  });
};
