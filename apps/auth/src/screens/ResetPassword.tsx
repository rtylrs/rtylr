import React, { useState } from 'react';
import { auth } from '../firebase';
import { Button, Input } from '@voxire/ui';

const ResetPassword: React.FC = () => {
  const [email, setEmail] = useState('');
  const [message, setMessage] = useState('');
  const [requestSent, setRequestSent] = useState(false);

  const handleReset = async () => {
    try {
      await auth.sendPasswordResetEmail(email);
      setRequestSent(true);
      setMessage('Password reset email sent. Check your inbox.');
    } catch (error) {
      setMessage(`Error: ${(error as any).message}`);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-[#f8f8fb] p-6">
      <div className="w-full max-w-md">
        <div className="bg-white rounded-2xl shadow-card border border-black/5 p-8">
          <div className="mb-8">
            <h2 className="text-2xl font-bold text-primary mb-1.5">Forgot Password</h2>
            <p className="text-sm text-muted">We'll send you a link to reset your password.</p>
          </div>

          <form onSubmit={(e) => { e.preventDefault(); handleReset(); }} className="flex flex-col gap-5">
            <Input id="reset-email" type="email" label="Email Address" placeholder="Enter your email" value={email} onChange={(e) => setEmail(e.target.value)} required autoFocus />

            {message && <p className={`text-sm text-center ${requestSent ? 'text-success' : 'text-destructive'}`}>{message}</p>}

            <Button variant="primary" size="md" fullWidth type="submit" disabled={requestSent}>Reset Password</Button>
          </form>

          <p className="mt-6 text-sm text-center text-muted">
            <button className="text-accent hover:underline" onClick={() => window.history.back()}>Back to login</button>
          </p>
        </div>
      </div>
    </div>
  );
};

export default ResetPassword;
