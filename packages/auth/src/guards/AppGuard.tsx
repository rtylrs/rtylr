import React, { useCallback, useEffect, useRef, useState } from 'react';
import { APP_URLS } from '@voxire/config';
import { getCookie } from '../cookies';
import { signOut } from '../signOut';
import { useUserStore } from '../store';
import type { AppId, User } from '../types';

type FetchUserFn = () => Promise<Record<string, unknown>>;

interface AppGuardProps {
  appId: AppId;
  fetchUser: FetchUserFn;
  children: React.ReactNode;
  loader?: React.ReactNode;
  denied?: React.ReactNode;
}

function DefaultLoader(): React.JSX.Element {
  return (
    <div
      style={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        height: '100vh',
        background: '#fff',
      }}
    >
      <div
        style={{
          width: 24,
          height: 24,
          border: '2px solid #e5e5e5',
          borderTopColor: '#4583C4',
          borderRadius: '50%',
          animation: 'spin 0.6s linear infinite',
        }}
      />
      <style>{'@keyframes spin { to { transform: rotate(360deg) } }'}</style>
    </div>
  );
}

function DefaultDenied({ appId }: { appId: AppId }): React.JSX.Element {
  return (
    <div
      style={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        height: '100vh',
        background: '#fff',
        color: '#111',
        gap: '12px',
        fontSize: '14px',
      }}
    >
      <p style={{ fontWeight: 600 }}>You don't have access to this app.</p>
      <p style={{ color: '#999', fontSize: '13px' }}>Contact your admin to get access.</p>
      <a href={APP_URLS.dash} style={{ color: '#4583C4', fontSize: '13px' }}>
        Go to Dash
      </a>
      <button
        type="button"
        onClick={() => signOut(appId)}
        style={{
          fontSize: '12px',
          color: '#999',
          background: 'none',
          border: 'none',
          cursor: 'pointer',
          textDecoration: 'underline',
          textUnderlineOffset: '2px',
        }}
      >
        Sign in with a different account
      </button>
    </div>
  );
}

/**
 * AppGuard wraps an app's root and enforces:
 * 1. A valid token cookie (redirects to auth if missing)
 * 2. The authenticated user has the requested app in their `apps` array
 *    (owners and admins are always allowed)
 *
 * It also hydrates the shared `useUserStore` from the `/user/me` response.
 *
 * Usage:
 * ```tsx
 * <AppGuard appId="flow" fetchUser={() => api.get('/user/me').then(r => r.data)}>
 *   <AppRouter />
 * </AppGuard>
 * ```
 */
function DefaultError({ onRetry }: { onRetry: () => void }): React.JSX.Element {
  return (
    <div
      style={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        height: '100vh',
        background: '#fff',
        color: '#111',
        gap: '12px',
        fontSize: '14px',
      }}
    >
      <p style={{ fontWeight: 600 }}>Could not verify your session.</p>
      <p style={{ color: '#999', fontSize: '13px' }}>Check your connection and try again.</p>
      <button
        type="button"
        onClick={onRetry}
        style={{
          marginTop: '8px',
          padding: '8px 16px',
          borderRadius: '8px',
          border: '1px solid #e5e5e5',
          background: '#fff',
          cursor: 'pointer',
          fontSize: '13px',
        }}
      >
        Retry
      </button>
    </div>
  );
}

export function AppGuard({
  appId,
  fetchUser,
  children,
  loader,
  denied,
}: AppGuardProps): React.JSX.Element {
  const [status, setStatus] = useState<'loading' | 'allowed' | 'denied' | 'error'>('loading');
  const [attempt, setAttempt] = useState(0);
  const { login } = useUserStore();
  const fetchUserRef = useRef(fetchUser);
  fetchUserRef.current = fetchUser;

  const verifySession = useCallback((): void => {
    const token = getCookie('token');
    if (!token) {
      window.location.href = `${APP_URLS.auth}?redirect=${appId}`;
      return;
    }

    setStatus('loading');

    fetchUserRef.current()
      .then((data) => {
        const user: User = {
          id: String(data.id ?? data.user_id ?? data.userId ?? ''),
          email: (data.email as string) ?? '',
          name: (data.fullName as string) ?? (data.full_name as string) ?? '',
          role: (data.role as User['role']) ?? undefined,
          apps: (data.apps as AppId[]) ?? [],
          planType: (data.planSlug as string) ?? (data.plan_slug as string) ?? undefined,
        };
        login(user);

        const apps: string[] = (data.apps as string[]) ?? [];
        const role: string = (data.role as string) ?? '';

        const allowed =
          role === 'owner' || role === 'admin' || apps.length === 0 || apps.includes(appId);
        setStatus(allowed ? 'allowed' : 'denied');
      })
      .catch((): void => {
        setStatus('error');
      });
  }, [appId, login]);

  useEffect((): void => {
    verifySession();
  }, [verifySession, attempt]);

  if (status === 'loading') return loader ? <>{loader}</> : <DefaultLoader />;
  if (status === 'denied') return denied ? <>{denied}</> : <DefaultDenied appId={appId} />;
  if (status === 'error') {
    return <DefaultError onRetry={() => setAttempt((n) => n + 1)} />;
  }
  return <>{children}</>;
}
