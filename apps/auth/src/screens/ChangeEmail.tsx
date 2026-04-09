import React, { useState } from 'react';
import { Mail, Lock } from 'lucide-react';

import AuthShell from '../components/AuthShell';
import AuthInput from '../components/AuthInput';
import SubmitButton from '../components/SubmitButton';
import BackLink from '../components/BackLink';
import { getEmailError } from '../lib/validation';
import notify from '../lib/notify/notify';

const ChangeEmail: React.FC = () => {
  const [currentEmail] = useState('you@company.com');
  const [newEmail, setNewEmail] = useState('');
  const [password, setPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const newEmailError = getEmailError(newEmail);

  const handleSubmit = () => {
    setLoading(true);
    setTimeout(() => {
      setLoading(false);
      notify.success('Verification link sent to ' + newEmail);
    }, 700);
  };

  return (
    <AuthShell
      title="Change email address"
      subtitle="Confirm your password, then we'll send a verification link to your new address."
      footer={
        <div className="flex justify-center">
          <BackLink label="Back" />
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
        <div className="rounded-xl border border-white/10 bg-white/[0.03] p-4">
          <p className="text-[11px] uppercase tracking-[0.08em] text-white/40">
            Current email
          </p>
          <p className="mt-1 text-sm text-white/85">{currentEmail}</p>
        </div>

        <AuthInput
          id="change-new-email"
          type="email"
          label="New email"
          icon={Mail}
          placeholder="new@company.com"
          autoComplete="email"
          value={newEmail}
          onChange={(e) => setNewEmail(e.target.value)}
          error={newEmailError}
          required
          autoFocus
        />

        <AuthInput
          id="change-password"
          type="password"
          label="Current password"
          icon={Lock}
          placeholder="Enter your password"
          autoComplete="current-password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          required
        />

        <SubmitButton
          label="Send verification link"
          loading={loading}
          disabled={!newEmail || !password || !!newEmailError}
        />
      </form>
    </AuthShell>
  );
};

export default ChangeEmail;
