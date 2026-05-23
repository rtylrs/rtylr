import React, { useState, useEffect, useMemo } from 'react';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { AlertCircle } from 'lucide-react';
import { GoogleAuthProvider, signInWithPopup, getIdToken } from 'firebase/auth';
import { FaGoogle } from 'react-icons/fa';
import { APP_URLS, APP_CONFIG } from '@voxire/config';

import { useUserStore } from '../../shared/stores/userStore';
import { useLogin, useGoogleAuth } from '../../shared/hooks/useAuth';
import { auth } from '../../firebase';
import notify from '../../shared/lib/notify/notify';
import { unwrap } from '@voxire/api';
import { safe } from '../../shared/lib/api';
import { Button } from '@voxire/components';
import AuthShell from '../../shared/components/AuthShell';

import { getEmailError } from '../../shared/lib/validation';

type KnownAppId = keyof typeof APP_URLS;

function getRedirectUrl(redirectParam: string | null): string {
  if (redirectParam && redirectParam in APP_URLS) {
    return APP_URLS[redirectParam as KnownAppId] as string;
  }
  return APP_URLS.dash as string;
}

function getAppDisplayName(redirectParam: string | null): string | null {
  if (!redirectParam) return null;
  if (redirectParam in APP_CONFIG) {
    return APP_CONFIG[redirectParam as keyof typeof APP_CONFIG].name;
  }
  return null;
}

const Login: React.FC = (): React.JSX.Element => {
  const navigate = useNavigate();
  const { t } = useTranslation();
  const provider = new GoogleAuthProvider();
  const { isAuthenticated, loading, error, clearError } = useUserStore();
  const loginMutation = useLogin();
  const googleAuthMutation = useGoogleAuth();

  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const emailError = getEmailError(email);

  const redirectParam = useMemo((): string | null => {
    const params = new URLSearchParams(window.location.search);
    return params.get('redirect');
  }, []);

  const appName = getAppDisplayName(redirectParam);

  const redirectTo = async (): Promise<void> => {
    // Explicit redirect param, honor it directly
    if (redirectParam && redirectParam in APP_URLS) {
      window.location.href = APP_URLS[redirectParam as KnownAppId] as string;
      return;
    }
    // No explicit redirect, decide based on the user's subscription apps
    try {
      const res = await unwrap(safe.get<{ role?: string; apps?: string[] }>('/user/me'));
      const role: string = res.role ?? '';
      const apps: string[] = res.apps ?? [];
      // Owners and admins always land on Dash (the account center)
      if (role === 'owner' || role === 'admin' || apps.length === 0) {
        window.location.href = APP_URLS.dash as string;
        return;
      }
      // Single-app subscriber, skip Dash and go straight to their app
      if (apps.length === 1 && apps[0] in APP_URLS) {
        window.location.href = APP_URLS[apps[0] as KnownAppId] as string;
        return;
      }
      // Multi-app, use Dash as the launcher
      window.location.href = APP_URLS.dash as string;
    } catch {
      window.location.href = APP_URLS.dash as string;
    }
  };

  const handleSubmit = (): void => {
    loginMutation.mutate({ email, password }, { onSuccess: () => void redirectTo() });
  };

  const handleGoogle = async (): Promise<void> => {
    try {
      const result = await signInWithPopup(auth, provider);
      const idToken = await getIdToken(result.user);
      googleAuthMutation.mutate({ idToken }, { onSuccess: () => void redirectTo() });
    } catch (err) {
      notify.error((err as Error).message);
    }
  };

  useEffect((): void => {
    clearError();
  }, [clearError]);

  useEffect((): void => {
    const tokenCookie = document.cookie.split(';').some((c) => c.trim().startsWith('token='));
    if (isAuthenticated && tokenCookie) {
      void redirectTo();
    }
  }, [isAuthenticated]);

  // Generic error message regardless of cause - never leak whether the
  // email exists in our system. The audit (auth.md P1 #5) flagged the
  // earlier "User not found" vs "Wrong password" split as a user-
  // enumeration vector.
  const errorMessage = error ? t('auth.login.genericError') : null;

  const title = t('auth.login.title');
  const subtitle = t('auth.login.subtitle');

  return (
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
                navigate(redirectParam ? `/register?redirect=${redirectParam}` : '/register')
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
            className="w-full h-10 rounded-lg border border-[#e5e5e5] bg-white px-3 text-sm text-[#111] placeholder:text-[#999] outline-none transition-colors focus:border-[#4583C4] focus:ring-2 focus:ring-[#4583C4]/20"
          />
          {emailError && <p className="text-xs text-red-500">{emailError}</p>}
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
              className="w-full h-10 rounded-lg border border-[#e5e5e5] bg-white px-3 pr-10 text-sm text-[#111] placeholder:text-[#999] outline-none transition-colors focus:border-[#4583C4] focus:ring-2 focus:ring-[#4583C4]/20"
            />
            <button
              type="button"
              onClick={() => setShowPassword((v) => !v)}
              aria-label={showPassword ? 'Hide password' : 'Show password'}
              className="absolute right-2.5 top-1/2 -translate-y-1/2 text-[#999] hover:text-[#666] text-xs font-medium"
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
            onClick={() => navigate('/password-reset')}
          >
            {t('auth.login.forgotPassword')}
          </button>
        </div>

        {/* Error */}
        {errorMessage && (
          <div
            className="flex items-start gap-2 rounded-lg border border-red-200 bg-red-50 px-3 py-2.5"
            role="alert"
          >
            <AlertCircle className="h-4 w-4 text-red-500 mt-0.5 shrink-0" />
            <p className="text-xs text-red-600 leading-relaxed">{errorMessage}</p>
          </div>
        )}

        {/* Sign in button */}
        <Button type="submit" variant="primary" fullWidth loading={loading} height={40}>
          {t('auth.login.submit')}
        </Button>

        {/* Google sign in */}
        <Button type="button" variant="outline" fullWidth height={40} onClick={handleGoogle}>
          <span className="inline-flex items-center gap-2">
            <FaGoogle className="text-[#444]" />
            {t('auth.common.continueWithGoogle')}
          </span>
        </Button>
      </form>
    </AuthShell>
  );
};

export default Login;
