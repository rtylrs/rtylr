import React, { useEffect, useMemo, useRef, useState } from 'react';
import { Trans, useTranslation } from 'react-i18next';
import { useNavigate } from '@tanstack/react-router';
import { FaGoogle } from 'react-icons/fa';
import { APP_CONFIG, APP_URLS, PORTAL_URL } from '@voxire/config';

import notify from '../../shared/lib/notify/notify';
import { useUserStore } from '../../shared/stores/userStore';
import { useRegisterRestaurant, checkEmailExists } from '../../shared/hooks/useAuth';
import { startGoogleLogin } from '../../shared/lib/social';
import { Button, HelmetContainer } from '@voxire/components';
import AuthShell from '../../shared/components/AuthShell';
import {
  AuthAlert,
  AuthDivider,
  AuthField,
  AuthPasswordField,
  authLinkClass,
} from '../../shared/components/AuthField';
import PasswordStrengthMeter, {
  scorePassword,
} from '../../shared/components/PasswordStrengthMeter';
import { getCookie } from '@voxire/auth';
import { getEmailError, getPasswordMatchError } from '../../shared/lib/validation';

const DEFAULT_SIGNUP_PLAN_SLUG = 'free';

type RedirectableAppId = keyof typeof APP_CONFIG;

function getRedirectAppId(value: string | null): RedirectableAppId | null {
  if (value && Object.prototype.hasOwnProperty.call(APP_CONFIG, value)) {
    return value as RedirectableAppId;
  }
  return null;
}

function getRedirectUrl(appId: RedirectableAppId | null): string {
  return appId ? (APP_URLS[appId] as string) : PORTAL_URL;
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
  const [emailTaken, setEmailTaken] = useState(false);
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

  const redirectParam = useMemo((): RedirectableAppId | null => {
    const params = new URLSearchParams(window.location.search);
    return getRedirectAppId(params.get('redirect'));
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

  const handleGoogleLogin = async (): Promise<void> => {
    if (!businessName.trim()) {
      notify.error('Enter your business name first');
      return;
    }
    try {
      await startGoogleLogin(redirectParam, businessName.trim());
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
          <p className="text-center">
            {t('auth.register.haveAccount')}{' '}
            <button
              type="button"
              className={authLinkClass}
              onClick={() => navigate({ to: redirectParam ? `/?redirect=${redirectParam}` : '/' })}
            >
              {t('auth.common.signIn')}
            </button>
          </p>
        }
      >
        <form
          onSubmit={(e) => {
            e.preventDefault();
            handleRegister();
          }}
          className="flex flex-col gap-4"
        >
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <AuthField
              id="register-name"
              label={t('auth.common.fullName')}
              type="text"
              placeholder={t('auth.register.namePlaceholder')}
              autoComplete="name"
              value={name}
              onChange={(e) => setName(e.target.value)}
            />
            <AuthField
              id="register-business"
              label={t('auth.common.businessName')}
              type="text"
              placeholder={t('auth.register.businessPlaceholder')}
              value={businessName}
              onChange={(e) => setBusinessName(e.target.value)}
              required
            />
          </div>

          <AuthField
            id="register-email"
            label={t('auth.common.email')}
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
            error={emailError || (emailTaken ? t('auth.register.emailTaken') : null)}
          />

          <div className="flex flex-col gap-2">
            <AuthPasswordField
              id="register-password"
              label={t('auth.common.password')}
              placeholder={t('auth.common.atLeastEight')}
              autoComplete="new-password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
            />
            {password && <PasswordStrengthMeter password={password} />}
          </div>

          <AuthPasswordField
            id="register-confirm"
            label={t('auth.common.confirmPassword')}
            placeholder={t('auth.common.repeatPassword')}
            autoComplete="new-password"
            value={confirmPassword}
            onChange={(e) => setConfirmPassword(e.target.value)}
            required
            error={matchError}
          />

          {errorMessage && <AuthAlert>{errorMessage}</AuthAlert>}

          <p className="text-xs leading-relaxed text-[#111]/50">
            <Trans
              i18nKey="auth.register.legalNotice"
              components={{
                terms: (
                  <a
                    href="https://rtylr.com/legal/terms/"
                    target="_blank"
                    rel="noopener noreferrer"
                    className={authLinkClass}
                  />
                ),
                privacy: (
                  <a
                    href="https://rtylr.com/legal/privacy-policy/"
                    target="_blank"
                    rel="noopener noreferrer"
                    className={authLinkClass}
                  />
                ),
              }}
            />
          </p>

          <Button
            type="submit"
            variant="primary"
            className="auth-submit-button"
            fullWidth
            loading={loading}
            height={44}
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

export default Register;
