import React, { useState } from 'react';
import { useNavigate } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { AlertCircle, ArrowLeft, CheckCircle2 } from 'lucide-react';

import Logo from '../../shared/components/Logo';
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
      <div className="min-h-screen w-full flex items-center justify-center bg-white px-4">
        <div className="w-full max-w-sm">
          <div className="flex justify-center mb-8">
            <Logo variant="dark" className="h-8 w-auto select-none" />
          </div>

          <h2 className="text-xl font-semibold text-[#111] mb-1">{t('auth.reset.title')}</h2>
          <p className="text-sm text-[#666] mb-6">{t('auth.reset.subtitle')}</p>

          <form
            onSubmit={(e) => {
              e.preventDefault();
              handleReset();
            }}
            className="flex flex-col gap-4"
          >
            <div className="flex flex-col gap-1.5">
              <label htmlFor="reset-email" className="text-sm font-medium text-[#111]">
                {t('auth.common.email')}
              </label>
              <input
                id="reset-email"
                type="email"
                placeholder={t('auth.login.emailPlaceholder')}
                autoComplete="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
                autoFocus
                disabled={requestSent}
                className="w-full h-10 rounded-lg border border-[#e5e5e5] bg-white px-3 text-base md:text-sm text-[#111] placeholder:text-[#999] outline-none transition-colors focus:border-[#4583C4] focus:ring-2 focus:ring-[#4583C4]/20 disabled:bg-[#f5f5f5] disabled:text-[#999]"
              />
            </div>

            {message && (
              <div
                className={[
                  'flex items-start gap-2 rounded-lg px-3 py-2.5 border',
                  requestSent ? 'border-green-200 bg-green-50' : 'border-red-200 bg-red-50',
                ].join(' ')}
                role="alert"
              >
                {requestSent ? (
                  <CheckCircle2 className="h-4 w-4 text-green-600 mt-0.5 shrink-0" />
                ) : (
                  <AlertCircle className="h-4 w-4 text-red-500 mt-0.5 shrink-0" />
                )}
                <p
                  className={[
                    'text-xs leading-relaxed',
                    requestSent ? 'text-green-700' : 'text-red-600',
                  ].join(' ')}
                >
                  {message}
                </p>
              </div>
            )}

            <Button
              type="submit"
              variant="primary"
              className="auth-submit-button"
              fullWidth
              loading={loading}
              disabled={requestSent}
              height={40}
            >
              {requestSent ? t('auth.reset.submitted') : t('auth.reset.submit')}
            </Button>
          </form>

          <div className="mt-6 text-center">
            <button
              type="button"
              onClick={() => navigate({ to: '/' })}
              className="inline-flex items-center gap-1.5 text-sm text-[#4583C4] font-medium hover:underline underline-offset-2"
            >
              <ArrowLeft className="h-4 w-4" />
              {t('auth.reset.backToSignIn')}
            </button>
          </div>
        </div>
      </div>
    </>
  );
};

export default ResetPassword;
