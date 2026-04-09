import React, { useEffect, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { motion } from 'framer-motion';
import { Mail, CheckCircle2, AlertCircle, Loader2 } from 'lucide-react';

import AuthShell from '../components/AuthShell';
import AuthButton from '../components/AuthButton';
import BackLink from '../components/BackLink';
import {
  useEmailVerification,
  useResendVerificationEmail,
} from '../hooks/useEmailVerification';
import notify from '../lib/notify/notify';

type Status = 'loading' | 'success' | 'error';

const VerifyEmail: React.FC = () => {
  const navigate = useNavigate();
  const [params] = useSearchParams();
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
        onSuccess: () => setStatus('success'),
        onError: () => setStatus('error'),
      }
    );
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const handleResend = () => {
    resendMutation.mutate(
      { email: 'you@company.com' },
      {
        onSuccess: () => notify.success('Verification email sent'),
      }
    );
  };

  return (
    <AuthShell
      title="Verifying your email"
      subtitle="This will only take a moment."
      footer={
        <div className="flex justify-center">
          <BackLink label="Back to sign in" />
        </div>
      }
    >
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        className="flex flex-col items-center gap-5 text-center py-4"
      >
        {status === 'loading' && (
          <>
            <div className="flex h-16 w-16 items-center justify-center rounded-2xl bg-[#4583C4]/15 border border-[#4583C4]/25">
              <Loader2 className="h-8 w-8 text-[#4583C4] animate-spin" />
            </div>
            <p className="text-sm text-white/65">Checking your token...</p>
          </>
        )}
        {status === 'success' && (
          <>
            <div className="flex h-16 w-16 items-center justify-center rounded-2xl bg-[#64C5B9]/15 border border-[#64C5B9]/25">
              <CheckCircle2 className="h-8 w-8 text-[#64C5B9]" />
            </div>
            <p className="text-base font-semibold text-white">
              Email verified
            </p>
            <AuthButton onClick={() => navigate('/')}>
              Continue to sign in
            </AuthButton>
          </>
        )}
        {status === 'error' && (
          <>
            <div className="flex h-16 w-16 items-center justify-center rounded-2xl bg-red-500/10 border border-red-400/25">
              <AlertCircle className="h-8 w-8 text-red-300" />
            </div>
            <p className="text-sm text-red-200">
              This verification link is invalid or has expired.
            </p>
            <AuthButton
              variant="outline"
              onClick={handleResend}
              loading={resendMutation.isPending}
            >
              <span className="inline-flex items-center justify-center gap-2">
                <Mail className="h-[18px] w-[18px]" />
                Resend verification email
              </span>
            </AuthButton>
          </>
        )}
      </motion.div>
    </AuthShell>
  );
};

export default VerifyEmail;
