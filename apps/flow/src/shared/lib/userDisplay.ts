import type { FlowUser } from '../stores/userStore';

/** Align with Hub `DefaultLayout` sidebar user line. */
export function displayUserEmail(user: FlowUser | null | undefined): string {
  return user?.email?.trim() ?? '';
}

export function displayUserName(user: FlowUser | null | undefined): string {
  if (!user) return '';
  return user.name?.trim() || user.firstName?.trim() || displayUserEmail(user) || '';
}
