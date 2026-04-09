import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import notify from '../lib/notify/notify';
import { useUserStore } from '../stores/userStore';
import { useRegisterRestaurant, useGoogleAuth } from '../hooks/useAuth';
import { GoogleAuthProvider, signInWithPopup, getIdToken } from 'firebase/auth';
import { auth } from '../firebase';
import { FaGoogle } from 'react-icons/fa';
import { Button, Input } from '@voxire/ui';
import { APP_URLS } from '@voxire/config';
import { useGoogleSignupVisible } from '../hooks/useFeatures';

const Register: React.FC = () => {
  const navigate = useNavigate();
  const provider = new GoogleAuthProvider();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const { isAuthenticated, loading, error, role, clearError } = useUserStore();
  const registerMutation = useRegisterRestaurant();
  const googleAuthMutation = useGoogleAuth();
  const showGoogle: boolean = useGoogleSignupVisible();

  const handleRegister = () => {
    if (password !== confirmPassword) { notify.error('Passwords do not match'); return; }
    registerMutation.mutate({ email, password, name: email.split('@')[0], referred: '' });
  };

  const handleGoogle = async () => {
    try {
      const result = await signInWithPopup(auth, provider);
      const idToken = await getIdToken(result.user);
      googleAuthMutation.mutate({ idToken });
    } catch (error) { notify.error((error as any).message); }
  };

  useEffect(() => { clearError(); }, [clearError]);
  useEffect(() => {
    if (isAuthenticated) navigate(role === 'business' ? '/auth/questionnaire' : '/');
  }, [isAuthenticated, navigate]);

  return (
    <div className="min-h-screen flex items-center justify-center bg-[#f8f8fb] p-6">
      <div className="w-full max-w-md">
        <div className="bg-white rounded-2xl shadow-card border border-black/5 p-8">
          <div className="mb-8">
            <h2 className="text-2xl font-bold text-primary mb-1.5">Welcome to Rtylr.</h2>
            <p className="text-sm text-muted">Sign up to start managing everything in one place.</p>
          </div>

          <form onSubmit={(e) => { e.preventDefault(); handleRegister(); }} className="flex flex-col gap-5">
            <Input id="register-email" type="email" label="Email" placeholder="Enter your email" value={email} onChange={(e) => setEmail(e.target.value)} required />
            <Input id="register-password" type="password" label="Password" placeholder="Enter your password" value={password} onChange={(e) => setPassword(e.target.value)} required />
            <Input id="register-confirm" type="password" label="Confirm Password" placeholder="Confirm your password" value={confirmPassword} onChange={(e) => setConfirmPassword(e.target.value)} required />

            {error && <p className="text-sm text-destructive text-center">{typeof error === 'string' ? error : (error as any)?.message || 'An error occurred'}</p>}

            <Button variant="primary" size="md" fullWidth loading={loading} type="submit">Sign Up</Button>

            {showGoogle && (
              <>
                <div className="flex items-center gap-3 my-1">
                  <div className="flex-1 h-px bg-black/8" />
                  <span className="text-xs text-muted">or</span>
                  <div className="flex-1 h-px bg-black/8" />
                </div>
                <Button variant="outline" size="md" fullWidth type="button" onClick={handleGoogle} prefix={<FaGoogle />}>Sign up with Google</Button>
              </>
            )}
          </form>

          <p className="mt-6 text-sm text-center text-muted">
            Already have an account?{' '}
            <button className="text-accent hover:underline font-medium" onClick={() => navigate('/')}>Sign in</button>
          </p>
        </div>

        <div className="mt-6 text-center space-y-2">
          <p className="text-xs text-muted/70">
            By using Rtylr, you agree to our{' '}
            <a href={`${APP_URLS.landing}/legal/privacy-policy`} target="_blank" rel="noopener noreferrer" className="text-accent hover:underline">Privacy Policy</a>.
          </p>
        </div>
      </div>
    </div>
  );
};

export default Register;
