import React, { useState } from 'react';
import { motion } from 'framer-motion';
import { Mail, Wand2, CheckCircle2 } from 'lucide-react';

import AuthShell from '../components/AuthShell';
import AuthInput from '../components/AuthInput';
import SubmitButton from '../components/SubmitButton';
import BackLink from '../components/BackLink';
import { useMagicLinkRequest } from '../hooks/useMagicLink';
import { getEmailError } from '../lib/validation';

const MagicLink: React.FC = () => {
  const [email, setEmail] = useState('');
  const [sent, setSent] = useState(false);
  const mutation = useMagicLinkRequest();
  const emailError = getEmailError(email);

  const handleSubmit = () => {
    mutation.mutate(
      { email },
      {
        onSuccess: () => setSent(true),
      }
    );
  };

  return (
    <AuthShell
      title="Sign in with a magic link"
      subtitle="We'll email you a secure link that logs you in instantly."
      footer={
        <div className="flex justify-center">
          <BackLink label="Back to sign in" />
        </div>
      }
    >
      {sent ? (
        <motion.div
          initial={{ opacity: 0, y: 12 }}
          animate={{ opacity: 1, y: 0 }}
          className="flex flex-col items-center gap-5 text-center"
        >
          <div className="flex h-16 w-16 items-center justify-center rounded-2xl bg-[#64C5B9]/15 border border-[#64C5B9]/25">
            <CheckCircle2 className="h-8 w-8 text-[#64C5B9]" />
          </div>
          <div>
            <p className="text-base font-semibold text-white">
              Check your inbox
            </p>
            <p className="mt-1 text-sm text-white/55">
              We sent a link to{' '}
              <span className="text-white/85 font-medium">{email}</span>. It
              expires in 15 minutes.
            </p>
          </div>
          <button
            type="button"
            onClick={() => setSent(false)}
            className="text-xs text-white/55 hover:text-[#64C5B9] transition-colors"
          >
            Use a different email
          </button>
        </motion.div>
      ) : (
        <form
          onSubmit={(e) => {
            e.preventDefault();
            handleSubmit();
          }}
          className="flex flex-col gap-5"
        >
          <div className="flex justify-center">
            <div className="flex h-14 w-14 items-center justify-center rounded-2xl bg-[#7E4496]/15 border border-[#7E4496]/25">
              <Wand2 className="h-6 w-6 text-[#c79edc]" />
            </div>
          </div>
          <AuthInput
            id="magic-email"
            type="email"
            label="Email"
            icon={Mail}
            placeholder="you@company.com"
            autoComplete="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            error={emailError}
            required
            autoFocus
          />
          <SubmitButton
            label="Send magic link"
            loading={mutation.isPending}
            disabled={!email || !!emailError}
          />
        </form>
      )}
    </AuthShell>
  );
};

export default MagicLink;
