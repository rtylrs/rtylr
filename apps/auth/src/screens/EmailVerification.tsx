import { useEffect, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { motion } from 'framer-motion';
import { CheckCircle2, AlertCircle, Loader2 } from 'lucide-react';

import api from '../lib/api';
import AuthShell from '../components/AuthShell';
import AuthButton from '../components/AuthButton';

const EmailVerification = () => {
  const navigate = useNavigate();
  const { token } = useParams();
  const [message, setMessage] = useState('');
  const [loading, setLoading] = useState(true);
  const [success, setSuccess] = useState(false);

  useEffect(() => {
    api
      .get(`/auth/verify/${token}`)
      .then((r) => {
        setMessage(r.data.message || 'Your email has been verified.');
        setSuccess(true);
      })
      .catch(() => {
        setMessage('Email verification failed. Please try again.');
        setSuccess(false);
      })
      .finally(() => setLoading(false));
  }, [token]);

  return (
    <AuthShell
      title="Email verification"
      subtitle="We're confirming your email address."
      footer={
        !loading && (
          <p className="text-sm text-center text-white/55">
            <button
              type="button"
              onClick={() => navigate('/')}
              className="text-[#64C5B9] font-medium hover:underline underline-offset-4"
            >
              Continue to sign in
            </button>
          </p>
        )
      }
    >
      <motion.div
        initial={{ opacity: 0, y: 12 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5, delay: 0.15 }}
        className="rounded-2xl border border-white/10 bg-white/[0.03] backdrop-blur-sm p-6 sm:p-8 text-center"
      >
        {loading ? (
          <>
            <Loader2 className="h-10 w-10 text-[#64C5B9] mx-auto animate-spin-slow" />
            <p className="mt-5 text-sm text-white/70">Verifying your email…</p>
          </>
        ) : success ? (
          <>
            <div className="h-14 w-14 rounded-full bg-[#64C5B9]/15 border border-[#64C5B9]/30 flex items-center justify-center mx-auto">
              <CheckCircle2 className="h-7 w-7 text-[#64C5B9]" />
            </div>
            <h3 className="mt-5 text-lg font-semibold text-white">Email verified</h3>
            <p className="mt-2 text-sm text-white/65 leading-relaxed">{message}</p>
            <div className="mt-6">
              <AuthButton type="button" onClick={() => navigate('/')}>
                Continue to sign in
              </AuthButton>
            </div>
          </>
        ) : (
          <>
            <div className="h-14 w-14 rounded-full bg-red-500/10 border border-red-400/30 flex items-center justify-center mx-auto">
              <AlertCircle className="h-7 w-7 text-red-300" />
            </div>
            <h3 className="mt-5 text-lg font-semibold text-white">
              Verification failed
            </h3>
            <p className="mt-2 text-sm text-white/65 leading-relaxed">{message}</p>
            <div className="mt-6">
              <AuthButton
                type="button"
                variant="outline"
                onClick={() => navigate('/')}
              >
                Back to sign in
              </AuthButton>
            </div>
          </>
        )}
      </motion.div>
    </AuthShell>
  );
};

export default EmailVerification;
