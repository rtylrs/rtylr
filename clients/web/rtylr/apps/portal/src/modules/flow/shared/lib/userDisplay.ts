import type { FlowUser } from '../stores/userStore';

export function displayUserEmail(user: FlowUser | null | undefined): string {
  return user?.email?.trim() ?? '';
}

export function displayUserName(user: FlowUser | null | undefined): string {
  if (!user) return '';
  return user.fullName?.trim() || displayUserEmail(user) || '';
}
