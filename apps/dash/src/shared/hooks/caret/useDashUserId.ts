import { useUserStore } from '../../stores/userStore';

/** Returns the numeric user id from the user store, or 0 when logged-out. */
export function useDashUserId(): number {
  const id = useUserStore((s) => s.user?.id);
  return Number(id) || 0;
}
