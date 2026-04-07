import { useState } from "react";
import { auth } from "../../firebase";
import { Button, Input } from '@voxire/components';
import styles from './Reset.module.scss';

function ForgotPassword() {
  const [email, setEmail] = useState("");
  const [message, setMessage] = useState("");
  const [requestSent, setRequestSent] = useState(false);

  const handleResetPassword = async () => {
    try {
      await auth.sendPasswordResetEmail(email);
      setRequestSent(true);
      setMessage("Password reset email sent. Check your inbox.");
    } catch (error) {
      setMessage(`Error: ${error.message}`);
    }
  };

  return (
    <div className={styles.container}>
      <h5 className={styles.title}>Forgot Password</h5>
      <p className={styles.subtitle}>
        Enter your email address, and we will send you a link to reset your password.
      </p>
      <Input
        type="email"
        id="email"
        name="email"
        label="Email Address"
        hasLabel
        autoComplete="email"
        autoFocus
        value={email}
        onChange={(e) => setEmail(e.target.value)}
      />
      <div className={styles.actions}>
        <Button
          fullWidth
          onClick={handleResetPassword}
          disabled={requestSent}
        >
          Reset Password
        </Button>
      </div>
      {message && <p className={styles.message}>{message}</p>}
    </div>
  );
}

export default ForgotPassword;
