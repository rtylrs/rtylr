import { unwrap } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
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
      const res = await unwrap(safe.get('/api/hub/permission/v1/permissions'));
      return permissionMatrixSchema.parse(res);
    },
  });
};

export const useCustomRoles = () => {
  return useQuery({
    queryKey: ['custom-roles'],
    queryFn: async () => {
      const res = await unwrap(safe.get('/api/hub/permission/v1/roles'));
      const raw = res?.roles ?? res;
      return z.array(customRoleSchema).parse(Array.isArray(raw) ? raw : []);
    },
  });
};
