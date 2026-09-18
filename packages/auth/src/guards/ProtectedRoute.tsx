import React, { useEffect } from 'react';
import { resolveAuthUrl } from '@voxire/config';
import { getCookie } from '../cookies';
import type { AppId } from '../types';

interface ProtectedRouteProps {
  appId: AppId;
  children: React.ReactNode;
}

export function ProtectedRoute({ appId, children }: ProtectedRouteProps) {
  const token = getCookie('token');

  useEffect(() => {
    if (!token) {
      const baseUrl = resolveAuthUrl();
      const u = new URL(baseUrl);
      u.searchParams.set('redirect', appId);
      window.location.href = u.toString();
    }
  }, [token, appId]);

  if (!token) {
    return null;
  }

  return <>{children}</>;
}
