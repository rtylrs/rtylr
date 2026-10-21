import { useEffect, useRef, useState } from 'react';
import { useNavigate, useParams } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { CheckCircle2, AlertCircle, Loader2 } from 'lucide-react';

import Logo from '../../shared/components/Logo';
import { Button, HelmetContainer } from '@voxire/components';
import { getOneKitErrorMessage } from '@voxire/api';
import { useEmailVerification } from './useEmailVerification';

const EmailVerification = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { token } = useParams({ strict: false });
  const [message, setMessage] = useState('');
  const [loading, setLoading] = useState(true);
  const [success, setSuccess] = useState(false);
  const verifyMutation = useEmailVerification();
  const done = useRef(false);

  useEffect(() => {
    if (done.current) return;
    done.current = true;

    if (!token) {
      setMessage(t('auth.verify.tokenMissing'));
      setSuccess(false);
      setLoading(false);
      return;
    }

    verifyMutation
      .mutateAsync({ token })
      .then((r) => {
        if (r?.verified === false) {
          setMessage(r.message || t('auth.verify.defaultFailure'));
          setSuccess(false);
        } else {
          setMessage(r.message || t('auth.verify.defaultSuccess'));
          setSuccess(true);
        }
      })
      .catch((error: unknown) => {
        // The server distinguishes a code that is too old (request a new one)
        // from a wrong one; its own wording is what the user needs here, and
        // this component cannot translate it.
        setMessage(getOneKitErrorMessage(error, t('auth.verify.defaultFailure')));
        setSuccess(false);
      })
      .finally(() => setLoading(false));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [token]);

  return (
    <>
      <HelmetContainer title={t('auth.verify.metaTitle')} />
      <div className="min-h-screen w-full flex items-center justify-center bg-white px-4">
        <div className="w-full max-w-sm text-center">
          {/* Logo */}
          <div className="flex justify-center mb-8">
            <Logo variant="dark" className="h-8 w-auto select-none" />
          </div>

          {loading ? (
            <>
              <Loader2 className="h-10 w-10 text-[#111]/70 mx-auto animate-spin" />
              <p className="mt-5 text-sm text-[#666]">{t('auth.verify.verifying')}</p>
            </>
          ) : success ? (
            <>
              <div className="h-14 w-14 rounded-full bg-green-50 border border-green-200 flex items-center justify-center mx-auto">
                <CheckCircle2 className="h-7 w-7 text-green-600" />
              </div>
              <h3 className="mt-5 text-lg font-semibold text-[#111]">
                {t('auth.verify.verifiedTitle')}
              </h3>
              {/* `message` is whatever the API returned, so it arrives in
                  whatever language the server chose: this screen cannot
                  translate it. Only the fallback below is ours. Worth fixing at
                  the source rather than papering over here. */}
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
                  {t('auth.verify.continueToSignIn')}
                </Button>
              </div>
            </>
          ) : (
            <>
              <div className="h-14 w-14 rounded-full bg-red-50 border border-red-200 flex items-center justify-center mx-auto">
                <AlertCircle className="h-7 w-7 text-red-500" />
              </div>
              <h3 className="mt-5 text-lg font-semibold text-[#111]">
                {t('auth.verify.failedTitle')}
              </h3>
              <p className="mt-2 text-sm text-[#666] leading-relaxed">{message}</p>
              <div className="mt-6">
                <Button
                  type="button"
                  variant="outline"
                  fullWidth
                  height={40}
                  onClick={() => navigate({ to: '/' })}
                >
                  {t('auth.reset.backToSignIn')}
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
