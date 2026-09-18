import { can, canEffectiveAccess, canonicalizePermission } from '@voxire/permissions';
import type { UserRole } from '@voxire/auth';
import type { Resource } from '@voxire/permissions';
import { useUserStore } from '@voxire/auth';
import type { EffectiveAccess } from '@voxire/types';

export type CRMPermission = Extract<Resource, `crm:${string}` | `hub:team:${string}`>;

function storedEffectiveAccess(): EffectiveAccess | undefined {
  return typeof useUserStore.getState === 'function'
    ? useUserStore.getState().user?.effectiveAccess
    : undefined;
}

export function canCrmPermission(
  role: UserRole | null | undefined,
  permission: CRMPermission,
  effectiveAccess: EffectiveAccess | null | undefined = storedEffectiveAccess()
): boolean {
  if (effectiveAccess)
    return canEffectiveAccess(effectiveAccess, canonicalizePermission(permission));
  return role ? can({ platformRole: role }, permission) : false;
}
