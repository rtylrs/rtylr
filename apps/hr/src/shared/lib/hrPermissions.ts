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

/** Compensation is an HR-management concern, not a general staff capability. */
export function canViewCompensation(role: UserRole | null | undefined): boolean {
  return canHrPermission(role, 'hr:staff:manage') || canHrPermission(role, 'hr:payroll');
}
