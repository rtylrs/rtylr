import React, { useEffect, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { motion } from 'framer-motion';
import { Loader2, CheckCircle2, AlertCircle } from 'lucide-react';

import AuthShell from '../components/AuthShell';
import AuthButton from '../components/AuthButton';
import BackLink from '../components/BackLink';
import { useMagicLinkConfirm } from '../hooks/useMagicLink';

type Status = 'loading' | 'success' | 'error';

const MagicLinkConfirm: React.FC = () => {
  const navigate = useNavigate();
  const [params] = useSearchParams();
  const mutation = useMagicLinkConfirm();
  const [status, setStatus] = useState<Status>('loading');

  useEffect(() => {
    const token = params.get('token');
    if (!token) {
      setStatus('error');
      return;
    }
    mutation.mutate(
      { token },
      {
        onSuccess: () => setStatus('success'),
        onError: () => setStatus('error'),
      }
    );
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return (
    <AuthShell
      title="Signing you in"
      subtitle="Using the magic link from your email."
      footer={
        <div className="flex justify-center">
          <BackLink label="Back to sign in" />
        </div>
      }
    >
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        className="flex flex-col items-center gap-5 text-center py-6"
      >
        {status === 'loading' && (
          <>
            <Loader2 className="h-10 w-10 text-[#4583C4] animate-spin" />
            <p className="text-sm text-white/60">Verifying your link...</p>
          </>
        )}
        {status === 'success' && (
          <>
            <div className="flex h-16 w-16 items-center justify-center rounded-2xl bg-[#64C5B9]/15 border border-[#64C5B9]/25">
              <CheckCircle2 className="h-8 w-8 text-[#64C5B9]" />
            </div>
            <p className="text-base font-semibold text-white">Signed in</p>
            <AuthButton onClick={() => navigate('/')}>Continue</AuthButton>
          </>
        )}
        {status === 'error' && (
          <>
            <div className="flex h-16 w-16 items-center justify-center rounded-2xl bg-red-500/10 border border-red-400/25">
              <AlertCircle className="h-8 w-8 text-red-300" />
            </div>
            <p className="text-sm text-red-200">
              This magic link is invalid or has expired.
            </p>
            <AuthButton variant="outline" onClick={() => navigate('/magic-link')}>
              Request a new link
            </AuthButton>
          </>
        )}
      </motion.div>
    </AuthShell>
  );
};

export default MagicLinkConfirm;
