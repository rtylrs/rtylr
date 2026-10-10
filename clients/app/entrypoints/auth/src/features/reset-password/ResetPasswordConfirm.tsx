import React, { useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from '@tanstack/react-router';
import AuthShell from '../../shared/components/AuthShell';
import { AuthAlert, AuthPasswordField } from '../../shared/components/AuthField';
import { Button, HelmetContainer } from '@voxire/components';
import { createOneKitClient, getOneKitErrorMessage } from '@voxire/api';
import { AuthServiceClient } from '@voxire/voxie/common/auth/v1/client';
import type { ResetPasswordConfirmRequest } from '@voxire/voxie/common/auth/v1/types';

import { isStrongPassword } from '../../shared/lib/validation';

const authClient = createOneKitClient(AuthServiceClient, 'auth');

const ResetPasswordConfirm: React.FC = () => {
  const { t } = useTranslation();
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
      <HelmetContainer title={t('auth.reset.metaTitle')} />
      <AuthShell
        title={t('auth.reset.newPasswordTitle')}
        subtitle={t('auth.reset.newPasswordSubtitle')}
      >
        {done ? (
          <div className="flex flex-col gap-4">
            <AuthAlert tone="success">{t('auth.reset.updated')}</AuthAlert>
            <Button
              type="button"
              variant="primary"
              className="auth-submit-button"
              fullWidth
              height={44}
              onClick={() => navigate({ to: '/' })}
            >
              {t('auth.common.goToSignIn')}
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
            <AuthPasswordField
              id="new-password"
              label={t('auth.reset.newPassword')}
              placeholder={t('auth.common.atLeastEight')}
              autoComplete="new-password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              autoFocus
            />

            <AuthPasswordField
              id="confirm-password"
              label={t('auth.common.confirmPassword')}
              placeholder={t('auth.common.repeatPassword')}
              autoComplete="new-password"
              value={confirm}
              onChange={(e) => setConfirm(e.target.value)}
              required
              error={mismatch ? t('auth.reset.passwordsDoNotMatch') : null}
            />

            {error && <AuthAlert>{error}</AuthAlert>}

            <Button
              type="submit"
              variant="primary"
              className="auth-submit-button"
              fullWidth
              loading={loading}
              disabled={!password || !!mismatch}
              height={44}
            >
              {t('auth.reset.submitNew')}
            </Button>
          </form>
        )}
      </AuthShell>
    </>
  );
};

export default ResetPasswordConfirm;
