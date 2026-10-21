import { canEffectiveAccess } from '@voxire/permissions';
import type { EffectiveAccess, PermissionKey } from '@voxire/types';
import { useUserStore } from '../../../../shared/stores/userStore';

export type ErpPermission = PermissionKey & (`erp:${string}` | `hub:team:${string}`);

function storedEffectiveAccess(): EffectiveAccess | undefined {
  const state = useUserStore.getState();
  return state.user?.effectiveAccess;
}

export function canErpPermission(
  permission: ErpPermission,
  effectiveAccess: EffectiveAccess | null | undefined = storedEffectiveAccess()
): boolean {
  return canEffectiveAccess(effectiveAccess, permission);
}
