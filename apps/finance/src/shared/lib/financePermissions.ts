import { can } from '@voxire/permissions';
import type { UserRole } from '@voxire/auth';
import type { Resource } from '@voxire/permissions';

export type FinancePermission = Extract<
  Resource,
  `finance:${string}` | 'hub:team:invite' | 'hub:team:remove'
>;

export function canFinancePermission(
  role: UserRole | null | undefined,
  permission: FinancePermission
): boolean {
  return role ? can({ platformRole: role }, permission) : false;
}
