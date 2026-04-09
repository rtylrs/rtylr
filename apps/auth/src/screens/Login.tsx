import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useUserStore } from '../stores/userStore';
import { useLogin, useGoogleAuth } from '../hooks/useAuth';
import { GoogleAuthProvider, signInWithPopup, getIdToken } from 'firebase/auth';
import { auth } from '../firebase';
import notify from '../lib/notify/notify';
import { FaGoogle } from 'react-icons/fa';
import { Button, Input } from '@voxire/ui';
import { APP_URLS } from '@voxire/config';
import { useGoogleSignupVisible } from '../hooks/useFeatures';

const Login: React.FC = () => {
  const navigate = useNavigate();
  const provider = new GoogleAuthProvider();
  const { isAuthenticated, loading, error, role, clearError } = useUserStore();
  const loginMutation = useLogin();
  const googleAuthMutation = useGoogleAuth();

  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const showGoogle: boolean = useGoogleSignupVisible();

  const handleSubmit = () => loginMutation.mutate({ email, password });

  const handleGoogle = async () => {
    try {
      const result = await signInWithPopup(auth, provider);
      const idToken = await getIdToken(result.user);
      googleAuthMutation.mutate({ idToken });
    } catch (error) {
      notify.error((error as any).message);
    }
  };

  useEffect(() => { clearError(); }, [clearError]);

  useEffect(() => {
    if (isAuthenticated) {
      navigate(role === 'business' ? '/app/business' : '/');
    }
  }, [isAuthenticated, navigate, role]);

  return (
    <div className="min-h-screen flex items-center justify-center bg-[#f8f8fb] p-6">
      <div className="w-full max-w-md">
        {/* Card */}
        <div className="bg-white rounded-2xl shadow-card border border-black/5 p-8">
          <div className="mb-8">
            <h2 className="text-2xl font-bold text-primary mb-1.5">Welcome back.</h2>
            <p className="text-sm text-muted">Log in to your account below.</p>
          </div>

          <form onSubmit={(e) => { e.preventDefault(); handleSubmit(); }} className="flex flex-col gap-5">
            <Input
              id="login-email"
              type="email"
              label="Email"
              placeholder="Enter your email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
            />
            <Input
              id="login-password"
              type="password"
              label="Password"
              placeholder="Enter your password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
            />

            {error && (
              <p className="text-sm text-destructive text-center">
                {typeof error === 'string' ? error : (error as any)?.message || 'An error occurred'}
              </p>
            )}

            <button
              type="button"
              className="text-sm text-accent hover:underline self-start -mt-2"
              onClick={() => navigate('/password-reset')}
            >
              Forgot password?
            </button>

            <Button variant="primary" size="md" fullWidth loading={loading} type="submit">
              Sign In
            </Button>

            {showGoogle && (
              <>
                <div className="flex items-center gap-3 my-1">
                  <div className="flex-1 h-px bg-black/8" />
                  <span className="text-xs text-muted">or</span>
                  <div className="flex-1 h-px bg-black/8" />
                </div>
                <Button variant="outline" size="md" fullWidth type="button" onClick={handleGoogle} prefix={<FaGoogle />}>
                  Sign in with Google
                </Button>
              </>
            )}
          </form>

          <p className="mt-6 text-sm text-center text-muted">
            Don't have an account?{' '}
            <button className="text-accent hover:underline font-medium" onClick={() => navigate('/register')}>
              Sign up
            </button>
          </p>
        </div>

        {/* Footer */}
        <div className="mt-6 text-center space-y-2">
          <p className="text-xs text-muted">
            Need help?{' '}
            <button className="text-accent hover:underline" onClick={() => navigate('/password-reset')}>
              Reset your password
            </button>
          </p>
          <p className="text-xs text-muted/70">
            By using Rtylr, you agree to our{' '}
            <a href={`${APP_URLS.landing}/legal/privacy-policy`} target="_blank" rel="noopener noreferrer" className="text-accent hover:underline">
              Privacy Policy
            </a>
            .
          </p>
        </div>
      </div>
    </div>
  );
};

export default Login;
