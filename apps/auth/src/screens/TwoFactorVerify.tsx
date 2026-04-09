import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { motion } from 'framer-motion';
import { ShieldCheck } from 'lucide-react';

import AuthShell from '../components/AuthShell';
import AuthButton from '../components/AuthButton';
import OTPInput from '../components/OTPInput';
import BackLink from '../components/BackLink';
import { use2FAVerify } from '../hooks/use2FA';
import notify from '../lib/notify/notify';

const TwoFactorVerify: React.FC = () => {
  const navigate = useNavigate();
  const verifyMutation = use2FAVerify();
  const [code, setCode] = useState('');
  const [useRecovery, setUseRecovery] = useState(false);
  const [recoveryCode, setRecoveryCode] = useState('');

  const handleVerify = () => {
    verifyMutation.mutate(
      { code: useRecovery ? recoveryCode : code },
      {
        onSuccess: () => {
          notify.success('Welcome back');
          navigate('/');
        },
        onError: () => notify.error('Invalid code'),
      }
    );
  };

  return (
    <AuthShell
      title="Two-factor authentication"
      subtitle="Enter the 6-digit code from your authenticator app."
      footer={
        <div className="flex flex-col items-center gap-3">
          <button
            type="button"
            onClick={() => setUseRecovery((v) => !v)}
            className="text-xs text-white/55 hover:text-[#64C5B9] transition-colors"
          >
            {useRecovery ? 'Use authenticator app instead' : 'Use a recovery code'}
          </button>
          <BackLink label="Back to sign in" />
        </div>
      }
    >
      <form
        onSubmit={(e) => {
          e.preventDefault();
          handleVerify();
        }}
        className="flex flex-col gap-5"
      >
        <div className="flex justify-center">
          <div className="flex h-16 w-16 items-center justify-center rounded-2xl bg-[#4583C4]/15 border border-[#4583C4]/25">
            <ShieldCheck className="h-8 w-8 text-[#4583C4]" />
          </div>
        </div>

        {useRecovery ? (
          <motion.div
            initial={{ opacity: 0, y: 6 }}
            animate={{ opacity: 1, y: 0 }}
            className="flex flex-col gap-2"
          >
            <label className="text-[11px] font-semibold uppercase tracking-[0.08em] text-white/50">
              Recovery code
            </label>
            <input
              autoFocus
              type="text"
              value={recoveryCode}
              onChange={(e) => setRecoveryCode(e.target.value.toUpperCase())}
              placeholder="XXXX-XXXX"
              className="w-full h-[52px] rounded-xl bg-white/[0.04] border border-white/10 hover:bg-white/[0.06] hover:border-white/15 focus:border-[#4583C4] focus:ring-2 focus:ring-[#4583C4]/25 outline-none px-4 text-[15px] font-mono text-white placeholder:text-white/30 transition-all"
            />
          </motion.div>
        ) : (
          <OTPInput value={code} onChange={setCode} onComplete={handleVerify} />
        )}

        <AuthButton
          type="submit"
          loading={verifyMutation.isPending}
          disabled={
            useRecovery ? !recoveryCode.trim() : code.length !== 6
          }
        >
          Verify and continue
        </AuthButton>
      </form>
    </AuthShell>
  );
};

export default TwoFactorVerify;
