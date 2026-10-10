import React, { useState, useEffect, useMemo, useCallback } from 'react';
import { useNavigate } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { FaGoogle } from 'react-icons/fa';
import { APP_URLS, PORTAL_URL, resolvePortalReturn } from '@voxire/config';

import { useUserStore } from '../../shared/stores/userStore';
import { useLogin } from '../../shared/hooks/useAuth';
import { startGoogleLogin } from '../../shared/lib/social';
import notify from '../../shared/lib/notify/notify';
import { createOneKitClient } from '@voxire/api';
import { UserServiceClient } from '@voxire/voxie/hub/user/v1/client';
import { Button, HelmetContainer } from '@voxire/components';
import AuthShell from '../../shared/components/AuthShell';
import {
  AuthAlert,
  AuthDivider,
  AuthField,
  AuthPasswordField,
  authLinkClass,
} from '../../shared/components/AuthField';

import { getCookie, hasEffectiveAppAccess, parseEffectiveAccess } from '@voxire/auth';
import { getEmailError } from '../../shared/lib/validation';

const userClient = createOneKitClient(UserServiceClient, 'auth');

type KnownAppId = keyof typeof APP_URLS;

const Login: React.FC = (): React.JSX.Element => {
  const navigate = useNavigate();
  const { t } = useTranslation();
  const { isAuthenticated, loading, error, clearError } = useUserStore();
  const loginMutation = useLogin();

  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const emailError = getEmailError(email);

  const redirectParam = useMemo((): string | null => {
    const params = new URLSearchParams(window.location.search);
    return params.get('redirect');
  }, []);

  const redirectTo = useCallback(async (): Promise<void> => {
    try {
      const res = await userClient.getMe({});
      const provisioningStatus = res.provisioningStatus;
      // MFA is resolved before any destination is chosen, including an explicit
      // `redirect` param. Honouring that param first let a user who owes MFA
      // enrolment name another app and skip the Portal enrolment gate entirely.
      if (
        provisioningStatus?.mfaRequired &&
        (!provisioningStatus.mfaEnrolled || !provisioningStatus.mfaVerified)
      ) {
        window.location.href = PORTAL_URL;
        return;
      }
      // Honor an explicit destination only when the resolved tenant access
      // grants it. A stale or unauthorized redirect always returns to Portal.
      if (redirectParam === 'dash' || redirectParam === 'portal') {
        window.location.href = resolvePortalReturn(
          new URLSearchParams(window.location.search).get('returnTo')
        );
        return;
      }
      if (redirectParam && redirectParam in APP_URLS) {
        const hasAppAccess = hasEffectiveAppAccess(
          parseEffectiveAccess(res.effectiveAccess),
          redirectParam
        );
        if (hasAppAccess) {
          window.location.href = APP_URLS[redirectParam as KnownAppId] as string;
          return;
        }
      }
      // Portal is the home for every signed-in member. App access never chooses
      // a default destination.
      window.location.href = PORTAL_URL;
    } catch {
      // Fail closed: if MFA state cannot be read, send the user to Portal rather
      // than to a redirect target that may be behind the gate.
      window.location.href = PORTAL_URL;
    }
  }, [redirectParam]);

  const handleSubmit = (): void => {
    if (emailError) return;
    loginMutation.mutate({ email, password }, { onSuccess: () => void redirectTo() });
  };

  const handleGoogleLogin = async (): Promise<void> => {
    try {
      await startGoogleLogin(
        redirectParam,
        undefined,
        new URLSearchParams(window.location.search).get('returnTo')
      );
    } catch {
      notify.error(t('auth.common.socialUnavailable'));
    }
  };

  useEffect((): void => {
    const params = new URLSearchParams(window.location.search);
    if (params.get('error') === 'oauth_failed') {
      notify.error(t('auth.common.socialUnavailable'));
    }
  }, [t]);

  useEffect((): void => {
    clearError();
  }, [clearError]);

  useEffect((): void => {
    const hasToken = !!getCookie('token');
    if (isAuthenticated && hasToken) {
      void redirectTo();
    }
  }, [isAuthenticated, redirectTo]);

  const errorMessage = error ? t('auth.login.genericError') : null;

  const title = t('auth.login.title');
  const subtitle = t('auth.login.subtitle');

  return (
    <>
      <HelmetContainer title={t('auth.login.metaTitle')} />
      <AuthShell
        title={title}
        subtitle={subtitle}
        footer={
          <p className="text-center">
            {t('auth.login.noAccount')}{' '}
            <button
              type="button"
              className={authLinkClass}
              onClick={() =>
                navigate({
                  to: redirectParam ? `/register?redirect=${redirectParam}` : '/register',
                })
              }
            >
              {t('auth.login.createOne')}
            </button>
          </p>
        }
      >
        <form
          onSubmit={(e) => {
            e.preventDefault();
            handleSubmit();
          }}
          className="flex flex-col gap-4"
        >
          <AuthField
            id="login-email"
            label={t('auth.common.email')}
            type="email"
            placeholder={t('auth.login.emailPlaceholder')}
            autoComplete="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            autoFocus
            required
            error={emailError}
          />

          <AuthPasswordField
            id="login-password"
            label={t('auth.common.password')}
            labelAside={
              <button
                type="button"
                className="text-xs text-[#111]/55 hover:text-[#111] transition-colors"
                onClick={() => navigate({ to: '/password-reset' })}
              >
                {t('auth.login.forgotPassword')}
              </button>
            }
            placeholder={t('auth.login.passwordPlaceholder')}
            autoComplete="current-password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
          />

          {errorMessage && <AuthAlert>{errorMessage}</AuthAlert>}

          <Button
            type="submit"
            variant="primary"
            className="auth-submit-button"
            fullWidth
            loading={loading}
            height={44}
          >
            {t('auth.login.submit')}
          </Button>

          <AuthDivider label={t('auth.common.or')} />

          <Button
            type="button"
            variant="outline"
            className="auth-provider-button"
            fullWidth
            height={44}
            onClick={() => void handleGoogleLogin()}
          >
            <span className="inline-flex items-center gap-2">
              <FaGoogle className="text-[#444]" />
              {t('auth.common.continueWithGoogle')}
            </span>
          </Button>
        </form>
      </AuthShell>
    </>
  );
};

export default Login;
