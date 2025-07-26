import { useState } from 'react';
import { Container, Typography, TextField } from '@mui/material';
import { auth } from '../firebase';
import { Button } from '@voxire/shared';

function ForgotPassword() {
  const [email, setEmail] = useState('');
  const [message, setMessage] = useState('');
  const [requestSent, setRequestSent] = useState(false);

  const handleResetPassword = async () => {
    try {
      await auth.sendPasswordResetEmail(email);
      setRequestSent(true);
      setMessage('Password reset email sent. Check your inbox.');
    } catch (error) {
      setMessage(`Error: ${error.message}`);
    }
  };

  return (
    <Container
      maxWidth="sm"
      style={{
        minHeight: '50vh',
        paddingTop: '10vh',
      }}
    >
      <Typography variant="h5" gutterBottom>
        Forgot Password
      </Typography>
      <Typography variant="body2" color="textSecondary">
        Enter your email address, and we will send you a link to reset your
        password.
      </Typography>
      <TextField
        variant="outlined"
        margin="normal"
        fullWidth
        id="email"
        label="Email Address"
        name="email"
        autoComplete="email"
        autoFocus
        value={email}
        onChange={(e) => setEmail(e.target.value)}
      />
      <div
        style={{
          marginTop: '1rem',
          display: 'flex',
          justifyContent: 'flex-end',
        }}
      >
        <Button fullWidth onClick={handleResetPassword} disabled={requestSent}>
          Reset Password
        </Button>
      </div>
      {message && (
        <Typography
          variant="body2"
          color="textSecondary"
          style={{ marginTop: '1rem' }}
        >
          {message}
        </Typography>
      )}
    </Container>
  );
}

export default ForgotPassword;
