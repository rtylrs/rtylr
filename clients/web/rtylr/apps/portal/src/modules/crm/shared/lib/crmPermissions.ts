import { canEffectiveAccess } from '@voxire/permissions';
import type { EffectiveAccess, PermissionKey } from '@voxire/types';
import { useUserStore } from '../../../../shared/stores/userStore';

export type CRMPermission = PermissionKey & (`crm:${string}` | `hub:team:${string}`);

function storedEffectiveAccess(): EffectiveAccess | undefined {
  const state = useUserStore.getState();
  return state.user?.effectiveAccess;
}

export function canCrmPermission(
  permission: CRMPermission,
  effectiveAccess: EffectiveAccess | null | undefined = storedEffectiveAccess()
): boolean {
  return canEffectiveAccess(effectiveAccess, permission);
}
