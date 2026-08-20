import { can } from '@voxire/permissions';
import type { UserRole } from '@voxire/auth';
import type { Resource } from '@voxire/permissions';

export type ErpPermission = Extract<
  Resource,
  `erp:${string}` | 'hub:team:invite' | 'hub:team:remove'
>;

export function canErpPermission(
  role: UserRole | null | undefined,
  permission: ErpPermission
): boolean {
  return role ? can({ platformRole: role }, permission) : false;
}
