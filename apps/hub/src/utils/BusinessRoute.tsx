import { Outlet } from 'react-router-dom';
import { useEffect } from 'react';

const getTokenFromCookie = (): string | null =>
  document.cookie
    .split(';')
    .map((c) => c.trim())
    .find((c) => c.startsWith('token='))
    ?.split('=')[1] ?? null;

const getAuthUrl = (): string => {
  const isLocalhost =
    window.location.hostname === 'localhost' ||
    window.location.hostname === '127.0.0.1';
  return isLocalhost
    ? 'http://localhost:5176?redirect=hub'
    : 'https://auth.rtylr.com?redirect=hub';
};

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
