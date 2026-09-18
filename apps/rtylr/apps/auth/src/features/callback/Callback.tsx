import React, { useEffect, useRef } from 'react';
import { useTranslation } from 'react-i18next';
import { APP_URLS } from '@voxire/config';
import type { GoogleAuthRequest } from '@voxire/voxie/common/auth/v1/types';
import { useGoogleAuth } from '../../shared/hooks/useAuth';
import { sessionStorageGet, sessionStorageRemove } from '../../shared/lib/storage';
import { SOCIAL_AUTH_BUSINESS_NAME_KEY, SOCIAL_AUTH_REDIRECT_KEY } from '../../shared/lib/social';

type KnownAppId = keyof typeof APP_URLS;

function getRedirectUrl(redirectParam: string | null): string {
  if (redirectParam && redirectParam in APP_URLS) {
    return APP_URLS[redirectParam as KnownAppId] as string;
  }
  return APP_URLS.dash as string;
}

const Callback: React.FC = (): React.JSX.Element => {
  const { t } = useTranslation();
  const googleAuthMutation = useGoogleAuth();
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

    const redirectParam = sessionStorageGet(SOCIAL_AUTH_REDIRECT_KEY);
    const businessName = sessionStorageGet(SOCIAL_AUTH_BUSINESS_NAME_KEY);
    sessionStorageRemove(SOCIAL_AUTH_REDIRECT_KEY);
    sessionStorageRemove(SOCIAL_AUTH_BUSINESS_NAME_KEY);

    const request: GoogleAuthRequest = {
      code,
      businessName: businessName || undefined,
    };

    googleAuthMutation.mutate(request, {
      onSuccess: () => window.location.replace(getRedirectUrl(redirectParam)),
      onError: () => window.location.replace('/?error=oauth_failed'),
    });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

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
      {t('auth.callback.signingIn')}
    </div>
  );
};

export default Callback;
