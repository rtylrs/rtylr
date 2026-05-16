import type { StoredUser } from '../../stores/userStore';

/** Best display name for the current user (fullName → email → "You"). */
export function userLabel(user: StoredUser | null | undefined): string {
  if (!user) return 'You';
  if (user.fullName) return user.fullName;
  if (typeof user.email === 'string' && user.email) return user.email;
  return 'You';
}
