import React from 'react';
import { canEffectiveAccess } from '@voxire/permissions';
import type { EffectiveAccess, PermissionKey, TenantAppId } from '@voxire/types';

interface RoleGuardProps {
  appId?: TenantAppId;
  permission?: PermissionKey;
  effectiveAccess?: EffectiveAccess | null;
  fallback?: React.ReactNode;
  children: React.ReactNode;
}

export function RoleGuard({
  appId,
  permission,
  effectiveAccess,
  fallback = null,
  children,
}: RoleGuardProps) {
  const allowed = permission !== undefined && canEffectiveAccess(effectiveAccess, permission, appId);
  if (!allowed) return <>{fallback}</>;
  return <>{children}</>;
}
