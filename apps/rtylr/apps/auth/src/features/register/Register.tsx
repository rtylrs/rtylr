import React, { useEffect, useMemo, useRef, useState } from 'react';
import { Trans, useTranslation } from 'react-i18next';
import { useNavigate } from '@tanstack/react-router';
import { AlertCircle } from 'lucide-react';
import { FaGoogle, FaGithub, FaMicrosoft, FaApple } from 'react-icons/fa';
import { APP_URLS } from '@voxire/config';

import notify from '../../shared/lib/notify/notify';
import { useUserStore } from '../../shared/stores/userStore';
import { useRegisterRestaurant, checkEmailExists } from '../../shared/hooks/useAuth';
import { useFeatureGate } from '../../shared/hooks/useFeatureGate';
import { startSocialLogin, type SocialProvider } from '../../shared/lib/social';
import { Button, HelmetContainer } from '@voxire/components';
import AuthShell from '../../shared/components/AuthShell';
import PasswordStrengthMeter, {
  scorePassword,
} from '../../shared/components/PasswordStrengthMeter';
import { getCookie } from '@voxire/auth';
import { getEmailError, getPasswordMatchError } from '../../shared/lib/validation';

const DEFAULT_SIGNUP_PLAN_SLUG = 'free';
const AUTH_SOCIAL_LOGIN_FLAG = 'auth.social_login_enabled';

type KnownAppId = keyof typeof APP_URLS;

function getRedirectUrl(redirectParam: string | null): string {
  if (redirectParam && redirectParam in APP_URLS) {
    return APP_URLS[redirectParam as KnownAppId] as string;
  }
  return APP_URLS.dash as string;
}

// ── Register ────────────────────────────────────────────────────────────────

