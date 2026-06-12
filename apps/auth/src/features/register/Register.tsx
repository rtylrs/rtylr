import React, { useEffect, useMemo, useState } from 'react';
import { useNavigate } from '@tanstack/react-router';
import { useQuery } from '@tanstack/react-query';
import { AlertCircle, ArrowLeft, Check, Loader2 } from 'lucide-react';
import { FaGoogle, FaGithub, FaMicrosoft, FaApple } from 'react-icons/fa';
import { APP_URLS } from '@voxire/config';

import { unwrap } from '@voxire/api';
import { safe } from '../../shared/lib/api';
import notify from '../../shared/lib/notify/notify';
import { useUserStore } from '../../shared/stores/userStore';
import { useRegisterRestaurant, checkEmailExists } from '../../shared/hooks/useAuth';
import { startSocialLogin, type SocialProvider } from '../../shared/lib/social';
import { Button, HelmetContainer } from '@voxire/components';
import AuthShell from '../../shared/components/AuthShell';
import PasswordStrengthMeter, {
  scorePassword,
} from '../../shared/components/PasswordStrengthMeter';
import { getEmailError, getPasswordMatchError } from '../../shared/lib/validation';

// ── Plan types (mirrors onion /subscriptions/plans response) ────────────────

interface ApiPlan {
  id: number;
  name: string;
  slug: string;
  apps: string[];
  pricePerSeatCents: number;
  annualPricePerSeatCents: number;
  trialDays: number;
  tagline: string;
  maxSeats: number | null;
  maxLocations: number | null;
  aiCreditsPerMonth: number;
  features: string[];
  tierOrder: number;
  isSolo: boolean;
  isFeatured: boolean;
  isActive: boolean;
}

const APP_COLORS: Record<string, string> = {
  flow: '#64C5B9',
  finance: '#ef4444',
  crm: '#8b5cf6',
  hr: '#f59e0b',
  erp: '#10b981',
  pos: '#4583C4',
  dash: '#111111',
  insights: '#eab308',
};

const APP_LABEL: Record<string, string> = {
  flow: 'Flow',
  finance: 'Finance',
  crm: 'CRM',
  hr: 'HR',
  erp: 'ERP',
  pos: 'POS',
  dash: 'Dash',
  insights: 'Insights',
};

// Redirect param → solo plan slug. Every app has a solo tier.
const APP_PLAN_MAP: Record<string, string> = {
  flow: 'flow_solo',
  finance: 'finance_solo',
  crm: 'crm_solo',
  hr: 'hr_solo',
  erp: 'erp_solo',
  pos: 'pos_solo',
};

type KnownAppId = keyof typeof APP_URLS;

function getRedirectUrl(redirectParam: string | null): string {
  if (redirectParam && redirectParam in APP_URLS) {
    return APP_URLS[redirectParam as KnownAppId] as string;
  }
  return APP_URLS.dash as string;
}

function formatPrice(cents: number): string {
  if (cents === 0) return 'Free';
  const dollars = cents / 100;
  return `$${dollars.toFixed(2)}`;
}

// ── Sub-components ──────────────────────────────────────────────────────────

function AppPill({ appId }: { appId: string }): React.JSX.Element {
  const color = APP_COLORS[appId] ?? '#999';
  const label = APP_LABEL[appId] ?? appId;
  return (
    <span
      className="inline-flex items-center gap-1 text-[11px] font-medium px-2 py-0.5 rounded-full"
      style={{ background: `${color}18`, color }}
    >
      <span className="w-1.5 h-1.5 rounded-full" style={{ background: color }} />
      {label}
    </span>
  );
}

interface PlanCardProps {
  plan: ApiPlan;
  selected: boolean;
  onSelect: () => void;
  compact?: boolean;
}

function PlanCard({ plan, selected, onSelect, compact }: PlanCardProps): React.JSX.Element {
  const priceLabel = formatPrice(plan.pricePerSeatCents);
  const showPrice = !plan.isSolo && plan.pricePerSeatCents > 0;
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
        <div className="min-w-0">
          <div className="flex items-center gap-2 flex-wrap">
            <span className="text-sm font-semibold text-[#111]">{plan.name}</span>
            {plan.isFeatured && (
              <span className="text-[10px] font-semibold text-white bg-[#4583C4] px-1.5 py-0.5 rounded-full leading-tight">
                Popular
              </span>
            )}
            {showPrice && (
              <span className="text-[11px] text-[#666] font-medium">
                {priceLabel}
                <span className="text-[#999]">/seat/mo</span>
              </span>
            )}
          </div>
          <p className="text-[11px] text-[#999] mt-0.5 line-clamp-2">{plan.tagline}</p>
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
      {!compact && (
        <div className="flex flex-wrap gap-1.5 mt-3">
          {plan.apps.slice(0, 8).map((appId) => (
            <AppPill key={appId} appId={appId} />
          ))}
        </div>
      )}
    </button>
  );
}

