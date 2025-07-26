import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useDispatch, useSelector } from 'react-redux';
import {
  loginAsync,
  selectIsAuthenticated,
  selectLoading,
  selectError,
  selectRole,
  restaurantWithGoogleAsync,
  clearError,
} from '../redux/reducers/userSlice';
import { RootState, AppDispatch } from '../redux/store';
import { GoogleAuthProvider, signInWithPopup, getIdToken } from 'firebase/auth';
import { auth } from '../firebase';
import notify from '../lib/notify/notify';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { faGoogle } from '@fortawesome/free-brands-svg-icons';
import { Button } from '@voxire/shared';
import { HelmetContainer } from '@voxire/shared';

const Login: React.FC = () => {
  const navigate = useNavigate();
  const provider = new GoogleAuthProvider();
  const dispatch: AppDispatch = useDispatch();
  const isAuthenticated = useSelector((state: RootState) =>
    selectIsAuthenticated(state)
  );
  const loading = useSelector((state: RootState) => selectLoading(state));
  const error = useSelector((state: RootState) => selectError(state));
  const role = useSelector((state: RootState) => selectRole(state));

  const [email, setEmail] = useState<string>('');
  const [password, setPassword] = useState<string>('');

  const handleSubmit = () => {
    dispatch(loginAsync({ email, password }));
  };

  const handleGoogle = async () => {
    try {
      const result = await signInWithPopup(auth, provider);
      const user = result.user;
      const idToken = await getIdToken(user);

      dispatch(restaurantWithGoogleAsync({ idToken }));
    } catch (error) {
      notify.error((error as any).message);
    }
  };

  useEffect(() => {
    // Clear any existing errors when component mounts
    dispatch(clearError());
  }, [dispatch]);

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
              <label htmlFor="login-email" className="auth-label">
                Email
              </label>
              <input
                id="login-email"
                className="auth-input"
                type="email"
                placeholder="Enter your email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
              />
              <label htmlFor="login-password" className="auth-label">
                Password
              </label>
              <input
                id="login-password"
                className="auth-input"
                type="password"
                placeholder="Enter your password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
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
                  onClick={() => navigate('/auth/forgot-password')}
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
                  onClick={() => navigate('/auth/register')}
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
