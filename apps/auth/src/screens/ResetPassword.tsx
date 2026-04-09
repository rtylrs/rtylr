import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { motion } from 'framer-motion';
import {
  Mail,
  AlertCircle,
  ArrowLeft,
  ArrowRight,
  CheckCircle2,
} from 'lucide-react';

import { auth } from '../firebase';

import AuthShell from '../components/AuthShell';
import AuthInput from '../components/AuthInput';
import AuthButton from '../components/AuthButton';

const fieldStagger = {
  hidden: { opacity: 0, y: 12 },
  show: (i: number) => ({
    opacity: 1,
    y: 0,
    transition: { delay: 0.2 + i * 0.06, duration: 0.5, ease: [0.16, 1, 0.3, 1] },
  }),
};

const ResetPassword: React.FC = () => {
  const navigate = useNavigate();
  const [email, setEmail] = useState('');
  const [message, setMessage] = useState('');
  const [loading, setLoading] = useState(false);
  const [requestSent, setRequestSent] = useState(false);

  const handleReset = async () => {
    setLoading(true);
    try {
      await auth.sendPasswordResetEmail(email);
      setRequestSent(true);
      setMessage('Password reset email sent. Check your inbox.');
    } catch (error) {
      setMessage(`${(error as any).message}`);
    } finally {
      setLoading(false);
    }
  };

  return (
    <AuthShell
      title="Forgot password?"
      subtitle="Enter your email and we'll send you a secure reset link."
      footer={
        <p className="text-sm text-center text-white/55">
          <button
            type="button"
            onClick={() => navigate('/')}
            className="inline-flex items-center gap-1.5 text-[#64C5B9] font-medium hover:underline underline-offset-4"
          >
            <ArrowLeft className="h-4 w-4" />
            Back to sign in
          </button>
        </p>
      }
    >
      <form
        onSubmit={(e) => { e.preventDefault(); handleReset(); }}
        className="flex flex-col gap-5"
      >
        <motion.div custom={0} initial="hidden" animate="show" variants={fieldStagger}>
          <AuthInput
            id="reset-email"
            type="email"
            label="Email"
            icon={Mail}
            placeholder="you@company.com"
            autoComplete="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
            autoFocus
            disabled={requestSent}
          />
        </motion.div>

        {message && (
          <motion.div
            initial={{ opacity: 0, y: -6 }}
            animate={{ opacity: 1, y: 0 }}
            className={[
              'flex items-start gap-2 rounded-lg px-3.5 py-2.5 border',
              requestSent
                ? 'border-[#64C5B9]/30 bg-[#64C5B9]/10'
                : 'border-red-400/25 bg-red-500/10',
            ].join(' ')}
            role="alert"
          >
            {requestSent ? (
              <CheckCircle2 className="h-4 w-4 text-[#64C5B9] mt-0.5 shrink-0" />
            ) : (
              <AlertCircle className="h-4 w-4 text-red-300 mt-0.5 shrink-0" />
            )}
            <p
              className={[
                'text-xs leading-relaxed',
                requestSent ? 'text-[#c3ede7]' : 'text-red-200',
              ].join(' ')}
            >
              {message}
            </p>
          </motion.div>
        )}

        <motion.div custom={1} initial="hidden" animate="show" variants={fieldStagger}>
          <AuthButton type="submit" loading={loading} disabled={requestSent}>
            <span className="inline-flex items-center justify-center gap-2">
              {requestSent ? 'Email sent' : 'Send reset link'}
              {!requestSent && <ArrowRight className="h-[18px] w-[18px]" />}
            </span>
          </AuthButton>
        </motion.div>
      </form>
    </AuthShell>
  );
};

export default ResetPassword;
