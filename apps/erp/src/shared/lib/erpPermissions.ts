import { can, canEffectiveAccess, canonicalizePermission } from '@voxire/permissions';
import type { UserRole } from '@voxire/auth';
import type { Resource } from '@voxire/permissions';
import { useUserStore } from '@voxire/auth';
import type { EffectiveAccess } from '@voxire/types';

export type ErpPermission = Extract<
  Resource,
  `erp:${string}` | 'hub:team:invite' | 'hub:team:remove'
>;

function storedEffectiveAccess(): EffectiveAccess | undefined {
  return typeof useUserStore.getState === 'function'
    ? useUserStore.getState().user?.effectiveAccess
    : undefined;
}

export function canErpPermission(
  role: UserRole | null | undefined,
  permission: ErpPermission,
  effectiveAccess: EffectiveAccess | null | undefined = storedEffectiveAccess()
): boolean {
  if (effectiveAccess)
    return canEffectiveAccess(effectiveAccess, canonicalizePermission(permission));
  return role ? can({ platformRole: role }, permission) : false;
}
