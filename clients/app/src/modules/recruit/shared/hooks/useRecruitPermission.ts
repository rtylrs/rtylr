import { canEffectiveAccess } from '@voxire/permissions';
import { useUserStore } from '../../../../shared/stores/userStore';
export function useRecruitPermission(permission: string): boolean {
  const access = useUserStore((state) => state.user?.effectiveAccess);
  return canEffectiveAccess(access, permission, 'recruit');
}
