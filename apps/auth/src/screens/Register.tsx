import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { AlertCircle } from 'lucide-react';
import { GoogleAuthProvider, signInWithPopup, getIdToken } from 'firebase/auth';
import { FaGoogle } from 'react-icons/fa';

import notify from '../lib/notify/notify';
import { useUserStore } from '../stores/userStore';
import { useRegisterRestaurant, useGoogleAuth } from '../hooks/useAuth';
import { auth } from '../firebase';
import { Button } from '@voxire/components';
import AuthShell from '../components/AuthShell';

import PasswordStrengthMeter, {
  scorePassword,
} from '../components/PasswordStrengthMeter';
import {
  getEmailError,
  getPasswordMatchError,
} from '../lib/validation';

const Register: React.FC = () => {
  const navigate = useNavigate();
  const provider = new GoogleAuthProvider();
  const [name, setName] = useState('');
  const [businessName, setBusinessName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirm, setShowConfirm] = useState(false);
  const emailError = getEmailError(email);
  const matchError = getPasswordMatchError(password, confirmPassword);
  const pwScore = scorePassword(password);
  const { isAuthenticated, loading, error, role, clearError } = useUserStore();
  const registerMutation = useRegisterRestaurant();
  const googleAuthMutation = useGoogleAuth();

  const handleRegister = () => {
    if (password !== confirmPassword) {
      notify.error('Passwords do not match');
      return;
    }
    registerMutation.mutate({
      email,
      password,
      name: name.trim() || email.split('@')[0],
      businessName: businessName.trim(),
      referred: '',
    });
  };

  const handleGoogle = async () => {
    try {
      const result = await signInWithPopup(auth, provider);
      const idToken = await getIdToken(result.user);
      googleAuthMutation.mutate({ idToken });
    } catch (err) {
      notify.error((err as any).message);
    }
  };

  useEffect(() => { clearError(); }, [clearError]);
  useEffect(() => {
    if (isAuthenticated) navigate(role === 'business' ? '/auth/questionnaire' : '/');
  }, [isAuthenticated, navigate, role]);

  const errorMessage =
    error && (typeof error === 'string' ? error : (error as any)?.message || 'An error occurred');

  return (
    <AuthShell
      title="Create your account."
      subtitle="Get started with rtylr for free."
      footer={
        <div className="text-center">
          <p className="text-sm text-[#666]">
            Already have an account?{' '}
            <button
              type="button"
              className="text-[#4583C4] font-medium hover:underline underline-offset-2"
              onClick={() => navigate('/')}
            >
              Sign in
            </button>
          </p>
        </div>
      }
    >
        <form
          onSubmit={(e) => { e.preventDefault(); handleRegister(); }}
          className="flex flex-col gap-4"
        >
          {/* Full name */}
          <div className="flex flex-col gap-1.5">
            <label htmlFor="register-name" className="text-sm font-medium text-[#111]">
              Full name
            </label>
            <input
              id="register-name"
              type="text"
              placeholder="Jane Doe"
              autoComplete="name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="w-full h-10 rounded-lg border border-[#e5e5e5] bg-white px-3 text-sm text-[#111] placeholder:text-[#999] outline-none transition-colors focus:border-[#4583C4] focus:ring-2 focus:ring-[#4583C4]/20"
            />
          </div>

          {/* Business name */}
          <div className="flex flex-col gap-1.5">
            <label htmlFor="register-business" className="text-sm font-medium text-[#111]">
              Business name
            </label>
            <input
              id="register-business"
              type="text"
              placeholder="My Coffee Shop"
              value={businessName}
              onChange={(e) => setBusinessName(e.target.value)}
              className="w-full h-10 rounded-lg border border-[#e5e5e5] bg-white px-3 text-sm text-[#111] placeholder:text-[#999] outline-none transition-colors focus:border-[#4583C4] focus:ring-2 focus:ring-[#4583C4]/20"
            />
          </div>

          {/* Email */}
          <div className="flex flex-col gap-1.5">
            <label htmlFor="register-email" className="text-sm font-medium text-[#111]">
              Email
            </label>
            <input
              id="register-email"
              type="email"
              placeholder="you@company.com"
              autoComplete="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
              className="w-full h-10 rounded-lg border border-[#e5e5e5] bg-white px-3 text-sm text-[#111] placeholder:text-[#999] outline-none transition-colors focus:border-[#4583C4] focus:ring-2 focus:ring-[#4583C4]/20"
            />
            {emailError && <p className="text-xs text-red-500">{emailError}</p>}
          </div>

          {/* Password */}
          <div className="flex flex-col gap-1.5">
            <label htmlFor="register-password" className="text-sm font-medium text-[#111]">
              Password
            </label>
            <div className="relative">
              <input
                id="register-password"
                type={showPassword ? 'text' : 'password'}
                placeholder="At least 8 characters"
                autoComplete="new-password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
                className="w-full h-10 rounded-lg border border-[#e5e5e5] bg-white px-3 pr-10 text-sm text-[#111] placeholder:text-[#999] outline-none transition-colors focus:border-[#4583C4] focus:ring-2 focus:ring-[#4583C4]/20"
              />
              <button
                type="button"
                onClick={() => setShowPassword((v) => !v)}
                aria-label={showPassword ? 'Hide password' : 'Show password'}
                className="absolute right-2.5 top-1/2 -translate-y-1/2 text-[#999] hover:text-[#666] text-xs font-medium"
              >
                {showPassword ? 'Hide' : 'Show'}
              </button>
            </div>
            {password && <PasswordStrengthMeter password={password} />}
          </div>

          {/* Confirm password */}
          <div className="flex flex-col gap-1.5">
            <label htmlFor="register-confirm" className="text-sm font-medium text-[#111]">
              Confirm password
            </label>
            <div className="relative">
              <input
                id="register-confirm"
                type={showConfirm ? 'text' : 'password'}
                placeholder="Repeat password"
                autoComplete="new-password"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                required
                className="w-full h-10 rounded-lg border border-[#e5e5e5] bg-white px-3 pr-10 text-sm text-[#111] placeholder:text-[#999] outline-none transition-colors focus:border-[#4583C4] focus:ring-2 focus:ring-[#4583C4]/20"
              />
              <button
                type="button"
                onClick={() => setShowConfirm((v) => !v)}
                aria-label={showConfirm ? 'Hide password' : 'Show password'}
                className="absolute right-2.5 top-1/2 -translate-y-1/2 text-[#999] hover:text-[#666] text-xs font-medium"
              >
                {showConfirm ? 'Hide' : 'Show'}
              </button>
            </div>
            {matchError && <p className="text-xs text-red-500">{matchError}</p>}
          </div>

          {/* Error */}
          {errorMessage && (
            <div
              className="flex items-start gap-2 rounded-lg border border-red-200 bg-red-50 px-3 py-2.5"
              role="alert"
            >
              <AlertCircle className="h-4 w-4 text-red-500 mt-0.5 shrink-0" />
              <p className="text-xs text-red-600 leading-relaxed">{errorMessage}</p>
            </div>
          )}

          {/* Sign up button */}
          <Button
            type="submit"
            variant="primary"
            fullWidth
            loading={loading}
            height={40}
            disabled={
              !email ||
              !!emailError ||
              pwScore < 2 ||
              !!matchError
            }
          >
            Create account
          </Button>

          {/* Google sign up */}
          <Button
            type="button"
            variant="outline"
            fullWidth
            height={40}
            onClick={handleGoogle}
          >
            <span className="inline-flex items-center gap-2">
              <FaGoogle className="text-[#444]" />
              Sign up with Google
            </span>
          </Button>
        </form>
    </AuthShell>
  );
};

export default Register;
