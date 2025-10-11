import React, { useState } from 'react';
import { auth } from '../firebase';
import { Button, Input, HelmetContainer } from '@voxire/shared';

const ResetPassword: React.FC = () => {
  const [email, setEmail] = useState('');
  const [message, setMessage] = useState('');
  const [requestSent, setRequestSent] = useState(false);

  const handleResetPassword = async () => {
    try {
      await auth.sendPasswordResetEmail(email);
      setRequestSent(true);
      setMessage('Password reset email sent. Check your inbox.');
    } catch (error) {
      setMessage(`Error: ${(error as any).message}`);
    }
  };

  return (
    <React.Fragment>
      <HelmetContainer title="Forgot Password | Rtylr" />
      <div className="auth-container">
        <div className="auth-form-container">
          <div className="auth-form">
            <h2>Forgot Password</h2>
            <p>
              Enter your email address, and we will send you a link to reset
              your password.
            </p>

            <form
              onSubmit={(e) => {
                e.preventDefault();
                handleResetPassword();
              }}
            >
              <Input
                id="reset-email"
                type="email"
                label="Email Address"
                placeholder="Enter your email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
                autoFocus
                style={{ marginBottom: 16 }}
              />

              {message && (
                <p
                  style={{
                    color: requestSent ? '#059669' : '#dc2626',
                    fontSize: 14,
                    marginBottom: 16,
                    textAlign: 'center',
                  }}
                >
                  {message}
                </p>
              )}

              <Button
                variant="primary"
                size="md"
                fullWidth
                type="submit"
                disabled={requestSent}
              >
                Reset Password
              </Button>
            </form>

            <div className="auth-footer">
              <p>
                Remember your password?{' '}
                <span
                  className="auth-link"
                  onClick={() => window.history.back()}
                >
                  Back to login
                </span>
              </p>
            </div>
          </div>
        </div>
      </div>
    </React.Fragment>
  );
};

export default ResetPassword;
