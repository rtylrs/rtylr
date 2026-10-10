import React, { useState, useMemo, useEffect } from 'react';
import { Shield, Check } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import AuthShell from '../../shared/components/AuthShell';
import { HelmetContainer } from '@voxire/components';
import { useLogin } from '../../shared/hooks/useAuth';
import { useIssueAuthCode } from './useOAuth';
import { getEmailError } from '../../shared/lib/validation';
import { authInputClass } from '../../shared/components/AuthField';

function isAllowedOAuthRedirect(url: string): boolean {
  try {
    const host = new URL(url).hostname.toLowerCase();
    if (import.meta.env.DEV && (host === 'localhost' || host === '127.0.0.1')) {
      return true;
    }
    return host === 'rtylr.com' || host.endsWith('.rtylr.com');
  } catch {
    return false;
  }
}

const SCOPE_LABELS: Record<string, string> = {
  'flow:read': 'View your projects and tasks',
  'flow:write': 'Create and update projects and tasks',
  'crm:read': 'View your customers and campaigns',
  'crm:write': 'Create and update customers and campaigns',
  'finance:read': 'View your invoices and expenses',
  'finance:write': 'Create and update invoices and expenses',
  'erp:read': 'View your inventory and suppliers',
  'erp:write': 'Manage inventory and purchase orders',
};

type Step = 'login' | 'consent';

