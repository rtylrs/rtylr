import React, { useState, useEffect, useMemo } from 'react';
import { useNavigate } from 'react-router-dom';
import { AlertCircle } from 'lucide-react';
import { GoogleAuthProvider, signInWithPopup, getIdToken } from 'firebase/auth';
import { FaGoogle } from 'react-icons/fa';

import { useUserStore } from '../stores/userStore';
import { useLogin, useGoogleAuth } from '../hooks/useAuth';
import { auth } from '../firebase';
import notify from '../lib/notify/notify';
import { Button } from '@voxire/components';
import AuthShell from '../components/AuthShell';

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

const Login: React.FC = () => {
  const navigate = useNavigate();
  const provider = new GoogleAuthProvider();
  const { isAuthenticated, loading, error, clearError } = useUserStore();
  const loginMutation = useLogin();
  const googleAuthMutation = useGoogleAuth();

  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
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
    // Only redirect if both the store says authenticated AND a token cookie exists.
    // This prevents a loop when the user logs out (cookie cleared but store still persisted).
    const tokenCookie = document.cookie.split(';').some((c) => c.trim().startsWith('token='));
    if (isAuthenticated && tokenCookie) {
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
      subtitle="Sign in to your account."
      footer={
        <div className="text-center space-y-3">
          {redirectAppName && (
            <p className="text-xs text-[#999]">
              You will be redirected to{' '}
              <span className="text-[#111] font-medium">{redirectAppName}</span> after sign-in.
            </p>
          )}
          <p className="text-sm text-[#666]">
            Don't have an account?{' '}
            <button
              type="button"
              className="text-[#4583C4] font-medium hover:underline underline-offset-2"
              onClick={() => navigate('/register')}
            >
              Sign up
            </button>
          </p>
        </div>
      }
    >
        <form
          onSubmit={(e) => { e.preventDefault(); handleSubmit(); }}
          className="flex flex-col gap-4"
        >
          {/* Email */}
          <div className="flex flex-col gap-1.5">
            <label htmlFor="login-email" className="text-sm font-medium text-[#111]">
              Email
            </label>
            <input
              id="login-email"
              type="email"
              placeholder="you@company.com"
              autoComplete="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              autoFocus
              required
              className="w-full h-10 rounded-lg border border-[#e5e5e5] bg-white px-3 text-sm text-[#111] placeholder:text-[#999] outline-none transition-colors focus:border-[#4583C4] focus:ring-2 focus:ring-[#4583C4]/20"
            />
            {emailError && <p className="text-xs text-red-500">{emailError}</p>}
          </div>

          {/* Password */}
          <div className="flex flex-col gap-1.5">
            <label htmlFor="login-password" className="text-sm font-medium text-[#111]">
              Password
            </label>
            <div className="relative">
              <input
                id="login-password"
                type={showPassword ? 'text' : 'password'}
                placeholder="Enter your password"
                autoComplete="current-password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
                className="w-full h-10 rounded-lg border border-[#e5e5e5] bg-white px-3 pr-10 text-sm text-[#111] placeholder:text-[#999] outline-none transition-colors focus:border-[#4583C4] focus:ring-2 focus:ring-[#4583C4]/20"
              />
              <button
                type="button"
                onClick={() => setShowPassword((v) => !v)}
                aria-label={showPassword ? 'Hide password' : 'Show password'}
                className="absolute right-2.5 top-1/2 -translate-y-1/2 text-[#999] hover:text-[#666] text-xs font-medium"
              >
                {showPassword ? 'Hide' : 'Show'}
              </button>
            </div>
          </div>

          {/* Forgot password */}
          <div className="flex justify-end -mt-1">
            <button
              type="button"
              className="text-xs text-[#4583C4] hover:underline underline-offset-2"
              onClick={() => navigate('/password-reset')}
            >
              Forgot password?
            </button>
          </div>

          {/* Error */}
          {errorMessage && (
            <div
              className="flex items-start gap-2 rounded-lg border border-red-200 bg-red-50 px-3 py-2.5"
              role="alert"
            >
              <AlertCircle className="h-4 w-4 text-red-500 mt-0.5 shrink-0" />
              <p className="text-xs text-red-600 leading-relaxed">{errorMessage}</p>
            </div>
          )}

          {/* Sign in button */}
          <Button
            type="submit"
            variant="primary"
            fullWidth
            loading={loading}
            height={40}
          >
            Sign in
          </Button>

          {/* Google sign in */}
          <Button
            type="button"
            variant="outline"
            fullWidth
            height={40}
            onClick={handleGoogle}
          >
            <span className="inline-flex items-center gap-2">
              <FaGoogle className="text-[#444]" />
              Continue with Google
            </span>
          </Button>
        </form>
    </AuthShell>
  );
};

export default Login;
