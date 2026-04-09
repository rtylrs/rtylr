import React, { useEffect, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { motion } from 'framer-motion';
import { AlertCircle, Check, Clock, Loader2, User } from 'lucide-react';

import AuthShell from '../components/AuthShell';
import AuthButton from '../components/AuthButton';
import AuthInput from '../components/AuthInput';
import PasswordStrengthMeter from '../components/PasswordStrengthMeter';
import TermsCheckbox from '../components/TermsCheckbox';
import { useInvitation, useAcceptInvitation } from '../hooks/useInvitation';
import { scorePassword } from '../components/PasswordStrengthMeter';

const InvitationAccept: React.FC = () => {
  const { token } = useParams<{ token: string }>();
  const navigate = useNavigate();
  const { data, isLoading, isError } = useInvitation(token);
  const acceptMutation = useAcceptInvitation();

  const [password, setPassword] = useState('');
  const [agree, setAgree] = useState(false);
  const [countdown, setCountdown] = useState<string>('');

  useEffect(() => {
    if (!data?.expiresAt) return;
    const tick = () => {
      const diff = new Date(data.expiresAt).getTime() - Date.now();
      if (diff <= 0) {
        setCountdown('Expired');
        return;
      }
      const h = Math.floor(diff / 3_600_000);
      const m = Math.floor((diff % 3_600_000) / 60_000);
      setCountdown(`${h}h ${m}m remaining`);
    };
    tick();
    const id = setInterval(tick, 60_000);
    return () => clearInterval(id);
  }, [data?.expiresAt]);

  const handleAccept = () => {
    if (!token) return;
    acceptMutation.mutate(
      { token, password },
      {
        onSuccess: () => navigate('/welcome'),
      }
    );
  };

  if (isLoading) {
    return (
      <AuthShell title="Loading invitation" subtitle="Checking your invite...">
        <div className="flex justify-center py-8">
          <Loader2 className="h-6 w-6 text-[#4583C4] animate-spin" />
        </div>
      </AuthShell>
    );
  }

  if (isError || !data || data.expired) {
    return (
      <AuthShell
        title="Invitation unavailable"
        subtitle="This invitation is invalid or has expired."
      >
        <motion.div
          initial={{ opacity: 0, y: 8 }}
          animate={{ opacity: 1, y: 0 }}
          className="flex flex-col items-center gap-4 text-center py-4"
        >
          <div className="flex h-16 w-16 items-center justify-center rounded-2xl bg-red-500/10 border border-red-400/25">
            <AlertCircle className="h-8 w-8 text-red-300" />
          </div>
          <p className="text-sm text-white/65">
            Ask the person who invited you to send a fresh invitation.
          </p>
          <AuthButton variant="outline" onClick={() => navigate('/')}>
            Back to sign in
          </AuthButton>
        </motion.div>
      </AuthShell>
    );
  }

  return (
    <AuthShell
      title="You're invited"
      subtitle={`Join ${data.organization} on rtylr.`}
    >
      <form
        onSubmit={(e) => {
          e.preventDefault();
          handleAccept();
        }}
        className="flex flex-col gap-5"
      >
        <div className="rounded-xl border border-white/10 bg-white/[0.03] p-4">
          <p className="text-xs text-white/50">
            You were invited by{' '}
            <span className="text-white/85 font-medium">{data.invitedBy}</span>{' '}
            as <span className="text-white/85 font-medium">{data.role}</span>.
          </p>
          <p className="mt-1 text-xs text-white/50">
            Email: <span className="text-white/85">{data.email}</span>
          </p>
          <p className="mt-2 inline-flex items-center gap-1 text-[11px] text-amber-200">
            <Clock className="h-3 w-3" />
            {countdown}
          </p>
        </div>

        <div className="flex flex-col gap-2">
          <AuthInput
            id="invite-password"
            type="password"
            label="Set a password"
            icon={User}
            placeholder="Create a strong password"
            autoComplete="new-password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
          />
          <PasswordStrengthMeter password={password} />
        </div>

        <TermsCheckbox checked={agree} onChange={setAgree} />

        <AuthButton
          type="submit"
          loading={acceptMutation.isPending}
          disabled={!agree || scorePassword(password) < 2}
        >
          <span className="inline-flex items-center justify-center gap-2">
            <Check className="h-[18px] w-[18px]" />
            Accept and join
          </span>
        </AuthButton>
      </form>
    </AuthShell>
  );
};

export default InvitationAccept;
