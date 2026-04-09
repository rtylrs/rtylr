import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { motion } from 'framer-motion';
import { Copy, Check, QrCode } from 'lucide-react';

import AuthShell from '../components/AuthShell';
import AuthButton from '../components/AuthButton';
import OTPInput from '../components/OTPInput';
import RecoveryCodeList from '../components/RecoveryCodeList';
import BackLink from '../components/BackLink';
import { use2FASetup, use2FAVerify } from '../hooks/use2FA';
import notify from '../lib/notify/notify';

type Step = 'scan' | 'verify' | 'recovery';

const TwoFactorSetup: React.FC = () => {
  const navigate = useNavigate();
  const { data, isLoading } = use2FASetup();
  const verifyMutation = use2FAVerify();

  const [step, setStep] = useState<Step>('scan');
  const [code, setCode] = useState('');
  const [copied, setCopied] = useState(false);

  const handleCopySecret = () => {
    if (!data?.secret) return;
    navigator.clipboard.writeText(data.secret);
    setCopied(true);
    notify.success('Secret copied');
    setTimeout(() => setCopied(false), 1500);
  };

  const handleVerify = () => {
    verifyMutation.mutate(
      { code },
      {
        onSuccess: () => setStep('recovery'),
        onError: () => notify.error('Invalid code, try again'),
      }
    );
  };

  return (
    <AuthShell
      title="Set up two-factor authentication"
      subtitle="Add an extra layer of security to your rtylr account."
      footer={
        <div className="flex justify-center">
          <BackLink label="Skip for now" />
        </div>
      }
    >
      <div className="flex flex-col gap-6">
        {step === 'scan' && (
          <motion.div
            key="scan"
            initial={{ opacity: 0, y: 12 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.4 }}
            className="flex flex-col gap-5"
          >
            <ol className="flex flex-col gap-2 text-sm text-white/70">
              <li className="flex gap-2">
                <span className="font-semibold text-[#64C5B9]">1.</span>
                Install Google Authenticator, 1Password, or Authy.
              </li>
              <li className="flex gap-2">
                <span className="font-semibold text-[#64C5B9]">2.</span>
                Scan the QR code below.
              </li>
              <li className="flex gap-2">
                <span className="font-semibold text-[#64C5B9]">3.</span>
                Enter the 6-digit code from the app.
              </li>
            </ol>

            <div className="flex justify-center rounded-2xl border border-white/10 bg-white/[0.03] p-5">
              {isLoading ? (
                <div className="h-[220px] w-[220px] flex items-center justify-center text-white/40">
                  <QrCode className="h-10 w-10 animate-pulse" />
                </div>
              ) : (
                <img
                  src={data?.qrCodeUrl}
                  alt="Two-factor QR code"
                  className="h-[220px] w-[220px] rounded-xl bg-white p-2"
                />
              )}
            </div>

            {data?.secret && (
              <button
                type="button"
                onClick={handleCopySecret}
                className="mx-auto inline-flex items-center gap-2 rounded-lg border border-white/10 bg-white/[0.04] hover:bg-white/[0.08] px-3 py-2 text-xs font-mono text-white/80 transition-colors"
              >
                {copied ? <Check className="h-3.5 w-3.5 text-[#64C5B9]" /> : <Copy className="h-3.5 w-3.5" />}
                {data.secret}
              </button>
            )}

            <AuthButton onClick={() => setStep('verify')}>Continue</AuthButton>
          </motion.div>
        )}

        {step === 'verify' && (
          <motion.div
            key="verify"
            initial={{ opacity: 0, y: 12 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.4 }}
            className="flex flex-col gap-5"
          >
            <p className="text-sm text-white/65">
              Enter the 6-digit code from your authenticator app.
            </p>
            <OTPInput value={code} onChange={setCode} onComplete={handleVerify} />
            <AuthButton
              onClick={handleVerify}
              loading={verifyMutation.isPending}
              disabled={code.length !== 6}
            >
              Verify and enable
            </AuthButton>
          </motion.div>
        )}

        {step === 'recovery' && data && (
          <motion.div
            key="recovery"
            initial={{ opacity: 0, y: 12 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.4 }}
            className="flex flex-col gap-5"
          >
            <p className="text-sm text-white/65">
              Save your recovery codes. You will not see them again.
            </p>
            <RecoveryCodeList codes={data.recoveryCodes} />
            <AuthButton onClick={() => navigate('/')}>
              I saved my codes
            </AuthButton>
          </motion.div>
        )}
      </div>
    </AuthShell>
  );
};

export default TwoFactorSetup;