const Register: React.FC = (): React.JSX.Element => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [name, setName] = useState('');
  const [businessName, setBusinessName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirm, setShowConfirm] = useState(false);
  const [emailTaken, setEmailTaken] = useState(false);
  const showSocialLogin = useFeatureGate(AUTH_SOCIAL_LOGIN_FLAG, false);
  const emailError = getEmailError(email);
  const matchError = getPasswordMatchError(password, confirmPassword);

  // Check email availability once the field is filled and well-formed, so the
  // user sees "email already exists" before submitting. A sequence counter
  // guards against out-of-order responses: only the newest lookup may update
  // `emailTaken`, otherwise a slow response for a previous email could
  // wrongly disable submission.
  const emailCheckSeq = useRef(0);
  const checkEmailAvailable = async (): Promise<void> => {
    if (getEmailError(email)) return;
    const seq = ++emailCheckSeq.current;
    try {
      const { exists } = await checkEmailExists(email);
      // Optional on the response. An answer that did not arrive is not a
      // reason to tell the user their email is taken.
      if (seq === emailCheckSeq.current) setEmailTaken(exists ?? false);
    } catch {
      // never block signup on a lookup failure
      if (seq === emailCheckSeq.current) setEmailTaken(false);
    }
  };
  const pwScore = scorePassword(password);
  const { isAuthenticated, loading, error, clearError } = useUserStore();
  const registerMutation = useRegisterRestaurant();

  const redirectParam = useMemo((): string | null => {
    const params = new URLSearchParams(window.location.search);
    return params.get('redirect');
  }, []);

  const handleRegister = (): void => {
    if (password !== confirmPassword) {
      notify.error('Passwords do not match');
      return;
    }
    registerMutation.mutate(
      {
        email,
        password,
        name: name.trim() || email.split('@')[0],
        businessName: businessName.trim(),
        planSlug: DEFAULT_SIGNUP_PLAN_SLUG,
      },
      {
        onSuccess: () => {
          window.location.href = getRedirectUrl(redirectParam);
        },
      }
    );
  };

  const handleSocial = async (provider: SocialProvider): Promise<void> => {
    if (!businessName.trim()) {
      notify.error('Enter your business name first');
      return;
    }
    try {
      await startSocialLogin(provider, redirectParam, businessName.trim());
    } catch {
      notify.error('Social sign-in is currently unavailable');
    }
  };

  useEffect((): void => {
    clearError();
  }, [clearError]);

  useEffect((): void => {
    if (isAuthenticated && getCookie('token')) {
      window.location.href = getRedirectUrl(redirectParam);
    }
  }, [isAuthenticated, redirectParam]);

  const errorMessage =
    error &&
    (typeof error === 'string'
      ? error
      : (error as { message?: string })?.message || 'An error occurred');

  return (
    <>
      <HelmetContainer title={t('auth.register.metaTitle')} />
      <AuthShell
        title={t('auth.register.heading')}
        subtitle={t('auth.register.planNotice')}
        footer={
          <div className="text-center">
            <p className="text-sm text-[#666]">
              Already have an account?{' '}
              <button
                type="button"
                className="text-[#4583C4] font-medium hover:underline underline-offset-2"
                onClick={() =>
                  navigate({ to: redirectParam ? `/?redirect=${redirectParam}` : '/' })
                }
              >
                {t('auth.common.signIn')}
              </button>
            </p>
          </div>
        }
      >
        <form
          onSubmit={(e) => {
            e.preventDefault();
            handleRegister();
          }}
          className="flex flex-col gap-4"
        >
          {/* Full name */}
          <div className="flex flex-col gap-1.5">
            <label htmlFor="register-name" className="text-sm font-medium text-[#111]">
              {t('auth.common.fullName')}
            </label>
            <input
              id="register-name"
              type="text"
              placeholder={t('auth.register.namePlaceholder')}
              autoComplete="name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="w-full h-10 rounded-lg border border-[#e5e5e5] bg-white px-3 text-sm text-[#111] placeholder:text-[#999] outline-none transition-colors focus:border-[#4583C4] focus:ring-2 focus:ring-[#4583C4]/20"
            />
          </div>

          {/* Business name */}
          <div className="flex flex-col gap-1.5">
            <label htmlFor="register-business" className="text-sm font-medium text-[#111]">
              {t('auth.common.businessName')}
            </label>
            <input
              id="register-business"
              type="text"
              placeholder={t('auth.register.businessPlaceholder')}
              value={businessName}
              onChange={(e) => setBusinessName(e.target.value)}
              required
              className="w-full h-10 rounded-lg border border-[#e5e5e5] bg-white px-3 text-sm text-[#111] placeholder:text-[#999] outline-none transition-colors focus:border-[#4583C4] focus:ring-2 focus:ring-[#4583C4]/20"
            />
          </div>

          {/* Email */}
          <div className="flex flex-col gap-1.5">
            <label htmlFor="register-email" className="text-sm font-medium text-[#111]">
              {t('auth.common.email')}
            </label>
            <input
              id="register-email"
              type="email"
              placeholder="you@company.com"
              autoComplete="email"
              value={email}
              onChange={(e) => {
                setEmail(e.target.value);
                if (emailTaken) setEmailTaken(false);
              }}
              onBlur={() => void checkEmailAvailable()}
              required
              className="w-full h-10 rounded-lg border border-[#e5e5e5] bg-white px-3 text-sm text-[#111] placeholder:text-[#999] outline-none transition-colors focus:border-[#4583C4] focus:ring-2 focus:ring-[#4583C4]/20"
            />
            {emailError && <p className="text-xs text-red-500">{emailError}</p>}
            {!emailError && emailTaken && (
              <p className="text-xs text-red-500">{t('auth.register.emailTaken')}</p>
            )}
          </div>

          {/* Password */}
          <div className="flex flex-col gap-1.5">
            <label htmlFor="register-password" className="text-sm font-medium text-[#111]">
              {t('auth.common.password')}
            </label>
            <div className="relative">
              <input
                id="register-password"
                type={showPassword ? 'text' : 'password'}
                placeholder={t('auth.common.atLeastEight')}
                autoComplete="new-password"
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
            {password && <PasswordStrengthMeter password={password} />}
          </div>

          {/* Confirm password */}
          <div className="flex flex-col gap-1.5">
            <label htmlFor="register-confirm" className="text-sm font-medium text-[#111]">
              {t('auth.common.confirmPassword')}
            </label>
            <div className="relative">
              <input
                id="register-confirm"
                type={showConfirm ? 'text' : 'password'}
                placeholder={t('auth.common.repeatPassword')}
                autoComplete="new-password"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                required
                className="w-full h-10 rounded-lg border border-[#e5e5e5] bg-white px-3 pr-10 text-sm text-[#111] placeholder:text-[#999] outline-none transition-colors focus:border-[#4583C4] focus:ring-2 focus:ring-[#4583C4]/20"
              />
              <button
                type="button"
                onClick={() => setShowConfirm((v) => !v)}
                aria-label={showConfirm ? 'Hide password' : 'Show password'}
                className="absolute right-2.5 top-1/2 -translate-y-1/2 text-[#999] hover:text-[#666] text-xs font-medium"
              >
                {showConfirm ? 'Hide' : 'Show'}
              </button>
            </div>
            {matchError && <p className="text-xs text-red-500">{matchError}</p>}
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

          <p className="text-xs leading-relaxed text-[#666]">
            <Trans
              i18nKey="auth.register.legalNotice"
              components={{
                terms: (
                  <a
                    href="https://rtylr.com/legal/terms/"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-[#4583C4] underline underline-offset-2"
                  />
                ),
                privacy: (
                  <a
                    href="https://rtylr.com/legal/privacy-policy/"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-[#4583C4] underline underline-offset-2"
                  />
                ),
              }}
            />
          </p>

          {/* Submit */}
          <Button
            type="submit"
            variant="primary"
            className="auth-submit-button"
            fullWidth
            loading={loading}
            height={40}
            disabled={
              !email ||
              !!emailError ||
              emailTaken ||
              password.length < 8 ||
              pwScore < 2 ||
              !!matchError ||
              !businessName.trim()
            }
          >
            {t('auth.register.submit')}
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

export default Register;
