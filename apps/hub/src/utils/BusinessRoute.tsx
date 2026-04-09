import { Outlet } from 'react-router-dom';
import { useEffect } from 'react';
import { APP_URLS } from '@voxire/config';

const getTokenFromCookie = (): string | null =>
  document.cookie
    .split(';')
    .map((c) => c.trim())
    .find((c) => c.startsWith('token='))
    ?.split('=')[1] ?? null;

const getAuthUrl = (): string => `${APP_URLS.auth}?redirect=hub`;

function BusinessRoute() {
  const token = getTokenFromCookie();

  useEffect(() => {
    if (!token) {
      window.location.href = getAuthUrl();
    }
  }, [token]);

  if (!token) {
    return null;
  }

  return <Outlet />;
}

export { BusinessRoute };
