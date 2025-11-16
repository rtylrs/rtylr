import React, { useEffect, useState } from 'react';
import { Activity } from 'react';
import { useNavigate } from 'react-router-dom';
import notify from '../lib/notify/notify';
import { useUserStore } from '../stores/userStore';
import { useRegisterRestaurant, useGoogleAuth } from '../hooks/useAuth';
import { GoogleAuthProvider, signInWithPopup, getIdToken } from 'firebase/auth';
import { auth } from '../firebase';
import { FaGoogle } from 'react-icons/fa';
import { Button, Input, HelmetContainer, APP_URLS } from '@voxire/shared';
import { useGoogleSignupVisible } from '../hooks/useFeatures';

const Register: React.FC = () => {
  const navigate = useNavigate();
  const provider = new GoogleAuthProvider();
  const [email, setEmail] = useState<string>('');
  const [password, setPassword] = useState<string>('');
  const [confirmPassword, setConfirmPassword] = useState<string>('');
  const { isAuthenticated, loading, error, role, clearError } = useUserStore();
  const registerMutation = useRegisterRestaurant();
  const googleAuthMutation = useGoogleAuth();
  const showGoogle: boolean = useGoogleSignupVisible();

  const handleRegister = () => {
    if (password === confirmPassword) {
      registerMutation.mutate({
        email,
        password,
        name: email.split('@')[0], // Use email prefix as name for now
        referred: '',
      });
    } else {
      notify.error('Password does not match');
    }
  };

  const handleRegisterWithGoogle = async () => {
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
    // Clear any existing errors when component mounts
    clearError();
  }, [clearError]);

  useEffect(() => {
    if (isAuthenticated) {
      if (role === 'business') {
        navigate('/auth/questionnaire');
      } else {
        navigate('/');
      }
    }
  }, [isAuthenticated, navigate]);

  return (
    <React.Fragment>
      <HelmetContainer title="Register | Rtylr" />
      <div className="auth-container">
        <div className="auth-form-container">
          <div className="auth-form">
            <h2>Welcome to Rtylr.</h2>
            <p>Sign up to start managing everything in one place.</p>
            <form
              onSubmit={(e) => {
                e.preventDefault();
                handleRegister();
              }}
            >
              <Input
                id="register-email"
                type="email"
                label="Email"
                placeholder="Enter your email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
                style={{ marginBottom: 20 }}
              />
              <Input
                id="register-password"
                type="password"
                label="Password"
                placeholder="Enter your password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
                style={{ marginBottom: 20 }}
              />
              <Input
                id="register-confirm-password"
                type="password"
                label="Confirm Password"
                placeholder="Confirm your password"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
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

              <Button
                variant="primary"
                size="md"
                fullWidth
                loading={loading}
                type="submit"
              >
                Sign Up
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
                  onClick={handleRegisterWithGoogle}
                  prefix={<FaGoogle />}
                >
                  Sign up with Google
                </Button>
              </Activity>
            </form>

            <div className="auth-footer" style={{ alignItems: 'flex-start' }}>
              <p>
                Already have an account?{' '}
                <span className="auth-link" onClick={() => navigate('/')}>
                  Sign in
                </span>
              </p>
            </div>
          </div>
        </div>

        <div className="auth-page-footer">
          <p>
            Need help signing up?{' '}
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

export default Register;
