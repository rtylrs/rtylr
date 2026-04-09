import React, { useState } from 'react';
import { Building2, Mail } from 'lucide-react';

import AuthShell from '../components/AuthShell';
import AuthInput from '../components/AuthInput';
import SubmitButton from '../components/SubmitButton';
import BackLink from '../components/BackLink';
import { getEmailError } from '../lib/validation';
import notify from '../lib/notify/notify';

const SSO: React.FC = () => {
  const [email, setEmail] = useState('');
  const [loading, setLoading] = useState(false);
  const emailError = getEmailError(email);

  const handleSubmit = () => {
    setLoading(true);
    setTimeout(() => {
      setLoading(false);
      notify.success('Redirecting to your identity provider...');
    }, 800);
  };

  return (
    <AuthShell
      title="Sign in with SSO"
      subtitle="Enter your work email to continue to your identity provider."
      footer={
        <div className="flex justify-center">
          <BackLink label="Back to sign in" />
        </div>
      }
    >
      <form
        onSubmit={(e) => {
          e.preventDefault();
          handleSubmit();
        }}
        className="flex flex-col gap-5"
      >
        <div className="flex justify-center">
          <div className="flex h-14 w-14 items-center justify-center rounded-2xl bg-[#4583C4]/15 border border-[#4583C4]/25">
            <Building2 className="h-6 w-6 text-[#4583C4]" />
          </div>
        </div>

        <AuthInput
          id="sso-email"
          type="email"
          label="Work email"
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
          label="Continue with SSO"
          loading={loading}
          disabled={!email || !!emailError}
        />

        <p className="text-[11px] text-center text-white/40 leading-relaxed">
          Your organization must have SSO configured with rtylr.
        </p>
      </form>
    </AuthShell>
  );
};

export default SSO;
