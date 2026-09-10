import React, { useMemo, useState } from 'react';
import { useNavigate } from '@tanstack/react-router';
import { AlertCircle, CheckCircle2 } from 'lucide-react';
import Logo from '../../shared/components/Logo';
import { Button, HelmetContainer } from '@voxire/components';
import { createOneKitClient, getOneKitErrorMessage } from '@voxire/api';
import { AuthServiceClient } from '@voxire/voxie/common/auth/v1/client';
import type { ResetPasswordConfirmRequest } from '@voxire/voxie/common/auth/v1/types';

import { isStrongPassword } from '../../shared/lib/validation';

const authClient = createOneKitClient(AuthServiceClient, 'auth');

const ResetPasswordConfirm: React.FC = () => {
  const navigate = useNavigate();
  const [password, setPassword] = useState('');
  const [confirm, setConfirm] = useState('');
  const [loading, setLoading] = useState(false);
  const [done, setDone] = useState(false);
  const [error, setError] = useState('');

  const token = useMemo(() => {
    return new URLSearchParams(window.location.search).get('token') ?? '';
  }, []);

  const mismatch = confirm.length > 0 && password !== confirm;

  const handleSubmit = async (): Promise<void> => {
    if (!token) {
      setError('Invalid reset link. Please request a new one.');
      return;
    }
    if (!isStrongPassword(password)) {
      setError(
        'Password must be at least 8 characters and include uppercase, lowercase, and numbers.'
      );
      return;
    }
    if (password !== confirm) {
      setError('Passwords do not match.');
      return;
    }
    setLoading(true);
    setError('');
    try {
      const request: ResetPasswordConfirmRequest = { token, password };
      await authClient.resetPasswordConfirm(request);
      setDone(true);
      try {
        window.history.replaceState({}, document.title, window.location.pathname);
      } catch {
        // ignore environment restriction
      }
    } catch (err: unknown) {
      setError(getOneKitErrorMessage(err, 'Failed to reset password.'));
    } finally {
      setLoading(false);
    }
  };

  return (
    <>
      <HelmetContainer title="Reset password · rtylr" />
      <div className="min-h-screen w-full flex items-center justify-center bg-white px-4">
        <div className="w-full max-w-sm">
          <div className="flex justify-center mb-8">
            <Logo variant="dark" className="h-8 w-auto select-none" />
          </div>

          <h2 className="text-xl font-semibold text-[#111] mb-1">
            {t('auth.reset.newPasswordTitle')}
          </h2>
          <p className="text-sm text-[#666] mb-6">{t('auth.reset.newPasswordSubtitle')}</p>

          {done ? (
            <div className="flex flex-col gap-4">
              <div
                className="flex items-start gap-2 rounded-lg border border-green-200 bg-green-50 px-3 py-2.5"
                role="alert"
              >
                <CheckCircle2 className="h-4 w-4 text-green-600 mt-0.5 shrink-0" />
                <p className="text-xs text-green-700 leading-relaxed">
                  Password updated. You can now sign in.
                </p>
              </div>
              <Button
                type="button"
                variant="primary"
                className="auth-submit-button"
                fullWidth
                height={40}
                onClick={() => navigate({ to: '/' })}
              >
                Go to sign in
              </Button>
            </div>
          ) : (
            <form
              onSubmit={(e) => {
                e.preventDefault();
                void handleSubmit();
              }}
              className="flex flex-col gap-4"
            >
              <div className="flex flex-col gap-1.5">
                <label htmlFor="new-password" className="text-sm font-medium text-[#111]">
                  New password
                </label>
                <input
                  id="new-password"
                  type="password"
                  placeholder="At least 8 characters"
                  autoComplete="new-password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  required
                  autoFocus
                  className="w-full h-10 rounded-lg border border-[#e5e5e5] bg-white px-3 text-sm text-[#111] placeholder:text-[#999] outline-none transition-colors focus:border-[#4583C4] focus:ring-2 focus:ring-[#4583C4]/20"
                />
              </div>

              <div className="flex flex-col gap-1.5">
                <label htmlFor="confirm-password" className="text-sm font-medium text-[#111]">
                  Confirm password
                </label>
                <input
                  id="confirm-password"
                  type="password"
                  placeholder="Repeat password"
                  autoComplete="new-password"
                  value={confirm}
                  onChange={(e) => setConfirm(e.target.value)}
                  required
                  className="w-full h-10 rounded-lg border border-[#e5e5e5] bg-white px-3 text-sm text-[#111] placeholder:text-[#999] outline-none transition-colors focus:border-[#4583C4] focus:ring-2 focus:ring-[#4583C4]/20"
                />
                {mismatch && (
                  <p className="text-xs text-red-500">{t('auth.reset.passwordsDoNotMatch')}</p>
                )}
              </div>

              {error && (
                <div
                  className="flex items-start gap-2 rounded-lg border border-red-200 bg-red-50 px-3 py-2.5"
                  role="alert"
                >
                  <AlertCircle className="h-4 w-4 text-red-500 mt-0.5 shrink-0" />
                  <p className="text-xs text-red-600 leading-relaxed">{error}</p>
                </div>
              )}

              <Button
                type="submit"
                variant="primary"
                className="auth-submit-button"
                fullWidth
                loading={loading}
                disabled={!password || !!mismatch}
                height={40}
              >
                Reset password
              </Button>
            </form>
          )}
        </div>
      </div>
    </>
  );
};

export default ResetPasswordConfirm;
