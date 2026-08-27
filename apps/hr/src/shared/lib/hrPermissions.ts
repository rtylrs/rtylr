import { can, canEffectiveAccess, canonicalizePermission } from '@voxire/permissions';
import type { UserRole } from '@voxire/auth';
import type { Resource } from '@voxire/permissions';
import { useUserStore } from '@voxire/auth';
import type { EffectiveAccess } from '@voxire/types';

export type HRPermission = Extract<Resource, `hr:${string}` | 'hub:settings' | 'insights:reports'>;

function storedEffectiveAccess(): EffectiveAccess | undefined {
  return typeof useUserStore.getState === 'function'
    ? useUserStore.getState().user?.effectiveAccess
    : undefined;
}

export function canHrPermission(
  role: UserRole | null | undefined,
  permission: HRPermission,
  effectiveAccess: EffectiveAccess | null | undefined = storedEffectiveAccess()
): boolean {
  if (effectiveAccess)
    return canEffectiveAccess(effectiveAccess, canonicalizePermission(permission));
  return role ? can({ platformRole: role }, permission) : false;
}

/** Compensation is an HR-management concern, not a general staff capability. */
export function canViewCompensation(role: UserRole | null | undefined): boolean {
  return canHrPermission(role, 'hr:staff:manage') || canHrPermission(role, 'hr:payroll');
}
