import { Outlet } from 'react-router-dom';
import { useEffect } from 'react';
import { APP_URLS } from '@voxire/config';
import { useUserStore } from '../stores/userStore';
import api from '../lib/api';

const getTokenFromCookie = (): string | null =>
  document.cookie
    .split(';')
    .map((c) => c.trim())
    .find((c) => c.startsWith('token='))
    ?.split('=')[1] ?? null;

const getAuthUrl = (): string => `${APP_URLS.auth}?redirect=hub`;

function BusinessRoute() {
  const token = getTokenFromCookie();
  const { user, setUser, setToken } = useUserStore();

  useEffect(() => {
    if (!token) {
      window.location.href = getAuthUrl();
      return;
    }

    setToken(token);

    // Fetch user profile if not already loaded
    if (!user) {
      api.get('/user/me')
        .then((res) => {
          if (res.data) {
            setUser(res.data);
          }
        })
        .catch(() => {
          // Token is invalid or expired
        });
    }
  }, [token]);

  if (!token) {
    return null;
  }

  return <Outlet />;
}

export { BusinessRoute };
