import React from 'react';
import { APP_URLS } from '@voxire/config';
import { getCookie } from '../cookies';
import type { AppId } from '../types';

interface ProtectedRouteProps {
  appId: AppId;
  children: React.ReactNode;
}

export function ProtectedRoute({ appId, children }: ProtectedRouteProps) {
  const token = getCookie('token');

  if (!token) {
    window.location.href = `${APP_URLS.auth}?redirect=${appId}`;
    return null;
  }

  return <>{children}</>;
}
