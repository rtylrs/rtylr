import React, { useState } from 'react';
import { motion } from 'framer-motion';
import { User, CheckCircle2 } from 'lucide-react';

import AuthShell from '../components/AuthShell';
import AuthInput from '../components/AuthInput';
import SubmitButton from '../components/SubmitButton';
import BackLink from '../components/BackLink';

const ForgotUsername: React.FC = () => {
  const [phone, setPhone] = useState('');
  const [sent, setSent] = useState(false);
  const [loading, setLoading] = useState(false);

  const handleSubmit = () => {
    setLoading(true);
    setTimeout(() => {
      setLoading(false);
      setSent(true);
    }, 700);
  };

  return (
    <AuthShell
      title="Forgot your username?"
      subtitle="We'll send your username to the phone number on file."
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
          <p className="text-sm text-white/65">
            If that phone number is associated with an rtylr account, you'll
            receive an SMS with your username shortly.
          </p>
        </motion.div>
      ) : (
        <form
          onSubmit={(e) => {
            e.preventDefault();
            handleSubmit();
          }}
          className="flex flex-col gap-5"
        >
          <AuthInput
            id="forgot-username-phone"
            type="tel"
            label="Phone number"
            icon={User}
            placeholder="+971 50 000 0000"
            value={phone}
            onChange={(e) => setPhone(e.target.value)}
            required
            autoFocus
          />
          <SubmitButton
            label="Send via SMS"
            loading={loading}
            disabled={!phone}
          />
        </form>
      )}
    </AuthShell>
  );
};

export default ForgotUsername;