// ── Register ────────────────────────────────────────────────────────────────

type Step = 'plan' | 'details';

const Register: React.FC = (): React.JSX.Element => {
  const navigate = useNavigate();
  const [name, setName] = useState('');
  const [businessName, setBusinessName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirm, setShowConfirm] = useState(false);
  const [emailTaken, setEmailTaken] = useState(false);
  const emailError = getEmailError(email);
  const matchError = getPasswordMatchError(password, confirmPassword);

  // Check email availability once the field is filled and well-formed, so the
  // user sees "email already exists" before submitting.
  const checkEmailAvailable = async (): Promise<void> => {
    if (getEmailError(email)) return;
    try {
      const { exists } = await checkEmailExists(email);
      setEmailTaken(exists);
    } catch {
      setEmailTaken(false); // never block signup on a lookup failure
    }
  };
  const pwScore = scorePassword(password);
  const { isAuthenticated, loading, error, role, clearError } = useUserStore();
  const registerMutation = useRegisterRestaurant();

  const redirectParam = useMemo((): string | null => {
    const params = new URLSearchParams(window.location.search);
    return params.get('redirect');
  }, []);

  const isSoloApp = redirectParam !== null && redirectParam in APP_PLAN_MAP;

  const defaultPlanSlug = redirectParam ? (APP_PLAN_MAP[redirectParam] ?? 'free') : 'free';

  const [selectedPlanSlug, setSelectedPlanSlug] = useState<string>(defaultPlanSlug);
  const [step, setStep] = useState<Step>(isSoloApp ? 'details' : 'plan');

  // Fetch plans from the backend
  const plansQuery = useQuery({
    queryKey: ['subscription-plans'],
    queryFn: async (): Promise<ApiPlan[]> => {
      try {
        const res = await unwrap(safe.get<{ plans: ApiPlan[] }>('/subscriptions/plans'));
        return res.plans ?? [];
      } catch {
        return [];
      }
    },
    staleTime: 5 * 60 * 1000,
  });

  const plans = plansQuery.data ?? [];
  const selectedPlan =
    plans.find((p) => p.slug === selectedPlanSlug) ??
    plans.find((p) => p.slug === 'free') ??
    plans[0];
  const isSoloPlan = selectedPlan?.isSolo ?? false;

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
        planSlug: selectedPlanSlug,
      },
      {
        onSuccess: () => {
          window.location.href = getRedirectUrl(redirectParam);
        },
      }
    );
  };

  const handleSocial = async (provider: SocialProvider): Promise<void> => {
    try {
      await startSocialLogin(provider, redirectParam);
    } catch {
      notify.error('Social sign-in is currently unavailable');
    }
  };

  useEffect((): void => {
    clearError();
  }, [clearError]);

  useEffect((): void => {
    const hasToken = document.cookie.split(';').some((c) => c.trim().startsWith('token='));
    if (isAuthenticated && hasToken) {
      window.location.href = getRedirectUrl(redirectParam);
    }
  }, [isAuthenticated, navigate, role, redirectParam]);

  const errorMessage =
    error &&
    (typeof error === 'string'
      ? error
      : (error as { message?: string })?.message || 'An error occurred');

  // ── Step 1: Plan selection ────────────────────────────────────────────────
  if (step === 'plan') {
    if (plansQuery.isLoading) {
      return (
        <AuthShell title="Choose your plan." subtitle="Loading plans…">
          <div className="flex items-center justify-center py-12">
            <Loader2 className="animate-spin text-[#4583C4]" size={24} />
          </div>
        </AuthShell>
      );
    }

    if (plansQuery.isError || plans.length === 0) {
      return (
        <AuthShell title="Choose your plan." subtitle="Couldn't load plans.">
          <div className="flex flex-col gap-3">
            <div className="flex items-start gap-2 rounded-lg border border-red-200 bg-red-50 px-3 py-2.5">
              <AlertCircle className="h-4 w-4 text-red-500 mt-0.5 shrink-0" />
              <p className="text-xs text-red-600 leading-relaxed">
                We couldn't reach our servers. Check your connection and retry.
              </p>
            </div>
            <Button
              type="button"
              variant="primary"
              className="auth-submit-button"
              fullWidth
              height={40}
              onClick={() => plansQuery.refetch()}
            >
              Try again
            </Button>
          </div>
        </AuthShell>
      );
    }

    const tierPlans = plans.filter((p) => !p.isSolo).sort((a, b) => a.tierOrder - b.tierOrder);
    const soloPlans = plans.filter((p) => p.isSolo).sort((a, b) => a.tierOrder - b.tierOrder);

    return (
      <AuthShell
        title="Pick a plan."
        subtitle="Start free. Upgrade when you outgrow it."
        footer={
          <p className="text-center text-sm text-[#666]">
            Already have an account?{' '}
            <button
              type="button"
              className="text-[#4583C4] font-medium hover:underline underline-offset-2"
              onClick={() => navigate({ to: '/' })}
            >
              Sign in
            </button>
          </p>
        }
      >
        <div className="flex flex-col gap-3">
          <p className="text-[11px] font-semibold text-[#bbb] uppercase tracking-wider">
            Full suite - every app, one subscription
          </p>
          <div className="flex flex-col gap-2">
            {tierPlans.map((plan) => (
              <PlanCard
                key={plan.slug}
                plan={plan}
                selected={selectedPlanSlug === plan.slug}
                onSelect={() => setSelectedPlanSlug(plan.slug)}
              />
            ))}
          </div>

          {soloPlans.length > 0 && (
            <>
              <p className="text-[11px] font-semibold text-[#bbb] uppercase tracking-wider mt-1">
                Or start with one app
              </p>
              <div className="grid grid-cols-2 gap-2">
                {soloPlans.map((plan) => (
                  <PlanCard
                    key={plan.slug}
                    plan={plan}
                    selected={selectedPlanSlug === plan.slug}
                    onSelect={() => setSelectedPlanSlug(plan.slug)}
                    compact
                  />
                ))}
              </div>
            </>
          )}

          <Button
            type="button"
            variant="primary"
            className="auth-submit-button"
            fullWidth
            height={40}
            onClick={() => setStep('details')}
            disabled={!selectedPlan}
          >
            Continue with {selectedPlan?.name ?? 'plan'}
          </Button>
        </div>
      </AuthShell>
    );
  }

  // ── Step 2: Account details ───────────────────────────────────────────────
  const planName = selectedPlan?.name ?? 'your plan';
  const title = isSoloApp ? `Get started with ${planName}.` : 'Create your account.';
  const subtitle = isSoloApp
    ? `Free forever. No credit card.`
    : `You picked ${planName}. Let's set you up.`;

  return (
    <>
      <HelmetContainer title="Create account · rtylr" />
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
                onClick={() =>
                  navigate({ to: redirectParam ? `/?redirect=${redirectParam}` : '/' })
                }
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

          {/* Business name, hidden for solo plans */}
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
              onChange={(e) => {
                setEmail(e.target.value);
                if (emailTaken) setEmailTaken(false);
              }}
              onBlur={() => void checkEmailAvailable()}
              required
              className="w-full h-10 rounded-lg border border-[#e5e5e5] bg-white px-3 text-sm text-[#111] placeholder:text-[#999] outline-none transition-colors focus:border-[#4583C4] focus:ring-2 focus:ring-[#4583C4]/20"
            />
            {emailError && <p className="text-xs text-red-500">{emailError}</p>}
            {!emailError && emailTaken && (
              <p className="text-xs text-red-500">An account with this email already exists.</p>
            )}
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
            className="auth-submit-button"
            fullWidth
            loading={loading}
            height={40}
            disabled={
              !email ||
              !!emailError ||
              emailTaken ||
              pwScore < 2 ||
              !!matchError ||
              (!isSoloPlan && !businessName.trim())
            }
          >
            Create account
          </Button>

          {/* Social sign up */}
          <div className="grid grid-cols-2 gap-2">
            <Button
              type="button"
              variant="outline"
              className="auth-provider-button"
              fullWidth
              height={40}
              onClick={() => handleSocial('google')}
            >
              <span className="inline-flex items-center gap-2">
                <FaGoogle className="text-[#444]" />
                Google
              </span>
            </Button>
            <Button
              type="button"
              variant="outline"
              className="auth-provider-button"
              fullWidth
              height={40}
              onClick={() => handleSocial('github')}
            >
              <span className="inline-flex items-center gap-2">
                <FaGithub className="text-[#444]" />
                GitHub
              </span>
            </Button>
            <Button
              type="button"
              variant="outline"
              className="auth-provider-button"
              fullWidth
              height={40}
              onClick={() => handleSocial('microsoft')}
            >
              <span className="inline-flex items-center gap-2">
                <FaMicrosoft className="text-[#444]" />
                Microsoft
              </span>
            </Button>
            <Button
              type="button"
              variant="outline"
              className="auth-provider-button"
              fullWidth
              height={40}
              onClick={() => handleSocial('apple')}
            >
              <span className="inline-flex items-center gap-2">
                <FaApple className="text-[#444]" />
                Apple
              </span>
            </Button>
          </div>
        </form>
      </AuthShell>
    </>
  );
};

export default Register;
