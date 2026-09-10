import React, { useState, useEffect, useMemo, useCallback } from 'react';
import { useNavigate } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { AlertCircle } from 'lucide-react';
import { FaGoogle, FaGithub, FaMicrosoft, FaApple } from 'react-icons/fa';
import { APP_URLS } from '@voxire/config';

import { useUserStore } from '../../shared/stores/userStore';
import { useLogin } from '../../shared/hooks/useAuth';
import { useFeatureGate } from '../../shared/hooks/useFeatureGate';
import { startSocialLogin, type SocialProvider } from '../../shared/lib/social';
import notify from '../../shared/lib/notify/notify';
import { createOneKitClient } from '@voxire/api';
import { UserServiceClient } from '@voxire/voxie/hub/user/v1/client';
import { licenseClient } from '../../shared/lib/clients';
import { Button, HelmetContainer, LicenseBanner } from '@voxire/components';
import AuthShell from '../../shared/components/AuthShell';

import { getCookie } from '@voxire/auth';
import { getEmailError } from '../../shared/lib/validation';

const userClient = createOneKitClient(UserServiceClient, 'auth');

type KnownAppId = keyof typeof APP_URLS;

const AUTH_SOCIAL_LOGIN_FLAG = 'auth.social_login_enabled';

const Login: React.FC = (): React.JSX.Element => {
  const navigate = useNavigate();
  const { t } = useTranslation();
  const { isAuthenticated, loading, error, clearError } = useUserStore();
  const loginMutation = useLogin();

  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const showSocialLogin = useFeatureGate(AUTH_SOCIAL_LOGIN_FLAG, false);
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
      // enrolment name another app and skip the Dash enrolment gate entirely.
      if (
        provisioningStatus?.mfaRequired &&
        (!provisioningStatus.mfaEnrolled || !provisioningStatus.mfaVerified)
      ) {
        window.location.href = APP_URLS.dash as string;
        return;
      }
      // Explicit redirect param, honoured once MFA is satisfied.
      if (redirectParam && redirectParam in APP_URLS) {
        window.location.href = APP_URLS[redirectParam as KnownAppId] as string;
        return;
      }
      // No explicit redirect, decide based on the user's subscription apps
      const role = res.role;
      // Both branches are optional on the response, and the reads below are
      // `apps.length` and `apps[0]`, so an absent list would throw here and
      // strand the user on the sign-in screen after a successful sign-in.
      // No apps is a real state, and it means Dash, which is what the
      // `apps.length === 0` branch already does.
      const apps =
        (res.effectiveAccess
          ? res.effectiveAccess.apps
              ?.filter((app) => app.licensed && app.hasAccess)
              .map((app) => app.appId)
          : res.apps) ?? [];
      const hasDashManagement = res.effectiveAccess
        ? res.effectiveAccess.isOwner || res.effectiveAccess.isAdmin
        : role === 'owner' || role === 'admin';
      // Owners and admins always land on Dash, the account center.
      if (hasDashManagement || apps.length === 0) {
        window.location.href = APP_URLS.dash as string;
        return;
      }
      // Single-app subscriber, skip Dash and go straight to their app
      const only = apps.length === 1 ? apps[0] : undefined;
      if (only !== undefined && only in APP_URLS) {
        window.location.href = APP_URLS[only as KnownAppId] as string;
        return;
      }
      // Multi-app, use Dash as the launcher
      window.location.href = APP_URLS.dash as string;
    } catch {
      // Fail closed: if MFA state cannot be read, send the user to Dash rather
      // than to a redirect target that may be behind the gate.
      window.location.href = APP_URLS.dash as string;
    }
  }, [redirectParam]);

  const handleSubmit = (): void => {
    if (emailError) return;
    loginMutation.mutate({ email, password }, { onSuccess: () => void redirectTo() });
  };

  const handleSocial = async (provider: SocialProvider): Promise<void> => {
    try {
      await startSocialLogin(provider, redirectParam);
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
      <LicenseBanner client={licenseClient} />
      <AuthShell
        title={title}
        subtitle={subtitle}
        footer={
          <div className="text-center space-y-3">
            <p className="text-sm text-[#666]">
              {t('auth.login.noAccount')}{' '}
              <button
                type="button"
                className="text-[#4583C4] font-medium hover:underline underline-offset-2"
                onClick={() =>
                  navigate({
                    to: redirectParam ? `/register?redirect=${redirectParam}` : '/register',
                  })
                }
              >
                {t('auth.login.createOne')}
              </button>
            </p>
          </div>
        }
      >
        <form
          onSubmit={(e) => {
            e.preventDefault();
            handleSubmit();
          }}
          className="flex flex-col gap-4"
        >
          {/* Email */}
          <div className="flex flex-col gap-1.5">
            <label htmlFor="login-email" className="text-sm font-medium text-[#111]">
              {t('auth.common.email')}
            </label>
            <input
              id="login-email"
              type="email"
              placeholder={t('auth.login.emailPlaceholder')}
              autoComplete="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              autoFocus
              required
              aria-invalid={!!emailError}
              aria-describedby={emailError ? 'login-email-error' : undefined}
              className="w-full h-10 rounded-lg border border-[#e5e5e5] bg-white px-3 text-base md:text-sm text-[#111] placeholder:text-[#999] outline-none transition-colors focus:border-[#4583C4] focus:ring-2 focus:ring-[#4583C4]/20"
            />
            {emailError && (
              <p id="login-email-error" className="text-xs text-red-500" role="alert">
                {emailError}
              </p>
            )}
          </div>

          {/* Password */}
          <div className="flex flex-col gap-1.5">
            <label htmlFor="login-password" className="text-sm font-medium text-[#111]">
              {t('auth.common.password')}
            </label>
            <div className="relative">
              <input
                id="login-password"
                type={showPassword ? 'text' : 'password'}
                placeholder={t('auth.login.passwordPlaceholder')}
                autoComplete="current-password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
                className="w-full h-10 rounded-lg border border-[#e5e5e5] bg-white px-3 pr-10 text-base md:text-sm text-[#111] placeholder:text-[#999] outline-none transition-colors focus:border-[#4583C4] focus:ring-2 focus:ring-[#4583C4]/20"
              />
              <button
                type="button"
                onClick={() => setShowPassword((v) => !v)}
                aria-label={showPassword ? 'Hide password' : 'Show password'}
                className="absolute right-2.5 top-1/2 -translate-y-1/2 text-[#999] hover:text-[#666] text-xs font-medium min-h-[36px] px-1 flex items-center"
              >
                {showPassword ? 'Hide' : 'Show'}
              </button>
            </div>
          </div>

          {/* Forgot password */}
          <div className="flex justify-end -mt-1">
            <button
              type="button"
              className="text-xs text-[#4583C4] hover:underline underline-offset-2"
              onClick={() => navigate({ to: '/password-reset' })}
            >
              {t('auth.login.forgotPassword')}
            </button>
          </div>

          {/* Error */}
          {errorMessage && (
            <div
              className="flex items-start gap-2 rounded-lg border border-red-200 bg-red-50 px-3 py-2.5"
              role="alert"
              aria-live="polite"
            >
              <AlertCircle className="h-4 w-4 text-red-500 mt-0.5 shrink-0" />
              <p className="text-xs text-red-600 leading-relaxed">{errorMessage}</p>
            </div>
          )}

          {/* Sign in button */}
          <Button
            type="submit"
            variant="primary"
            className="auth-submit-button"
            fullWidth
            loading={loading}
            height={40}
          >
            {t('auth.login.submit')}
          </Button>

          {showSocialLogin && (
            <div className="grid grid-cols-2 gap-2">
              <Button
                type="button"
                variant="outline"
                className="auth-provider-button"
                fullWidth
                height={40}
                onClick={() => handleSocial('google')}
              >
                <span className="inline-flex items-center gap-2">
                  <FaGoogle className="text-[#444]" />
                  {t('auth.common.google')}
                </span>
              </Button>
              <Button
                type="button"
                variant="outline"
                className="auth-provider-button"
                fullWidth
                height={40}
                onClick={() => handleSocial('github')}
              >
                <span className="inline-flex items-center gap-2">
                  <FaGithub className="text-[#444]" />
                  {t('auth.common.github')}
                </span>
              </Button>
              <Button
                type="button"
                variant="outline"
                className="auth-provider-button"
                fullWidth
                height={40}
                onClick={() => handleSocial('microsoft')}
              >
                <span className="inline-flex items-center gap-2">
                  <FaMicrosoft className="text-[#444]" />
                  {t('auth.common.microsoft')}
                </span>
              </Button>
              <Button
                type="button"
                variant="outline"
                className="auth-provider-button"
                fullWidth
                height={40}
                onClick={() => handleSocial('apple')}
              >
                <span className="inline-flex items-center gap-2">
                  <FaApple className="text-[#444]" />
                  {t('auth.common.apple')}
                </span>
              </Button>
            </div>
          )}
        </form>
      </AuthShell>
    </>
  );
};

export default Login;
