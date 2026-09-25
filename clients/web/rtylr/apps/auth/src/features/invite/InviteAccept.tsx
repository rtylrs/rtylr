import React, { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { useParams, useNavigate } from '@tanstack/react-router';
import { AlertCircle } from 'lucide-react';
import { Button, HelmetContainer } from '@voxire/components';
import { APP_URLS } from '@voxire/config';
import AuthShell from '../../shared/components/AuthShell';
import { createOneKitClient, getOneKitErrorMessage } from '@voxire/api';
import { InvitationsServiceClient } from '@voxire/voxie/hub/team/invitations/v1/client';
import { AuthServiceClient } from '@voxire/voxie/common/auth/v1/client';
import type { InviteAcceptRequest } from '@voxire/voxie/common/auth/v1/types';
import type { AppId } from '@voxire/types';
import { isStrongPassword } from '../../shared/lib/validation';

const invitationsClient = createOneKitClient(InvitationsServiceClient, 'auth');
const authClient = createOneKitClient(AuthServiceClient, 'auth');

const InviteAccept: React.FC = () => {
  const { t } = useTranslation();
  const { token } = useParams({ strict: false });
  const navigate = useNavigate();

  const [email, setEmail] = useState('');
  const [fullName, setFullName] = useState('');
  const [apps, setApps] = useState<string[]>([]);
  const [valid, setValid] = useState<boolean | null>(null);
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);

  // Verify the token on mount
  useEffect(() => {
    if (!token) return;
    invitationsClient
      .verifyInvite({ token })
      .then((res) => {
        if (res.valid) {
          setValid(true);
          setEmail(res.email || '');
          setFullName(res.fullName || '');
          setApps(res.apps ?? []);
        } else {
          setValid(false);
        }
      })
      .catch(() => setValid(false));
  }, [token]);

  const handleSubmit = async () => {
    if (!isStrongPassword(password)) {
      setError(
        'Password must be at least 8 characters and include uppercase, lowercase, and numbers'
      );
      return;
    }
    if (password !== confirmPassword) {
      setError('Passwords do not match');
      return;
    }

    setLoading(true);
    setError(null);

    try {
      const request: InviteAcceptRequest = { token: token ?? '', password };
      await authClient.inviteAccept(request);
      setSuccess(true);
    } catch (err) {
      setError(getOneKitErrorMessage(err, 'Failed to set password'));
    } finally {
      setLoading(false);
    }
  };

  // Loading state
  if (valid === null) {
    return (
      <AuthShell title={t('auth.invite.verifying')} subtitle={t('auth.invite.pleaseWait')}>
        <div className="flex justify-center py-8">
          <div className="h-6 w-6 border-2 border-[#4583C4] border-t-transparent rounded-full animate-spin" />
        </div>
      </AuthShell>
    );
  }

  // Invalid/expired token
  if (valid === false) {
    return (
      <AuthShell
        title={t('auth.invite.invalidTitle')}
        subtitle={t('auth.invite.invalidBody')}
        footer={
          <p className="text-sm text-[#666]">
            <button
              type="button"
              className="text-[#4583C4] font-medium hover:underline underline-offset-2"
              onClick={() => navigate({ to: '/' })}
            >
              {t('auth.common.goToSignIn')}
            </button>
          </p>
        }
      >
        <div
          className="flex items-start gap-2 rounded-lg border border-red-200 bg-red-50 px-3 py-2.5"
          role="alert"
        >
          <AlertCircle className="h-4 w-4 text-red-500 mt-0.5 shrink-0" />
          <p className="text-xs text-red-600 leading-relaxed">{t('auth.invite.askAdmin')}</p>
        </div>
      </AuthShell>
    );
  }

  // Success, redirect to the app they were invited to, falling back to dash.
  // Validate the first app is a known AppId before indexing APP_URLS - an
  // empty/malformed apps array used to slip through `as AppId | undefined`.
  const firstApp =
    apps.length > 0 && (apps[0] as string) in APP_URLS ? (apps[0] as AppId) : undefined;
  const signInHref = firstApp ? APP_URLS[firstApp] : APP_URLS.dash;

  if (success) {
    return (
      <AuthShell
        title={t('auth.invite.allSet')}
        subtitle={t('auth.invite.passwordSet')}
        footer={
          <Button
            variant="primary"
            className="auth-submit-button"
            fullWidth
            height={40}
            onClick={() => (window.location.href = signInHref)}
          >
            {t('auth.common.signIn')}
          </Button>
        }
      >
        <div className="rounded-lg border border-green-200 bg-green-50 px-3 py-2.5">
          <p className="text-xs text-green-700 leading-relaxed">
            Welcome to rtylr, {fullName || email}. Sign in to get started.
          </p>
        </div>
      </AuthShell>
    );
  }

  // Set password form
  return (
    <>
      <HelmetContainer title={t('auth.invite.metaTitle')} />
      <AuthShell
        title={t('auth.invite.setPasswordTitle')}
        subtitle={`You've been invited to join rtylr${apps.length > 0 ? ` (${apps.join(', ')})` : ''}${email ? ` as ${email}` : ''}.`}
      >
        <form
          onSubmit={(e) => {
            e.preventDefault();
            handleSubmit();
          }}
          className="flex flex-col gap-4"
        >
          {/* Password */}
          <div className="flex flex-col gap-1.5">
            <label htmlFor="invite-password" className="text-sm font-medium text-[#111]">
              {t('auth.common.password')}
            </label>
            <div className="relative">
              <input
                id="invite-password"
                type={showPassword ? 'text' : 'password'}
                placeholder={t('auth.common.atLeastEight')}
                autoComplete="new-password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                autoFocus
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

          {/* Confirm password */}
          <div className="flex flex-col gap-1.5">
            <label htmlFor="invite-confirm" className="text-sm font-medium text-[#111]">
              {t('auth.common.confirmPassword')}
            </label>
            <div className="relative">
              <input
                id="invite-confirm"
                type={showConfirmPassword ? 'text' : 'password'}
                placeholder={t('auth.common.repeatPassword')}
                autoComplete="new-password"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                required
                className="w-full h-10 rounded-lg border border-[#e5e5e5] bg-white px-3 pr-10 text-base md:text-sm text-[#111] placeholder:text-[#999] outline-none transition-colors focus:border-[#4583C4] focus:ring-2 focus:ring-[#4583C4]/20"
              />
              <button
                type="button"
                onClick={() => setShowConfirmPassword((v) => !v)}
                aria-label={showConfirmPassword ? 'Hide confirm password' : 'Show confirm password'}
                className="absolute right-2.5 top-1/2 -translate-y-1/2 text-[#999] hover:text-[#666] text-xs font-medium min-h-[36px] px-1 flex items-center"
              >
                {showConfirmPassword ? 'Hide' : 'Show'}
              </button>
            </div>
          </div>

          {/* Error */}
          {error && (
            <div
              className="flex items-start gap-2 rounded-lg border border-red-200 bg-red-50 px-3 py-2.5"
              role="alert"
            >
              <AlertCircle className="h-4 w-4 text-red-500 mt-0.5 shrink-0" />
              <p className="text-xs text-red-600 leading-relaxed">{error}</p>
            </div>
          )}

          {/* Submit */}
          <Button
            type="submit"
            variant="primary"
            className="auth-submit-button"
            fullWidth
            loading={loading}
            height={40}
          >
            {t('auth.invite.submitJoin')}
          </Button>
        </form>
      </AuthShell>
    </>
  );
};

export default InviteAccept;
