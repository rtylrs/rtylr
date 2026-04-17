import React, { useEffect, useMemo, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { AlertCircle, ArrowLeft, Check } from 'lucide-react';
import { GoogleAuthProvider, signInWithPopup, getIdToken } from 'firebase/auth';
import { FaGoogle } from 'react-icons/fa';
import { APP_URLS } from '@voxire/config';

import notify from '../../shared/lib/notify/notify';
import { useUserStore } from '../../shared/stores/userStore';
import { useRegisterRestaurant, useGoogleAuth } from '../../shared/hooks/useAuth';
import { auth } from '../../firebase';
import { Button } from '@voxire/components';
import AuthShell from '../../shared/components/AuthShell';
import PasswordStrengthMeter, {
  scorePassword,
} from '../../shared/components/PasswordStrengthMeter';
import { getEmailError, getPasswordMatchError } from '../../shared/lib/validation';

// ── Plan definitions ────────────────────────────────────────────────────────

interface PlanDef {
  slug: string;
  name: string;
  tagline: string;
  apps: string[];
  isSolo: boolean;
  featured?: boolean;
}

const APP_COLORS: Record<string, string> = {
  Flow: '#64C5B9',
  Finance: '#ef4444',
  CRM: '#8b5cf6',
  HR: '#f59e0b',
  ERP: '#10b981',
  POS: '#4583C4',
};

const PLANS: PlanDef[] = [
  {
    slug: 'flow_solo',
    name: 'Flow',
    tagline: 'Projects & tasks',
    apps: ['Flow'],
    isSolo: true,
  },
  {
    slug: 'finance_solo',
    name: 'Finance',
    tagline: 'Invoices & expenses',
    apps: ['Finance'],
    isSolo: true,
  },
  {
    slug: 'starter',
    name: 'Starter',
    tagline: 'For small businesses',
    apps: ['Flow', 'Finance', 'CRM'],
    isSolo: false,
  },
  {
    slug: 'pro',
    name: 'Pro',
    tagline: 'For growing teams',
    apps: ['Flow', 'Finance', 'CRM', 'HR', 'ERP'],
    isSolo: false,
    featured: true,
  },
  {
    slug: 'enterprise',
    name: 'Enterprise',
    tagline: 'Full suite with POS',
    apps: ['Flow', 'Finance', 'CRM', 'HR', 'ERP', 'POS'],
    isSolo: false,
  },
];

// Which redirect params map to a solo plan (skip plan-selection step)
const APP_PLAN_MAP: Record<string, string> = {
  flow: 'flow_solo',
  finance: 'finance_solo',
};

type KnownAppId = keyof typeof APP_URLS;

function getRedirectUrl(redirectParam: string | null): string {
  if (redirectParam && redirectParam in APP_URLS) {
    return APP_URLS[redirectParam as KnownAppId] as string;
  }
  return APP_URLS.dash as string;
}

// ── Sub-components ──────────────────────────────────────────────────────────

function AppPill({ name }: { name: string }): React.JSX.Element {
  const color = APP_COLORS[name] ?? '#999';
  return (
    <span
      className="inline-flex items-center gap-1 text-[11px] font-medium px-2 py-0.5 rounded-full"
      style={{ background: `${color}18`, color }}
    >
      <span className="w-1.5 h-1.5 rounded-full" style={{ background: color }} />
      {name}
    </span>
  );
}

interface PlanCardProps {
  plan: PlanDef;
  selected: boolean;
  onSelect: () => void;
}

function PlanCard({ plan, selected, onSelect }: PlanCardProps): React.JSX.Element {
  return (
    <button
      type="button"
      onClick={onSelect}
      className={[
        'w-full text-left rounded-xl border p-4 transition-all',
        selected
          ? 'border-[#4583C4] ring-1 ring-[#4583C4] bg-[#4583C4]/5'
          : 'border-[#e5e5e5] bg-white hover:border-[#4583C4]/50',
      ].join(' ')}
    >
      <div className="flex items-start justify-between gap-2">
        <div>
          <div className="flex items-center gap-2">
            <span className="text-sm font-semibold text-[#111]">{plan.name}</span>
            {plan.featured && (
              <span className="text-[10px] font-semibold text-white bg-[#4583C4] px-1.5 py-0.5 rounded-full leading-tight">
                Popular
              </span>
            )}
          </div>
          <p className="text-[11px] text-[#999] mt-0.5">{plan.tagline}</p>
        </div>
        <div
          className={[
            'w-4 h-4 rounded-full border-2 flex items-center justify-center flex-shrink-0 mt-0.5 transition-colors',
            selected ? 'bg-[#4583C4] border-[#4583C4]' : 'border-[#ddd] bg-white',
          ].join(' ')}
        >
          {selected && <Check size={9} color="white" strokeWidth={3} />}
        </div>
      </div>
      <div className="flex flex-wrap gap-1.5 mt-3">
        {plan.apps.map((app) => (
          <AppPill key={app} name={app} />
        ))}
      </div>
    </button>
  );
}

// ── Register ────────────────────────────────────────────────────────────────

type Step = 'plan' | 'details';

const Register: React.FC = (): React.JSX.Element => {
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

  const redirectParam = useMemo((): string | null => {
    const params = new URLSearchParams(window.location.search);
    return params.get('redirect');
  }, []);

  // isSoloApp: coming in with a redirect for an app that has its own solo plan
  const isSoloApp = redirectParam !== null && redirectParam in APP_PLAN_MAP;

  // Pre-select plan based on redirect param
  const defaultPlanSlug = redirectParam ? (APP_PLAN_MAP[redirectParam] ?? 'starter') : 'starter';

  const [selectedPlanSlug, setSelectedPlanSlug] = useState<string>(defaultPlanSlug);
  const [step, setStep] = useState<Step>(isSoloApp ? 'details' : 'plan');

  const selectedPlan = PLANS.find((p) => p.slug === selectedPlanSlug) ?? PLANS[2];
  const isSoloPlan = selectedPlan.isSolo;

  const handleRegister = (): void => {
    if (password !== confirmPassword) {
      notify.error('Passwords do not match');
      return;
    }
    registerMutation.mutate(
      {
        email,
        password,
        name: name.trim() || email.split('@')[0],
        businessName: isSoloPlan ? '' : businessName.trim(),
        referred: '',
        planSlug: selectedPlanSlug,
      },
      {
        onSuccess: () => {
          window.location.href = getRedirectUrl(redirectParam);
        },
      }
    );
  };

  const handleGoogle = async (): Promise<void> => {
    try {
      const result = await signInWithPopup(auth, provider);
      const idToken = await getIdToken(result.user);
      googleAuthMutation.mutate(
        { idToken },
        {
          onSuccess: () => {
            window.location.href = getRedirectUrl(redirectParam);
          },
        }
      );
    } catch (err) {
      notify.error((err as Error).message);
    }
  };

  useEffect((): void => {
    clearError();
  }, [clearError]);

  useEffect((): void => {
    const hasToken = document.cookie.split(';').some((c) => c.trim().startsWith('token='));
    if (isAuthenticated && hasToken) {
      navigate(role === 'business' ? '/auth/questionnaire' : '/');
    }
  }, [isAuthenticated, navigate, role]);

  const errorMessage =
    error &&
    (typeof error === 'string'
      ? error
      : (error as { message?: string })?.message || 'An error occurred');

  // ── Step 1: Plan selection ────────────────────────────────────────────────
  if (step === 'plan') {
    const soloPlans = PLANS.filter((p) => p.isSolo);
    const businessPlans = PLANS.filter((p) => !p.isSolo);

    return (
      <AuthShell
        title="Choose your plan."
        subtitle="Start free — no credit card required."
        footer={
          <p className="text-center text-sm text-[#666]">
            Already have an account?{' '}
            <button
              type="button"
              className="text-[#4583C4] font-medium hover:underline underline-offset-2"
              onClick={() => navigate('/')}
            >
              Sign in
            </button>
          </p>
        }
      >
        <div className="flex flex-col gap-3">
          <p className="text-[11px] font-semibold text-[#bbb] uppercase tracking-wider">
            Individual apps
          </p>
          <div className="grid grid-cols-2 gap-2">
            {soloPlans.map((plan) => (
              <PlanCard
                key={plan.slug}
                plan={plan}
                selected={selectedPlanSlug === plan.slug}
                onSelect={() => setSelectedPlanSlug(plan.slug)}
              />
            ))}
          </div>

          <p className="text-[11px] font-semibold text-[#bbb] uppercase tracking-wider mt-1">
            Full suite
          </p>
          <div className="flex flex-col gap-2">
            {businessPlans.map((plan) => (
              <PlanCard
                key={plan.slug}
                plan={plan}
                selected={selectedPlanSlug === plan.slug}
                onSelect={() => setSelectedPlanSlug(plan.slug)}
              />
            ))}
          </div>

          <Button
            type="button"
            variant="primary"
            fullWidth
            height={40}
            onClick={() => setStep('details')}
          >
            Continue with {selectedPlan.name}
          </Button>
        </div>
      </AuthShell>
    );
  }

  // ── Step 2: Account details ───────────────────────────────────────────────
  const title = isSoloApp ? `Get started with ${selectedPlan.name}.` : 'Create your account.';
  const subtitle = isSoloApp
    ? `Create your free ${selectedPlan.name} account.`
    : `Setting up ${selectedPlan.name} plan.`;

  return (
    <AuthShell
      title={title}
      subtitle={subtitle}
      footer={
        <div className="text-center">
          <p className="text-sm text-[#666]">
            Already have an account?{' '}
            <button
              type="button"
              className="text-[#4583C4] font-medium hover:underline underline-offset-2"
              onClick={() => navigate(redirectParam ? `/?redirect=${redirectParam}` : '/')}
            >
              Sign in
            </button>
          </p>
        </div>
      }
    >
      {/* Back to plan selection (not shown for solo-app redirect flows) */}
      {!isSoloApp && (
        <button
          type="button"
          onClick={() => setStep('plan')}
          className="flex items-center gap-1.5 text-[13px] text-[#666] hover:text-[#111] mb-5 -mt-2 transition-colors"
        >
          <ArrowLeft size={14} />
          Change plan
        </button>
      )}

      <form
        onSubmit={(e) => {
          e.preventDefault();
          handleRegister();
        }}
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

        {/* Business name — hidden for solo plans */}
        {!isSoloPlan && (
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
        )}

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

        {/* Submit */}
        <Button
          type="submit"
          variant="primary"
          fullWidth
          loading={loading}
          height={40}
          disabled={!email || !!emailError || pwScore < 2 || !!matchError}
        >
          Create account
        </Button>

        {/* Google sign up */}
        <Button type="button" variant="outline" fullWidth height={40} onClick={handleGoogle}>
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
