import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { motion } from 'framer-motion';
import { Mail, Lock, User, AlertCircle, ArrowRight } from 'lucide-react';
import { GoogleAuthProvider, signInWithPopup, getIdToken } from 'firebase/auth';
import { FaGoogle } from 'react-icons/fa';

import notify from '../lib/notify/notify';
import { useUserStore } from '../stores/userStore';
import { useRegisterRestaurant, useGoogleAuth } from '../hooks/useAuth';
import { auth } from '../firebase';
import { APP_URLS } from '@voxire/config';
import { useGoogleSignupVisible } from '../hooks/useFeatures';

import AuthShell from '../components/AuthShell';
import AuthInput from '../components/AuthInput';
import AuthButton from '../components/AuthButton';
import PasswordStrengthMeter, {
  scorePassword,
} from '../components/PasswordStrengthMeter';
import TermsCheckbox from '../components/TermsCheckbox';
import AuthDivider from '../components/AuthDivider';
import {
  getEmailError,
  getPasswordMatchError,
} from '../lib/validation';

const fieldStagger = {
  hidden: { opacity: 0, y: 12 },
  show: (i: number) => ({
    opacity: 1,
    y: 0,
    transition: { delay: 0.2 + i * 0.06, duration: 0.5, ease: [0.16, 1, 0.3, 1] },
  }),
};

const Register: React.FC = () => {
  const navigate = useNavigate();
  const provider = new GoogleAuthProvider();
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [agree, setAgree] = useState(false);
  const emailError = getEmailError(email);
  const matchError = getPasswordMatchError(password, confirmPassword);
  const pwScore = scorePassword(password);
  const { isAuthenticated, loading, error, role, clearError } = useUserStore();
  const registerMutation = useRegisterRestaurant();
  const googleAuthMutation = useGoogleAuth();
  const showGoogle: boolean = useGoogleSignupVisible();

  const handleRegister = () => {
    if (password !== confirmPassword) {
      notify.error('Passwords do not match');
      return;
    }
    if (!agree) {
      notify.error('Please accept the terms to continue');
      return;
    }
    registerMutation.mutate({
      email,
      password,
      name: name.trim() || email.split('@')[0],
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
      subtitle="Start running your entire business from one place."
      footer={
        <div className="space-y-3 text-center">
          <p className="text-sm text-white/55">
            Already have an account?{' '}
            <button
              type="button"
              className="text-[#64C5B9] font-medium hover:underline underline-offset-4"
              onClick={() => navigate('/')}
            >
              Sign in
            </button>
          </p>
          <p className="text-[11px] text-white/35">
            By creating an account you agree to our{' '}
            <a
              href={`${APP_URLS.landing}/legal/privacy-policy`}
              target="_blank"
              rel="noopener noreferrer"
              className="text-white/55 hover:text-white underline underline-offset-2"
            >
              Privacy Policy
            </a>
            .
          </p>
        </div>
      }
    >
      <form
        onSubmit={(e) => { e.preventDefault(); handleRegister(); }}
        className="flex flex-col gap-5"
      >
        <motion.div custom={0} initial="hidden" animate="show" variants={fieldStagger}>
          <AuthInput
            id="register-name"
            type="text"
            label="Full name"
            icon={User}
            placeholder="Jane Doe"
            autoComplete="name"
            value={name}
            onChange={(e) => setName(e.target.value)}
          />
        </motion.div>

        <motion.div custom={1} initial="hidden" animate="show" variants={fieldStagger}>
          <AuthInput
            id="register-email"
            type="email"
            label="Email"
            icon={Mail}
            placeholder="you@company.com"
            autoComplete="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            error={emailError}
            required
          />
        </motion.div>

        <motion.div
          custom={2}
          initial="hidden"
          animate="show"
          variants={fieldStagger}
          className="flex flex-col gap-2"
        >
          <AuthInput
            id="register-password"
            type="password"
            label="Password"
            icon={Lock}
            placeholder="At least 8 characters"
            autoComplete="new-password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
          />
          {password && <PasswordStrengthMeter password={password} />}
        </motion.div>

        <motion.div custom={3} initial="hidden" animate="show" variants={fieldStagger}>
          <AuthInput
            id="register-confirm"
            type="password"
            label="Confirm password"
            icon={Lock}
            placeholder="Repeat password"
            autoComplete="new-password"
            value={confirmPassword}
            onChange={(e) => setConfirmPassword(e.target.value)}
            error={matchError}
            required
          />
        </motion.div>

        <motion.div custom={4} initial="hidden" animate="show" variants={fieldStagger}>
          <TermsCheckbox checked={agree} onChange={setAgree} />
        </motion.div>

        {errorMessage && (
          <motion.div
            initial={{ opacity: 0, y: -6 }}
            animate={{ opacity: 1, y: 0 }}
            className="flex items-start gap-2 rounded-lg border border-red-400/25 bg-red-500/10 px-3.5 py-2.5"
            role="alert"
          >
            <AlertCircle className="h-4 w-4 text-red-300 mt-0.5 shrink-0" />
            <p className="text-xs text-red-200 leading-relaxed">{errorMessage}</p>
          </motion.div>
        )}

        <motion.div custom={5} initial="hidden" animate="show" variants={fieldStagger}>
          <AuthButton
            type="submit"
            loading={loading}
            disabled={
              !email ||
              !!emailError ||
              pwScore < 2 ||
              !!matchError ||
              !agree
            }
          >
            <span className="inline-flex items-center justify-center gap-2">
              Create account
              <ArrowRight className="h-[18px] w-[18px]" />
            </span>
          </AuthButton>
        </motion.div>

        {showGoogle && (
          <motion.div
            custom={6}
            initial="hidden"
            animate="show"
            variants={fieldStagger}
            className="space-y-4"
          >
            <AuthDivider />

            <AuthButton type="button" variant="outline" onClick={handleGoogle}>
              <span className="inline-flex items-center justify-center gap-2.5">
                <FaGoogle className="text-white/85" />
                Sign up with Google
              </span>
            </AuthButton>
          </motion.div>
        )}
      </form>
    </AuthShell>
  );
};

export default Register;
