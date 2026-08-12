import { can } from '@voxire/permissions';
import type { UserRole } from '@voxire/auth';
import type { Resource } from '@voxire/permissions';

export type HRPermission = Extract<Resource, `hr:${string}` | 'hub:settings' | 'insights:reports'>;

export function canHrPermission(
  role: UserRole | null | undefined,
  permission: HRPermission
): boolean {
  return role ? can({ platformRole: role }, permission) : false;
}
