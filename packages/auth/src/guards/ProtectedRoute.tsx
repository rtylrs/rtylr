import React from 'react';
import { getCookie } from '../cookies';
import type { AppId } from '../types';

const AUTH_URL = typeof window !== 'undefined' && window.location.hostname === 'localhost'
  ? 'http://localhost:5176'
  : 'https://auth.rtylr.com';

interface ProtectedRouteProps {
  appId: AppId;
  children: React.ReactNode;
}

export function ProtectedRoute({ appId, children }: ProtectedRouteProps) {
  const token = getCookie('token');

  if (!token) {
    window.location.href = `${AUTH_URL}?redirect=${appId}`;
    return null;
  }

  return <>{children}</>;
}
