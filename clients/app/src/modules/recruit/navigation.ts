import { Briefcase, Inbox, LayoutGrid, Palette } from 'lucide-react';
import { canEffectiveAccess, hasEffectiveAppAccess } from '@voxire/permissions';
import type { StoredUser } from '../../shared/stores/userStore';
export const recruitNavItems = [
  { key: 'careers', path: '/recruit/careers', icon: Briefcase, permission: 'recruit:jobs:view' },
  {
    key: 'submissions',
    path: '/recruit/submissions',
    icon: Inbox,
    permission: 'recruit:candidates:view',
  },
  {
    key: 'sections',
    path: '/recruit/sections',
    icon: LayoutGrid,
    permission: 'recruit:jobs:manage',
  },
  { key: 'branding', path: '/settings/recruit', icon: Palette, permission: 'recruit:jobs:manage' },
] as const;
export function canAccessRecruitPath(user: StoredUser | null, path: string): boolean {
  const access = user?.effectiveAccess;
  if (!hasEffectiveAppAccess(access, 'recruit')) return false;
  if (path === '/recruit' || access?.isOwner || access?.isAdmin) return true;
  if (path === '/recruit/customization') path = '/settings/recruit';
  const item = recruitNavItems.find(
    (item) => item.path === path || path.startsWith(`${item.path}/`)
  );
  return (
    !!item &&
    (canEffectiveAccess(access, item.permission, 'recruit') ||
      (item.key === 'careers' && canEffectiveAccess(access, 'recruit:jobs:manage', 'recruit')))
  );
}
export function getRecruitHomePath(user: StoredUser | null): string {
  return recruitNavItems.find((item) => canAccessRecruitPath(user, item.path))?.path ?? '/';
}
