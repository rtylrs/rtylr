import React, { useState } from 'react';
import { motion } from 'framer-motion';
import { Mail, KeyRound, MessageSquare, Check } from 'lucide-react';

import AuthShell from '../components/AuthShell';
import AuthInput from '../components/AuthInput';
import AuthButton from '../components/AuthButton';
import SubmitButton from '../components/SubmitButton';
import BackLink from '../components/BackLink';
import OTPInput from '../components/OTPInput';
import { getEmailError } from '../lib/validation';

type Method = 'email' | 'sms' | 'questions';
type Step = 1 | 2 | 3;

const AccountRecovery: React.FC = () => {
  const [step, setStep] = useState<Step>(1);
  const [method, setMethod] = useState<Method>('email');
  const [email, setEmail] = useState('');
  const [otp, setOtp] = useState('');
  const [done, setDone] = useState(false);

  const next = () => setStep((s) => (Math.min((s + 1), 3) as Step));

  const methods: Array<{
    key: Method;
    icon: React.ComponentType<{ className?: string }>;
    title: string;
    subtitle: string;
  }> = [
    {
      key: 'email',
      icon: Mail,
      title: 'Email',
      subtitle: 'Receive a recovery link by email',
    },
    {
      key: 'sms',
      icon: MessageSquare,
      title: 'SMS',
      subtitle: 'Text me a one-time code',
    },
    {
      key: 'questions',
      icon: KeyRound,
      title: 'Security questions',
      subtitle: 'Use my backup questions',
    },
  ];

  return (
    <AuthShell
      title="Recover your account"
      subtitle={`Step ${step} of 3`}
      footer={
        <div className="flex justify-center">
          <BackLink label="Back to sign in" />
        </div>
      }
    >
      <div className="flex flex-col gap-5">
        <div className="flex gap-1.5">
          {[1, 2, 3].map((s) => (
            <div
              key={s}
              className={[
                'h-1 flex-1 rounded-full transition-colors',
                s <= step ? 'bg-[#4583C4]' : 'bg-white/10',
              ].join(' ')}
            />
          ))}
        </div>

        {step === 1 && (
          <motion.div
            key="s1"
            initial={{ opacity: 0, y: 12 }}
            animate={{ opacity: 1, y: 0 }}
            className="flex flex-col gap-3"
          >
            <p className="text-sm text-white/65">
              Choose how you want to recover your account.
            </p>
            {methods.map(({ key, icon: Icon, title, subtitle }) => (
              <button
                key={key}
                type="button"
                onClick={() => setMethod(key)}
                className={[
                  'flex items-start gap-3 rounded-xl border px-4 py-3.5 text-left transition-colors',
                  method === key
                    ? 'border-[#4583C4] bg-[#4583C4]/10'
                    : 'border-white/10 bg-white/[0.03] hover:bg-white/[0.05]',
                ].join(' ')}
              >
                <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-white/[0.05] border border-white/10">
                  <Icon className="h-4 w-4 text-white/80" />
                </div>
                <div className="flex-1">
                  <p className="text-sm font-semibold text-white">{title}</p>
                  <p className="text-[11px] text-white/50 mt-0.5">{subtitle}</p>
                </div>
                {method === key && (
                  <Check className="h-4 w-4 text-[#4583C4] mt-2" />
                )}
              </button>
            ))}
            <AuthButton onClick={next}>Continue</AuthButton>
          </motion.div>
        )}

        {step === 2 && (
          <motion.div
            key="s2"
            initial={{ opacity: 0, y: 12 }}
            animate={{ opacity: 1, y: 0 }}
            className="flex flex-col gap-5"
          >
            <AuthInput
              id="recovery-email"
              type="email"
              label={method === 'sms' ? 'Phone number' : 'Email'}
              icon={Mail}
              placeholder={
                method === 'sms' ? '+971 50 000 0000' : 'you@company.com'
              }
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              error={method === 'email' ? getEmailError(email) : undefined}
              autoFocus
            />
            <SubmitButton
              label="Send code"
              onClick={next}
              disabled={!email}
            />
          </motion.div>
        )}

        {step === 3 && !done && (
          <motion.div
            key="s3"
            initial={{ opacity: 0, y: 12 }}
            animate={{ opacity: 1, y: 0 }}
            className="flex flex-col gap-5"
          >
            <p className="text-sm text-white/65">
              Enter the 6-digit code we sent to {email}.
            </p>
            <OTPInput value={otp} onChange={setOtp} onComplete={() => setDone(true)} />
            <AuthButton
              onClick={() => setDone(true)}
              disabled={otp.length !== 6}
            >
              Recover account
            </AuthButton>
          </motion.div>
        )}

        {done && (
          <motion.div
            key="done"
            initial={{ opacity: 0, y: 12 }}
            animate={{ opacity: 1, y: 0 }}
            className="flex flex-col items-center gap-4 text-center py-4"
          >
            <div className="flex h-16 w-16 items-center justify-center rounded-2xl bg-[#64C5B9]/15 border border-[#64C5B9]/25">
              <Check className="h-8 w-8 text-[#64C5B9]" strokeWidth={3} />
            </div>
            <p className="text-sm text-white/80">
              You're all set. Redirecting you to reset your password...
            </p>
          </motion.div>
        )}
      </div>
    </AuthShell>
  );
};

export default AccountRecovery;
