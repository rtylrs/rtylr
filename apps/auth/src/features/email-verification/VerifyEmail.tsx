import React, { useEffect, useState } from 'react';
import { useNavigate, useRouterState } from '@tanstack/react-router';
import { Mail, CheckCircle2, AlertCircle, Loader2 } from 'lucide-react';

import Logo from '../../shared/components/Logo';
import { Button } from '@voxire/components';
import { useEmailVerification, useResendVerificationEmail } from './useEmailVerification';
import notify from '../../shared/lib/notify/notify';

type Status = 'loading' | 'success' | 'error';

const VerifyEmail: React.FC = () => {
  const navigate = useNavigate();
  const searchStr = useRouterState({ select: (s) => s.location.searchStr });
  const params = new URLSearchParams(searchStr);
  const verifyMutation = useEmailVerification();
  const resendMutation = useResendVerificationEmail();
  const [status, setStatus] = useState<Status>('loading');

  useEffect(() => {
    const token = params.get('token');
    if (!token) {
      setStatus('error');
      return;
    }
    verifyMutation.mutate(
      { token },
      {
        onSuccess: () => {
          setStatus('success');
          try {
            window.history.replaceState({}, document.title, window.location.pathname);
          } catch {
            // ignore environment restriction
          }
        },
        onError: () => setStatus('error'),
      }
    );
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const handleResend = (): void => {
    const email = params.get('email') ?? '';
    if (!email) {
      notify.error('No email address found to resend verification');
      return;
    }
    resendMutation.mutate(
      { email },
      {
        onSuccess: () => notify.success('Verification email sent'),
        onError: () => notify.error('Failed to send verification email'),
      }
    );
  };

  return (
    <div className="min-h-screen w-full flex items-center justify-center bg-white px-4">
      <div className="w-full max-w-sm text-center">
        {/* Logo */}
        <div className="flex justify-center mb-8">
          <Logo variant="dark" className="h-8 w-auto select-none" />
        </div>

        {status === 'loading' && (
          <>
            <div className="flex h-16 w-16 items-center justify-center rounded-2xl bg-blue-50 border border-blue-200 mx-auto">
              <Loader2 className="h-8 w-8 text-[#4583C4] animate-spin" />
            </div>
            <p className="mt-4 text-sm text-[#666]">Checking your token...</p>
          </>
        )}
        {status === 'success' && (
          <>
            <div className="flex h-16 w-16 items-center justify-center rounded-2xl bg-green-50 border border-green-200 mx-auto">
              <CheckCircle2 className="h-8 w-8 text-green-600" />
            </div>
            <p className="mt-4 text-base font-semibold text-[#111]">Email verified</p>
            <div className="mt-6">
              <Button
                type="button"
                variant="primary"
                fullWidth
                height={40}
                onClick={() => navigate({ to: '/' })}
              >
                Continue to sign in
              </Button>
            </div>
          </>
        )}
        {status === 'error' && (
          <>
            <div className="flex h-16 w-16 items-center justify-center rounded-2xl bg-red-50 border border-red-200 mx-auto">
              <AlertCircle className="h-8 w-8 text-red-500" />
            </div>
            <p className="mt-4 text-sm text-red-600">
              This verification link is invalid or has expired.
            </p>
            <div className="mt-6">
              <Button
                type="button"
                variant="outline"
                fullWidth
                height={40}
                onClick={handleResend}
                loading={resendMutation.isPending}
              >
                <span className="inline-flex items-center gap-2">
                  <Mail className="h-4 w-4" />
                  Resend verification email
                </span>
              </Button>
            </div>
          </>
        )}

        <div className="mt-6">
          <button
            type="button"
            onClick={() => navigate({ to: '/' })}
            className="text-sm text-[#4583C4] font-medium hover:underline underline-offset-2"
          >
            Back to sign in
          </button>
        </div>
      </div>
    </div>
  );
};

export default VerifyEmail;
