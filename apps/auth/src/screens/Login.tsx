import React, { useState, useEffect, useMemo } from 'react';
import { useNavigate } from 'react-router-dom';
import { motion } from 'framer-motion';
import {
  Mail,
  Lock,
  AlertCircle,
  ArrowRight,
  Wand2,
  Fingerprint,
  Building2,
} from 'lucide-react';
import { GoogleAuthProvider, signInWithPopup, getIdToken } from 'firebase/auth';
import { FaGoogle } from 'react-icons/fa';

import { useUserStore } from '../stores/userStore';
import { useLogin, useGoogleAuth } from '../hooks/useAuth';
import { auth } from '../firebase';
import notify from '../lib/notify/notify';
import { APP_URLS } from '@voxire/config';
import { useGoogleSignupVisible } from '../hooks/useFeatures';

import AuthShell from '../components/AuthShell';
import AuthInput from '../components/AuthInput';
import AuthButton from '../components/AuthButton';
import AuthDivider from '../components/AuthDivider';
import RememberMe from '../components/RememberMe';
import HelpLink from '../components/HelpLink';
import { getEmailError } from '../lib/validation';

const REDIRECT_APP_URLS: Record<string, string> = {
  hub:     window.location.hostname === 'localhost' ? 'http://localhost:5175' : 'https://hub.rtylr.com',
  pos:     window.location.hostname === 'localhost' ? 'http://localhost:5178' : 'https://pos.rtylr.com',
  erp:     window.location.hostname === 'localhost' ? 'http://localhost:5179' : 'https://erp.rtylr.com',
  hr:      window.location.hostname === 'localhost' ? 'http://localhost:5180' : 'https://hr.rtylr.com',
  crm:     window.location.hostname === 'localhost' ? 'http://localhost:5181' : 'https://crm.rtylr.com',
  finance: window.location.hostname === 'localhost' ? 'http://localhost:5182' : 'https://finance.rtylr.com',
  flow:    window.location.hostname === 'localhost' ? 'http://localhost:5183' : 'https://flow.rtylr.com',
};

const REDIRECT_APP_NAMES: Record<string, string> = {
  hub: 'Hub',
  pos: 'POS',
  erp: 'ERP',
  hr: 'HR',
  crm: 'CRM',
  finance: 'Finance',
  flow: 'Flow',
};

const fieldStagger = {
  hidden: { opacity: 0, y: 12 },
  show: (i: number) => ({
    opacity: 1,
    y: 0,
    transition: { delay: 0.2 + i * 0.06, duration: 0.5, ease: [0.16, 1, 0.3, 1] },
  }),
};

