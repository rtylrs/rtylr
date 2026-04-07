import React, { useState, useEffect } from 'react';
import { Activity } from 'react';
import { useNavigate } from 'react-router-dom';
import { useUserStore } from '../stores/userStore';
import { useLogin, useGoogleAuth } from '../hooks/useAuth';
import { GoogleAuthProvider, signInWithPopup, getIdToken } from 'firebase/auth';
import { auth } from '../firebase';
import notify from '../lib/notify/notify';
import { FaGoogle } from 'react-icons/fa';
import { Button, Input, HelmetContainer } from '@voxire/components';
import { APP_URLS } from '@voxire/config';
import { useGoogleSignupVisible } from '../hooks/useFeatures';

const Login: React.FC = () => {
  const navigate = useNavigate();
  const provider = new GoogleAuthProvider();
  const { isAuthenticated, loading, error, role, clearError } = useUserStore();
  const loginMutation = useLogin();
  const googleAuthMutation = useGoogleAuth();

  const [email, setEmail] = useState<string>('');
  const [password, setPassword] = useState<string>('');
  const showGoogle: boolean = useGoogleSignupVisible();

  const handleSubmit = () => {
    loginMutation.mutate({ email, password });
  };

  const handleGoogle = async () => {
    try {
      const result = await signInWithPopup(auth, provider);
      const user = result.user;
      const idToken = await getIdToken(user);

      googleAuthMutation.mutate({ idToken });
    } catch (error) {
      notify.error((error as any).message);
    }
  };

  useEffect(() => {
    clearError();
  }, [clearError]);

  useEffect(() => {
    if (isAuthenticated) {
      if (role === 'business') {
        navigate('/app/business');
      } else {
        navigate('/');
      }
    }
  }, [isAuthenticated, navigate, role]);

  return (
    <React.Fragment>
      <HelmetContainer title="Login | Rtylr" />
      <div className="auth-container">
        <div className="auth-form-container">
          <div className="auth-form">
            <h2>Welcome back.</h2>
            <p>Log in to your account below.</p>
            <form
              onSubmit={(e) => {
                e.preventDefault();
                handleSubmit();
              }}
            >
              <Input
                id="login-email"
                type="email"
                label="Email"
                placeholder="Enter your email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
                style={{ marginBottom: 20 }}
              />
              <Input
                id="login-password"
                type="password"
                label="Password"
                placeholder="Enter your password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
                style={{ marginBottom: 20 }}
              />

              {error && (
                <p
                  style={{
                    color: '#dc2626',
                    fontSize: 14,
                    marginBottom: 16,
                    textAlign: 'center',
                  }}
                >
                  {typeof error === 'string'
                    ? error
                    : (error as any)?.message || 'An error occurred'}
                </p>
              )}

              <div style={{ width: '100%', marginBottom: 16 }}>
                <span
                  className="auth-link"
                  onClick={() => navigate('/password-reset')}
                  style={{ fontSize: 14, cursor: 'pointer' }}
                >
                  Forgot password?
                </span>
              </div>

              <Button
                variant="primary"
                size="md"
                fullWidth
                loading={loading}
                type="submit"
              >
                Sign In
              </Button>

              <Activity mode={showGoogle ? 'visible' : 'hidden'}>
                <div className="auth-separator">
                  <span>or</span>
                </div>
                <Button
                  variant="outline"
                  size="md"
                  fullWidth
                  type="button"
                  onClick={handleGoogle}
                  prefix={<FaGoogle />}
                >
                  Sign in with Google
                </Button>
              </Activity>
            </form>

            <div className="auth-footer" style={{ alignItems: 'flex-start' }}>
              <p>
                Don't have an account?{' '}
                <span
                  className="auth-link"
                  onClick={() => navigate('/register')}
                >
                  Sign up
                </span>
              </p>
            </div>
          </div>
        </div>

        <div className="auth-page-footer">
          <p>
            Need help logging in?{' '}
            <span
              className="auth-link"
              onClick={() => navigate('/password-reset')}
            >
              Reset your password
            </span>
          </p>
          <p className="privacy-text">
            By using Rtylr, you acknowledge that we collect and use your
            personal information as described in our{' '}
            <a
              href={`${APP_URLS.landing}/legal/privacy-policy`}
              target="_blank"
              rel="noopener noreferrer"
              className="auth-link"
            >
              Privacy Policy
            </a>
            .
          </p>
        </div>
      </div>
    </React.Fragment>
  );
};

export default Login;
