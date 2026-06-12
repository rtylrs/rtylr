import { useEffect, useState } from 'react';
import { useNavigate, useParams } from '@tanstack/react-router';
import { CheckCircle2, AlertCircle, Loader2 } from 'lucide-react';

import { unwrap } from '@voxire/api';
import { safe } from '../../shared/lib/api';
import Logo from '../../shared/components/Logo';
import { Button, HelmetContainer } from '@voxire/components';

const EmailVerification = () => {
  const navigate = useNavigate();
  const { token } = useParams({ strict: false });
  const [message, setMessage] = useState('');
  const [loading, setLoading] = useState(true);
  const [success, setSuccess] = useState(false);

  useEffect(() => {
    unwrap(safe.get<{ message?: string }>(`/auth/verify/${token}`))
      .then((r) => {
        setMessage(r.message || 'Your email has been verified.');
        setSuccess(true);
      })
      .catch(() => {
        setMessage('Email verification failed. Please try again.');
        setSuccess(false);
      })
      .finally(() => setLoading(false));
  }, [token]);

  return (
    <>
      <HelmetContainer title="Verify email · rtylr" />
      <div className="min-h-screen w-full flex items-center justify-center bg-white px-4">
        <div className="w-full max-w-sm text-center">
          {/* Logo */}
          <div className="flex justify-center mb-8">
            <Logo variant="dark" className="h-8 w-auto select-none" />
          </div>

          {loading ? (
            <>
              <Loader2 className="h-10 w-10 text-[#4583C4] mx-auto animate-spin" />
              <p className="mt-5 text-sm text-[#666]">Verifying your email...</p>
            </>
          ) : success ? (
            <>
              <div className="h-14 w-14 rounded-full bg-green-50 border border-green-200 flex items-center justify-center mx-auto">
                <CheckCircle2 className="h-7 w-7 text-green-600" />
              </div>
              <h3 className="mt-5 text-lg font-semibold text-[#111]">Email verified</h3>
              <p className="mt-2 text-sm text-[#666] leading-relaxed">{message}</p>
              <div className="mt-6">
                <Button
                  type="button"
                  variant="primary"
                  className="auth-submit-button"
                  fullWidth
                  height={40}
                  onClick={() => navigate({ to: '/' })}
                >
                  Continue to sign in
                </Button>
              </div>
            </>
          ) : (
            <>
              <div className="h-14 w-14 rounded-full bg-red-50 border border-red-200 flex items-center justify-center mx-auto">
                <AlertCircle className="h-7 w-7 text-red-500" />
              </div>
              <h3 className="mt-5 text-lg font-semibold text-[#111]">Verification failed</h3>
              <p className="mt-2 text-sm text-[#666] leading-relaxed">{message}</p>
              <div className="mt-6">
                <Button
                  type="button"
                  variant="outline"
                  fullWidth
                  height={40}
                  onClick={() => navigate({ to: '/' })}
                >
                  Back to sign in
                </Button>
              </div>
            </>
          )}
        </div>
      </div>
    </>
  );
};

export default EmailVerification;
