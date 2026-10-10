import React, { useState } from 'react';
import { useNavigate } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { ArrowLeft } from 'lucide-react';

import AuthShell from '../../shared/components/AuthShell';
import { AuthAlert, AuthField } from '../../shared/components/AuthField';
import { Button, HelmetContainer } from '@voxire/components';
import { createOneKitClient } from '@voxire/api';
import { AuthServiceClient } from '@voxire/voxie/common/auth/v1/client';
import type { ResetPasswordRequest } from '@voxire/voxie/common/auth/v1/types';

import { getEmailError } from '../../shared/lib/validation';

const authClient = createOneKitClient(AuthServiceClient, 'auth');

const ResetPassword: React.FC = () => {
  const navigate = useNavigate();
  const { t } = useTranslation();
  const [email, setEmail] = useState('');
  const [message, setMessage] = useState('');
  const [loading, setLoading] = useState(false);
  const [requestSent, setRequestSent] = useState(false);

  const emailError = getEmailError(email);

  const handleReset = async (): Promise<void> => {
    if (emailError) {
      setMessage(t('auth.login.invalidEmailFormat', 'Please enter a valid email address.'));
      return;
    }
    setLoading(true);
    try {
      const request: ResetPasswordRequest = { email };
      await authClient.resetPassword(request);
      setRequestSent(true);
      setMessage(t('auth.reset.successMessage'));
    } catch {
      setMessage(t('auth.reset.genericError'));
    } finally {
      setLoading(false);
    }
  };

  return (
    <>
      <HelmetContainer title={t('auth.reset.metaTitle')} />
      <AuthShell
        title={t('auth.reset.title')}
        subtitle={t('auth.reset.subtitle')}
        footer={
          <div className="text-center">
            <button
              type="button"
              onClick={() => navigate({ to: '/' })}
              className="inline-flex items-center gap-1.5 font-medium text-[#111] hover:text-[#111]/70 transition-colors"
            >
              <ArrowLeft className="h-4 w-4" />
              {t('auth.reset.backToSignIn')}
            </button>
          </div>
        }
      >
        <form
          onSubmit={(e) => {
            e.preventDefault();
            handleReset();
          }}
          className="flex flex-col gap-4"
        >
          <AuthField
            id="reset-email"
            label={t('auth.common.email')}
            type="email"
            placeholder={t('auth.login.emailPlaceholder')}
            autoComplete="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
            autoFocus
            disabled={requestSent}
          />

          {message && <AuthAlert tone={requestSent ? 'success' : 'error'}>{message}</AuthAlert>}

          <Button
            type="submit"
            variant="primary"
            className="auth-submit-button"
            fullWidth
            loading={loading}
            disabled={requestSent}
            height={44}
          >
            {requestSent ? t('auth.reset.submitted') : t('auth.reset.submit')}
          </Button>
        </form>
      </AuthShell>
    </>
  );
};

export default ResetPassword;
