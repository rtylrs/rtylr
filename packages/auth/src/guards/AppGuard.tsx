import React, { useCallback, useEffect, useRef, useState } from 'react';
import { resolveAppUrl, resolveAuthUrl } from '@voxire/config';
import { Sentry } from '@voxire/observability';
import { getCookie } from '../cookies';
import { signOut } from '../signOut';
import { useUserStore } from '../store';
import { parseUserRole } from '../types';
import type { AppId, User } from '../types';

type FetchUserFn = () => Promise<Record<string, unknown>>;

interface AppGuardProps {
  appId: AppId;
  fetchUser: FetchUserFn;
  children: React.ReactNode;
  loader?: React.ReactNode;
  denied?: React.ReactNode;
  error?: (retry: () => void) => React.ReactNode;
}

function DefaultLoader(): React.JSX.Element {
  return (
    <div
      role="status"
      aria-label="Loading"
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
      <a href={resolveAppUrl('dash')} style={{ color: '#4583C4', fontSize: '13px' }}>
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
 * The loading and denied states accept a `React.ReactNode` override; the
 * error state accepts a render prop receiving the retry callback so apps can
 * translate or replace the default session-error UI (#42).
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
  error,
}: AppGuardProps): React.JSX.Element {
  const [status, setStatus] = useState<'loading' | 'allowed' | 'denied' | 'error'>('loading');
  const [attempt, setAttempt] = useState(0);
  const { login } = useUserStore();
  const fetchUserRef = useRef(fetchUser);
  fetchUserRef.current = fetchUser;

  const verifySession = useCallback((): void => {
    const token = getCookie('token');
    if (!token) {
      window.location.href = `${resolveAuthUrl()}?redirect=${appId}`;
      return;
    }

    setStatus('loading');

    fetchUserRef
      .current()
      .then((data) => {
        // The @voxire/api response interceptor already converts wire
        // snake_case keys to camelCase, so only camelCase fields exist here.
        const user: User = {
          id: String(data.id ?? data.userId ?? ''),
          email: (data.email as string) ?? '',
          name: (data.fullName as string) ?? '',
          role: parseUserRole(data.role) ?? undefined,
          apps: (data.apps as AppId[]) ?? [],
          planType: (data.planSlug as string) ?? undefined,
          isConsoleAdmin: data.isConsoleAdmin === true || undefined,
        };
        login(user);

        const apps: string[] = (data.apps as string[]) ?? [];
        const role = parseUserRole(data.role);

        // A role the frontend cannot interpret fails closed, and does so even
        // when the apps array would otherwise admit the user. This used to be an
        // unchecked cast, so an unrecognised backend role became a valid
        // UserRole with nothing complaining, and access was decided by a value
        // no code here understood (#795).
        //
        // "Absent" and "unrecognised" are deliberately different. A response
        // with no role at all is the pre-existing case and still goes through
        // the apps array; only a present-but-unknown value denies.
        const roleWasSent = data.role !== undefined && data.role !== null && data.role !== '';
        if (roleWasSent && role === null) {
          // Report before denying. A role the frontend cannot interpret means
          // the backend knows about one this build does not, and without this
          // the only symptom is a user seeing an access-denied screen they
          // cannot explain (#795).
          //
          // The value is a role name, not user data, so it is safe to send.
          // captureMessage is a no-op when Sentry has not initialised, which is
          // why this does not need to know whether boot has run.
          Sentry.captureMessage('AppGuard: unrecognised user role', {
            level: 'error',
            extra: { appId, role: String(data.role) },
          });
          setStatus('denied');
          return;
        }

        const allowed =
          role === 'owner' || role === 'admin' || (apps.length > 0 && apps.includes(appId));
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
    const retry = (): void => setAttempt((n) => n + 1);
    if (error) return <>{error(retry)}</>;
    return <DefaultError onRetry={retry} />;
  }
  return <>{children}</>;
}
