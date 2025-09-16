import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useUserStore } from '../stores/userStore';
import { useLogin, useGoogleAuth } from '../hooks/useAuth';
import { GoogleAuthProvider, signInWithPopup, getIdToken } from 'firebase/auth';
import { auth } from '../firebase';
import notify from '../lib/notify/notify';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { faGoogle } from '@fortawesome/free-brands-svg-icons';
import { Button, Input } from '@voxire/shared';
import { HelmetContainer } from '@voxire/shared';

const Login: React.FC = () => {
  const navigate = useNavigate();
  const provider = new GoogleAuthProvider();
  const { isAuthenticated, loading, error, role, clearError } = useUserStore();
  const loginMutation = useLogin();
  const googleAuthMutation = useGoogleAuth();

  const [email, setEmail] = useState<string>('');
  const [password, setPassword] = useState<string>('');

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
      <HelmetContainer title="Login" />
      <div className="auth-container">
        <div className="auth-form-container">
          <div className="auth-form">
            <h2>Welcome back!</h2>
            <p>Please enter your details to sign in</p>
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
                style={{ marginBottom: 16 }}
              />
              <Input
                id="login-password"
                type="password"
                label="Password"
                placeholder="Enter your password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
                style={{ marginBottom: 16 }}
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
                Login
              </Button>

              <div className="auth-separator">
                <span>or</span>
              </div>

              <Button
                variant="outline"
                size="md"
                fullWidth
                type="button"
                onClick={handleGoogle}
                prefix={<FontAwesomeIcon icon={faGoogle} />}
              >
                Sign in with Google
              </Button>
            </form>

            <div className="auth-footer">
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
      </div>
    </React.Fragment>
  );
};

export default Login;
