import { can } from '@voxire/permissions';
import type { UserRole } from '@voxire/auth';
import type { Resource } from '@voxire/permissions';

export type CRMPermission = Extract<Resource, `crm:${string}` | `hub:team:${string}`>;

export function canCrmPermission(
  role: UserRole | null | undefined,
  permission: CRMPermission
): boolean {
  return role ? can({ platformRole: role }, permission) : false;
}
