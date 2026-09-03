import React from 'react';
import { canEffectiveAccess } from '@voxire/permissions';
import type { EffectiveAccess, PermissionKey, TenantAppId } from '@voxire/types';
import type { UserRole } from '../types';

interface RoleGuardProps {
  allowed?: UserRole[];
  role?: UserRole | null;
  appId?: TenantAppId;
  permission?: PermissionKey;
  effectiveAccess?: EffectiveAccess | null;
  fallback?: React.ReactNode;
  children: React.ReactNode;
}

export function RoleGuard({
  allowed = [],
  role = null,
  appId,
  permission,
  effectiveAccess,
  fallback = null,
  children,
}: RoleGuardProps) {
  const allowedByPermission =
    permission !== undefined && canEffectiveAccess(effectiveAccess, permission, appId);
  const allowedByLegacyRole = permission === undefined && role !== null && allowed.includes(role);
  if (!allowedByPermission && !allowedByLegacyRole) return <>{fallback}</>;
  return <>{children}</>;
}
