import React from 'react';
import type { UserRole } from '../types';

interface RoleGuardProps {
  allowed: UserRole[];
  role: UserRole | null;
  fallback?: React.ReactNode;
  children: React.ReactNode;
}

export function RoleGuard({ allowed, role, fallback = null, children }: RoleGuardProps) {
  if (!role || !allowed.includes(role)) return <>{fallback}</>;
  return <>{children}</>;
}