const OAuthConsent: React.FC = (): React.JSX.Element => {
  const { t } = useTranslation();
  const [step, setStep] = useState<Step>('login');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState<string | null>(null);

  const loginMutation = useLogin();
  const issueMutation = useIssueAuthCode();

  const params = useMemo(() => {
    const q = new URLSearchParams(window.location.search);
    return {
      clientId: q.get('client_id') ?? '',
      redirectUri: q.get('redirect_uri') ?? '',
      scope: q.get('scope') ?? '',
      state: q.get('state') ?? '',
      codeChallenge: q.get('code_challenge') ?? '',
      codeChallengeMethod: q.get('code_challenge_method') ?? 'S256',
    };
  }, []);

  const scopes = useMemo(() => params.scope.split(' ').filter(Boolean), [params.scope]);

  const emailError = getEmailError(email);
  const isValid = !emailError && password.length >= 8;

  const isInvalidRequest = useMemo(() => {
    if (!params.clientId || !params.redirectUri || !params.codeChallenge) {
      return 'Invalid authorization request - missing required parameters.';
    }
    // An empty scope would render a consent screen approving no permissions
    // at all - the server must be told what access is being granted.
    if (scopes.length === 0) {
      return 'Invalid authorization request - no scopes requested.';
    }
    if (!isAllowedOAuthRedirect(params.redirectUri)) {
      return 'Invalid redirect URI - must belong to an authorized domain.';
    }
    return null;
  }, [params, scopes]);

  useEffect((): void => {
    if (isInvalidRequest) {
      setError(isInvalidRequest);
    }
  }, [isInvalidRequest]);

  const handleLogin = (): void => {
    if (isInvalidRequest) return;
    setError(null);
    loginMutation.mutate(
      { email, password },
      {
        onSuccess: () => setStep('consent'),
        onError: () => setError(t('auth.login.genericError')),
      }
    );
  };

  const handleConsent = (): void => {
    setError(null);
    issueMutation.mutate(
      {
        clientId: params.clientId,
        redirectUri: params.redirectUri,
        scope: params.scope,
        state: params.state,
        codeChallenge: params.codeChallenge,
        codeChallengeMethod: params.codeChallengeMethod,
      },
      {
        onSuccess: (data) => {
          // The field is optional on the response. Without this the check
          // below would be handed undefined and, if it passed, the browser
          // would navigate to the string "undefined".
          if (data.redirectUrl === undefined || !isAllowedOAuthRedirect(data.redirectUrl)) {
            setError('Redirect URI is not on an allowed domain.');
            return;
          }
          window.location.href = data.redirectUrl;
        },
        onError: () => setError('Failed to authorize. Please try again.'),
      }
    );
  };

  const handleDeny = (): void => {
    if (!params.redirectUri || !isAllowedOAuthRedirect(params.redirectUri)) {
      setError('Redirect URI is not on an allowed domain.');
      return;
    }
    const u = new URL(params.redirectUri);
    u.searchParams.set('error', 'access_denied');
    u.searchParams.set('error_description', 'User denied access');
    if (params.state) u.searchParams.set('state', params.state);
    window.location.href = u.toString();
  };

  if (step === 'login') {
    return (
      <AuthShell
        title={t('auth.oauth.signInToAuthorize')}
        subtitle={`rtylr MCP is requesting access to your account`}
      >
        <form
          onSubmit={(e) => {
            e.preventDefault();
            handleLogin();
          }}
          className="flex flex-col gap-4"
        >
          {error && <p className="text-sm text-red-500 bg-red-50 px-3 py-2 rounded-lg">{error}</p>}

          <div className="flex flex-col gap-1.5">
            <label htmlFor="oauth-email" className="text-sm font-medium text-[#111]">
              {t('auth.common.email')}
            </label>
            <input
              id="oauth-email"
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="you@example.com"
              className={authInputClass}
              required
              autoFocus
            />
          </div>

          <div className="flex flex-col gap-1.5">
            <label htmlFor="oauth-password" className="text-sm font-medium text-[#111]">
              {t('auth.common.password')}
            </label>
            <input
              id="oauth-password"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="••••••••"
              className={authInputClass}
              required
            />
          </div>

          <button
            type="submit"
            disabled={!isValid || loginMutation.isPending || !!isInvalidRequest}
            className="w-full rounded-lg bg-[#111] py-2.5 text-sm font-semibold text-white transition-opacity hover:opacity-90 disabled:opacity-50"
          >
            {loginMutation.isPending ? 'Signing in…' : 'Continue'}
          </button>
        </form>
      </AuthShell>
    );
  }

  return (
    <>
      <HelmetContainer title={t('auth.oauth.metaTitle')} />
      <AuthShell title={t('auth.oauth.title')} subtitle={t('auth.oauth.subtitle')}>
        <div className="flex flex-col gap-6">
          {error && <p className="text-sm text-red-500 bg-red-50 px-3 py-2 rounded-lg">{error}</p>}

          <div className="rounded-xl border border-[#e5e5e5] overflow-hidden">
            <div className="flex items-center gap-3 px-4 py-3 bg-[#f9f9f9] border-b border-[#e5e5e5]">
              <Shield size={16} className="text-[#111]/70 shrink-0" />
              <span className="text-sm font-medium text-[#333]">
                {t('auth.oauth.permissionsRequested')}
              </span>
            </div>
            <ul className="divide-y divide-[#f0f0f0]">
              {scopes.map((scope) => (
                <li key={scope} className="flex items-start gap-3 px-4 py-3">
                  <Check size={14} className="text-emerald-500 mt-0.5 shrink-0" />
                  <span className="text-sm text-[#444]">{SCOPE_LABELS[scope] ?? scope}</span>
                </li>
              ))}
            </ul>
          </div>

          <div className="flex flex-col gap-2">
            <button
              type="button"
              onClick={handleConsent}
              disabled={issueMutation.isPending}
              className="w-full rounded-lg bg-[#111] py-2.5 text-sm font-semibold text-white transition-opacity hover:opacity-90 disabled:opacity-50"
            >
              {issueMutation.isPending ? t('auth.oauth.authorizing') : t('auth.oauth.allow')}
            </button>
            <button
              type="button"
              onClick={handleDeny}
              className="w-full h-11 rounded-[10px] border border-black/10 bg-white text-sm font-medium text-[#111]/70 hover:bg-black/[0.03] hover:text-[#111] transition-colors"
            >
              {t('auth.oauth.deny')}
            </button>
          </div>

          <p className="text-xs text-[#999] text-center">
            {t('auth.oauth.signedInAs')} <span className="font-medium text-[#555]">{email}</span>
          </p>
        </div>
      </AuthShell>
    </>
  );
};

export default OAuthConsent;
