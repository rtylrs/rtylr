import React, { useEffect, useRef } from 'react';
import { APP_URLS, APP_CONFIG } from '@voxire/config';
import { unwrap } from '@voxire/api';
import type { GoogleAuthRequest } from '@voxire/contracts/common/auth/v1/google_auth';
import { safe } from '../../shared/lib/api';
import { useUserStore } from '../../shared/stores/userStore';
import { setCookie, getDefaultCookieDomain } from '../../shared/lib/cookies';
import { storageGet, storageRemove } from '../../shared/lib/storage';

type KnownAppId = keyof typeof APP_URLS;

interface GoogleAuthResponse {
  token: string;
  role?: string;
}

function getRedirectUrl(redirectParam: string | null): string {
  if (redirectParam && redirectParam in APP_URLS) {
    return APP_URLS[redirectParam as KnownAppId] as string;
  }
  return APP_URLS.dash as string;
}

const Callback: React.FC = (): React.JSX.Element => {
  const { login } = useUserStore();
  const done = useRef(false);

  useEffect(() => {
    if (done.current) return;
    done.current = true;

    const params = new URLSearchParams(window.location.search);
    const code = params.get('code');

    if (!code) {
      window.location.replace('/?error=oauth_failed');
      return;
    }

    const redirectParam = storageGet('social_auth_redirect');
    storageRemove('social_auth_redirect');

    const request: GoogleAuthRequest = { code };

    unwrap(safe.post<GoogleAuthResponse>('/auth/register/google', request))
      .then((data) => {
        setCookie('token', data.token, {
          days: 30,
          sameSite: 'lax',
          domain: getDefaultCookieDomain(),
        });
        login(data.role ?? '');
        window.location.replace(getRedirectUrl(redirectParam));
      })
      .catch(() => {
        window.location.replace('/?error=oauth_failed');
      });
  }, [login]);

  return (
    <div
      style={{
        minHeight: '100vh',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif',
        color: '#666',
        fontSize: 14,
      }}
    >
      Signing you in...
    </div>
  );
};

export default Callback;