const Login: React.FC = () => {
  const navigate = useNavigate();
  const provider = new GoogleAuthProvider();
  const { isAuthenticated, loading, error, clearError } = useUserStore();
  const loginMutation = useLogin();
  const googleAuthMutation = useGoogleAuth();

  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [rememberMe, setRememberMe] = useState(true);
  const showGoogle: boolean = useGoogleSignupVisible();
  const emailError = getEmailError(email);

  const redirectParam = useMemo(() => {
    const params = new URLSearchParams(window.location.search);
    return params.get('redirect');
  }, []);

  const redirectAppName = redirectParam ? (REDIRECT_APP_NAMES[redirectParam] ?? null) : null;

  const handleSubmit = () => loginMutation.mutate({ email, password });

  const handleGoogle = async () => {
    try {
      const result = await signInWithPopup(auth, provider);
      const idToken = await getIdToken(result.user);
      googleAuthMutation.mutate({ idToken });
    } catch (err) {
      notify.error((err as any).message);
    }
  };

  useEffect(() => { clearError(); }, [clearError]);

  useEffect(() => {
    if (isAuthenticated) {
      const destination =
        (redirectParam && REDIRECT_APP_URLS[redirectParam])
          ? REDIRECT_APP_URLS[redirectParam]
          : REDIRECT_APP_URLS.hub;
      window.location.href = destination;
    }
  }, [isAuthenticated, redirectParam]);

  const errorMessage =
    error && (typeof error === 'string' ? error : (error as any)?.message || 'An error occurred');

  return (
    <AuthShell
      title="Welcome back."
      subtitle="Sign in to continue to your rtylr workspace."
      footer={
        <div className="space-y-3 text-center">
          {redirectAppName && (
            <p className="text-xs text-white/45">
              You'll be redirected to{' '}
              <span className="text-white/80 font-medium">{redirectAppName}</span> after sign-in.
            </p>
          )}
          <p className="text-sm text-white/55">
            Don't have an account?{' '}
            <button
              type="button"
              className="text-[#64C5B9] font-medium hover:underline underline-offset-4"
              onClick={() => navigate('/register')}
            >
              Sign up
            </button>
          </p>
          <div className="flex justify-center">
            <HelpLink />
          </div>
          <p className="text-[11px] text-white/35">
            By signing in you agree to our{' '}
            <a
              href={`${APP_URLS.landing}/legal/privacy-policy`}
              target="_blank"
              rel="noopener noreferrer"
              className="text-white/55 hover:text-white underline underline-offset-2"
            >
              Privacy Policy
            </a>
            .
          </p>
        </div>
      }
    >
      <form
        onSubmit={(e) => { e.preventDefault(); handleSubmit(); }}
        className="flex flex-col gap-5"
      >
        <motion.div custom={0} initial="hidden" animate="show" variants={fieldStagger}>
          <AuthInput
            id="login-email"
            type="email"
            label="Email"
            icon={Mail}
            placeholder="you@company.com"
            autoComplete="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            error={emailError}
            autoFocus
            required
          />
        </motion.div>

        <motion.div custom={1} initial="hidden" animate="show" variants={fieldStagger}>
          <AuthInput
            id="login-password"
            type="password"
            label="Password"
            icon={Lock}
            placeholder="Enter your password"
            autoComplete="current-password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
          />
        </motion.div>

        <motion.div
          custom={2}
          initial="hidden"
          animate="show"
          variants={fieldStagger}
          className="flex items-center justify-between -mt-1"
        >
          <RememberMe checked={rememberMe} onChange={setRememberMe} />
          <button
            type="button"
            className="text-xs text-white/55 hover:text-[#64C5B9] transition-colors"
            onClick={() => navigate('/password-reset')}
          >
            Forgot password?
          </button>
        </motion.div>

        {errorMessage && (
          <motion.div
            initial={{ opacity: 0, y: -6 }}
            animate={{ opacity: 1, y: 0 }}
            className="flex items-start gap-2 rounded-lg border border-red-400/25 bg-red-500/10 px-3.5 py-2.5"
            role="alert"
          >
            <AlertCircle className="h-4 w-4 text-red-300 mt-0.5 shrink-0" />
            <p className="text-xs text-red-200 leading-relaxed">{errorMessage}</p>
          </motion.div>
        )}

        <motion.div custom={3} initial="hidden" animate="show" variants={fieldStagger}>
          <AuthButton type="submit" loading={loading}>
            <span className="inline-flex items-center justify-center gap-2">
              Sign in
              <ArrowRight className="h-[18px] w-[18px]" />
            </span>
          </AuthButton>
        </motion.div>

        <motion.div
          custom={4}
          initial="hidden"
          animate="show"
          variants={fieldStagger}
          className="space-y-3"
        >
          <AuthDivider />

          {showGoogle && (
            <AuthButton type="button" variant="outline" onClick={handleGoogle}>
              <span className="inline-flex items-center justify-center gap-2.5">
                <FaGoogle className="text-white/85" />
                Continue with Google
              </span>
            </AuthButton>
          )}

          <div className="grid grid-cols-3 gap-2">
            <button
              type="button"
              onClick={() => navigate('/magic-link')}
              className="h-[46px] rounded-xl bg-white/[0.03] border border-white/12 hover:bg-white/[0.08] hover:border-white/20 text-white/80 hover:text-white transition-colors inline-flex items-center justify-center gap-1.5 text-xs font-medium"
              aria-label="Magic link"
            >
              <Wand2 className="h-[15px] w-[15px]" />
              Magic link
            </button>
            <button
              type="button"
              onClick={() => navigate('/sso')}
              className="h-[46px] rounded-xl bg-white/[0.03] border border-white/12 hover:bg-white/[0.08] hover:border-white/20 text-white/80 hover:text-white transition-colors inline-flex items-center justify-center gap-1.5 text-xs font-medium"
              aria-label="SSO"
            >
              <Building2 className="h-[15px] w-[15px]" />
              SSO
            </button>
            <button
              type="button"
              onClick={() => notify.info('Biometric login coming soon')}
              className="h-[46px] rounded-xl bg-white/[0.03] border border-white/12 hover:bg-white/[0.08] hover:border-white/20 text-white/80 hover:text-white transition-colors inline-flex items-center justify-center gap-1.5 text-xs font-medium"
              aria-label="Biometric login"
            >
              <Fingerprint className="h-[15px] w-[15px]" />
              Biometric
            </button>
          </div>
        </motion.div>
      </form>
    </AuthShell>
  );
};

export default Login;
