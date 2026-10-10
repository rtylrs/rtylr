import React, { useEffect, useState } from 'react';
import { useNavigate, useRouterState } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { Mail, CheckCircle2, AlertCircle, Loader2 } from 'lucide-react';

import Logo from '../../shared/components/Logo';
import { Button } from '@voxire/components';
import { getOneKitErrorMessage } from '@voxire/api';
import { useEmailVerification, useResendVerificationEmail } from './useEmailVerification';
import {
  isLegacyVerificationCode,
  isVerificationToken,
  normalizeVerificationToken,
  VERIFICATION_TOKEN_MAX_LENGTH,
} from './verificationToken';
import notify from '../../shared/lib/notify/notify';
import { authInputClass } from '../../shared/components/AuthField';

type Status = 'loading' | 'success' | 'error' | 'awaitingCode';

const VerifyEmail: React.FC = () => {
  const navigate = useNavigate();
  const { t } = useTranslation();
  const searchStr = useRouterState({ select: (s) => s.location.searchStr });
  const params = new URLSearchParams(searchStr);
  const verifyMutation = useEmailVerification();
  const resendMutation = useResendVerificationEmail();
  const [status, setStatus] = useState<Status>('loading');
  const [verificationCode, setVerificationCode] = useState('');
  // Whatever the server said about the last attempt, so its "request a new one"
  // outcome is not overwritten by the generic "invalid" copy.
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const hasLinkToken = Boolean(params.get('token'));

  useEffect(() => {
    // A link token is forwarded as it arrived rather than shape-checked: the
    // value came from our own email, and the server already answers every
    // unusable shape distinctly.
    const token = params.get('token')?.trim();
    if (!token) {
      // Not an error: arriving here without a token is the normal typed-entry
      // path. Showing a red alert reading "the verification code is invalid"
      // before the user has typed anything told them their code had failed when
      // they had not used it yet.
      setStatus('awaitingCode');
      return;
    }
    verifyMutation.mutate(
      { token },
      {
        onSuccess: (data) => {
          if (data?.verified === false) {
            setStatus('error');
          } else {
            setStatus('success');
            try {
              window.history.replaceState({}, document.title, window.location.pathname);
            } catch {
              // ignore environment restriction
            }
          }
        },
        onError: (error) => {
          setErrorMessage(getOneKitErrorMessage(error, t('auth.verify.failed')));
          setStatus('error');
        },
      }
    );
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const [resendEmail, setResendEmail] = useState(params.get('email') ?? '');

  const handleResend = (targetEmail?: string): void => {
    const emailToUse = targetEmail || resendEmail || params.get('email') || '';
    if (!emailToUse) {
      notify.error(t('auth.verify.emailRequired'));
      return;
    }
    resendMutation.mutate(
      { email: emailToUse },
      {
        onSuccess: () => notify.success(t('auth.verify.resent')),
        onError: () => notify.error(t('auth.verify.resendFailed')),
      }
    );
  };

  const handleVerifyCode = (): void => {
    // The field holds the whole token, id half included: that is what onion
    // needs to recompute the account-bound digest, and a code stripped of it
    // verifies nothing. A legacy six-digit code is sent rather than refused so
    // the server can say it is too old instead of the user retrying forever.
    const token = normalizeVerificationToken(verificationCode);
    if (!isVerificationToken(token) && !isLegacyVerificationCode(token)) {
      setErrorMessage(t('auth.verify.codeInvalid'));
      setStatus('error');
      return;
    }
    setErrorMessage(null);
    setStatus('loading');
    verifyMutation.mutate(
      { token },
      {
        onSuccess: (data) => setStatus(data?.verified === false ? 'error' : 'success'),
        onError: (error) => {
          setErrorMessage(getOneKitErrorMessage(error, t('auth.verify.codeInvalid')));
          setStatus('error');
        },
      }
    );
  };

  return (
    <div className="min-h-screen w-full flex items-center justify-center bg-white px-4">
      <div className="w-full max-w-sm text-center">
        {/* Logo */}
        <div className="flex justify-center mb-8">
          <Logo variant="dark" className="h-8 w-auto select-none" />
        </div>

        {status === 'loading' && (
          <>
            <div className="flex h-16 w-16 items-center justify-center rounded-2xl bg-blue-50 border border-blue-200 mx-auto">
              <Loader2 className="h-8 w-8 text-[#111]/70 animate-spin" />
            </div>
            <p className="mt-4 text-sm text-[#666]">{t('auth.verify.verifying')}</p>
          </>
        )}
        {status === 'success' && (
          <>
            <div className="flex h-16 w-16 items-center justify-center rounded-2xl bg-green-50 border border-green-200 mx-auto">
              <CheckCircle2 className="h-8 w-8 text-green-600" />
            </div>
            <p className="mt-4 text-base font-semibold text-[#111]">{t('auth.verify.success')}</p>
            <div className="mt-6">
              <Button
                type="button"
                variant="primary"
                fullWidth
                height={40}
                onClick={() => navigate({ to: '/' })}
              >
                {t('auth.verify.continueToSignIn')}
              </Button>
            </div>
          </>
        )}
        {(status === 'error' || status === 'awaitingCode') && (
          <>
            {status === 'awaitingCode' ? (
              <>
                <div className="flex h-16 w-16 items-center justify-center rounded-2xl bg-blue-50 border border-blue-200 mx-auto">
                  <Mail className="h-8 w-8 text-[#111]/70" />
                </div>
                <p className="mt-4 text-sm text-[#666]">{t('auth.verify.enterCodePrompt')}</p>
              </>
            ) : (
              <>
                <div className="flex h-16 w-16 items-center justify-center rounded-2xl bg-red-50 border border-red-200 mx-auto">
                  <AlertCircle className="h-8 w-8 text-red-500" />
                </div>
                <p className="mt-4 text-sm text-red-600">
                  {errorMessage ??
                    (hasLinkToken ? t('auth.verify.failed') : t('auth.verify.codeInvalid'))}
                </p>
              </>
            )}
            {/* A link token that has already failed leaves the form behind:
                hiding it there stranded the user with no way to try a code
                they had just been asked to request a new one for. */}
            {(!hasLinkToken || status === 'error') && (
              <form
                onSubmit={(event) => {
                  event.preventDefault();
                  handleVerifyCode();
                }}
                className="mt-6 space-y-3"
              >
                <label htmlFor="verification-code" className="sr-only">
                  {t('auth.verify.codeLabel')}
                </label>
                <input
                  id="verification-code"
                  type="text"
                  // Not "numeric": the token carries a "." separator, and the
                  // numeric keypad on iOS has no key for one.
                  inputMode="text"
                  autoComplete="one-time-code"
                  maxLength={VERIFICATION_TOKEN_MAX_LENGTH}
                  placeholder={t('auth.verify.codePlaceholder')}
                  value={verificationCode}
                  onChange={(event) =>
                    setVerificationCode(normalizeVerificationToken(event.target.value))
                  }
                  className={authInputClass}
                />
                <Button type="submit" variant="primary" fullWidth height={44}>
                  {t('auth.verify.verifyCode')}
                </Button>
              </form>
            )}
            <form
              onSubmit={(e) => {
                e.preventDefault();
                handleResend();
              }}
              className="mt-6 space-y-3"
            >
              <input
                type="email"
                placeholder={t('auth.common.email')}
                value={resendEmail}
                onChange={(e) => setResendEmail(e.target.value)}
                required
                className={authInputClass}
              />
              <Button
                type="submit"
                variant="outline"
                fullWidth
                height={40}
                loading={resendMutation.isPending}
              >
                <span className="inline-flex items-center gap-2">
                  <Mail className="h-4 w-4" />
                  {t('auth.verify.resend')}
                </span>
              </Button>
            </form>
          </>
        )}

        <div className="mt-6">
          <button
            type="button"
            onClick={() => navigate({ to: '/' })}
            className="text-sm text-[#111] font-medium underline decoration-black/25 hover:decoration-black/70 underline-offset-2"
          >
            {t('auth.reset.backToSignIn')}
          </button>
        </div>
      </div>
    </div>
  );
};

export default VerifyEmail;
