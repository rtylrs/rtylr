import React, { useMemo, useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { useTranslation } from 'react-i18next';
import {
  AlertCircle,
  ArrowRight,
  CalendarDays,
  Check,
  CheckCircle2,
  Clock,
  CreditCard,
  Download,
  Loader2,
  ReceiptText,
  Sparkles,
  Users,
  WalletCards,
  XCircle,
} from 'lucide-react';
import { toast } from 'sonner';
import { APP_CONFIG, APP_URLS } from '@voxire/config';
import type { AppId } from '@voxire/config';
import { EmptyState } from '@voxire/ui';
import { useUserStore } from '../../shared/stores/userStore';
import {
  useBillingDetail,
  useBillingHistory,
  useCheckout,
  useListPlans,
  useMySubscription,
  centsToDisplay,
  trialDaysLeft,
  type BillingDetailPayload,
  type BillingHistoryMonth,
  type BillingLineItem,
  type PlanPayload,
  type SubscriptionPayload,
} from '../../shared/hooks/queries/useSubscription';

type BillingCycle = 'monthly' | 'yearly';
type SubStatus = SubscriptionPayload['status'];

// Checkout URLs come from the subscription backend but trigger a full-page
// navigation, so validate them against the Paddle hosted-checkout hosts
// before assigning to window.location.
const CHECKOUT_HOSTS = ['buy.paddle.com', 'checkout.paddle.com', 'paddle.com'];

function isAllowedCheckoutUrl(raw: unknown): raw is string {
  if (typeof raw !== 'string' || raw.length === 0 || raw.length > 2048) return false;
  let url: URL;
  try {
    url = new URL(raw);
  } catch {
    return false;
  }
  if (url.protocol !== 'https:') return false;
  const host = url.hostname.toLowerCase();
  return CHECKOUT_HOSTS.some((allowed) => host === allowed || host.endsWith(`.${allowed}`));
}

function navigateToCheckoutUrl(checkoutUrl: unknown): void {
  if (!isAllowedCheckoutUrl(checkoutUrl)) {
    toast.error('Failed to start checkout. Please try again.');
    return;
  }
  window.location.href = checkoutUrl;
}

interface StatusConfig {
  label: string;
  color: string;
  bg: string;
  Icon: React.ComponentType<{ size?: number }>;
}

const STATUS_MAP: Record<SubStatus, StatusConfig> = {
  active: {
    label: 'Active',
    color: 'var(--status-active)',
    bg: 'var(--status-active-bg)',
    Icon: CheckCircle2,
  },
  trial: {
    label: 'Trial',
    color: 'var(--status-trial)',
    bg: 'var(--status-trial-bg)',
    Icon: Clock,
  },
  cancelled: {
    label: 'Cancelled',
    color: 'var(--status-cancelled)',
    bg: 'var(--status-cancelled-bg)',
    Icon: XCircle,
  },
  expired: {
    label: 'Expired',
    color: 'var(--status-expired)',
    bg: 'var(--status-expired-bg)',
    Icon: AlertCircle,
  },
};

function formatMonth(iso: string): string {
  if (!iso) return 'Billing Period';
  const normalizedIso = iso.includes('-') && iso.length === 7 ? `${iso}-01` : iso;
  const date = new Date(normalizedIso);
  if (Number.isNaN(date.getTime())) return iso;
  return date.toLocaleDateString('en-US', { month: 'long', year: 'numeric', timeZone: 'UTC' });
}

function formatDate(iso: string): string {
  if (!iso) return '';
  const date = new Date(iso);
  if (Number.isNaN(date.getTime())) return iso;
  return date.toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
    timeZone: 'UTC',
  });
}

function planPrice(plan: PlanPayload, cycle: BillingCycle): number {
  if (cycle === 'yearly' && plan.annualPricePerSeatCents > 0) {
    return plan.annualPricePerSeatCents;
  }
  return plan.pricePerSeatCents;
}

function discountLabel(plan: PlanPayload): string | null {
  if (plan.pricePerSeatCents <= 0 || plan.annualPricePerSeatCents <= 0) return null;
  const annualMonthlyEquivalent = plan.annualPricePerSeatCents / 12;
  const discount = Math.round((1 - annualMonthlyEquivalent / plan.pricePerSeatCents) * 100);
  return discount > 0 ? `Save ${discount}%` : null;
}

function grantDetail(planName: string, grantExpiresAt?: string | null): string {
  if (!grantExpiresAt) return `${planName} access granted`;
  return `${planName} access granted until ${formatDate(grantExpiresAt)}`;
}

function splitIncludedSeats(
  seats: number,
  includedLimit: number
): { included: number; billable: number } {
  const safeSeats = Math.max(0, seats);
  const safeLimit = Math.max(0, includedLimit);
  const included = Math.min(safeSeats, safeLimit);
  return { included, billable: Math.max(0, safeSeats - included) };
}

const SectionHeader: React.FC<{
  title: string;
  description?: string;
  action?: React.ReactNode;
}> = ({ title, description, action }) => (
  <div className="flex items-end justify-between gap-4">
    <div>
      <h2 className="text-sm font-semibold" style={{ color: 'var(--dash-text)' }}>
        {title}
      </h2>
      {description ? (
        <p className="text-xs mt-0.5" style={{ color: 'var(--dash-text-dim)' }}>
          {description}
        </p>
      ) : null}
    </div>
    {action}
  </div>
);

const StatusBadge: React.FC<{ status: SubStatus }> = ({ status }) => {
  const cfg = STATUS_MAP[status] ?? STATUS_MAP.expired;
  const { Icon } = cfg;
  return (
    <span
      className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-semibold"
      style={{ color: cfg.color, background: cfg.bg }}
    >
      <Icon size={10} />
      {cfg.label}
    </span>
  );
};

const AppPill: React.FC<{ appId: string }> = ({ appId }) => {
  const id = appId as AppId;
  const config = APP_CONFIG[id];
  const name = config?.name ?? appId;
  // APP_URLS is typed Record<AppId, string>, but `id` is an `as AppId` cast on
  // an arbitrary string, so an unknown app id lands here with no entry.
  // Annotated honestly so the guard below is doing something.
  const url: string | undefined = APP_URLS[id];

  const pill = (
    <span
      className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium"
      style={{
        background: 'var(--accent-muted)',
        color: 'var(--accent)',
        border: '1px solid var(--accent-border)',
      }}
    >
      {name}
    </span>
  );

  if (!url) return pill;

  return (
    <a href={url} target="_blank" rel="noopener noreferrer">
      {pill}
    </a>
  );
};

const CycleToggle: React.FC<{
  cycle: BillingCycle;
  onChange: (cycle: BillingCycle) => void;
  disabled?: boolean;
}> = ({ cycle, onChange, disabled = false }) => (
  <div
    className="inline-flex rounded-lg p-0.5"
    style={{ background: 'var(--dash-card)', border: '1px solid var(--dash-border)' }}
  >
    {(['monthly', 'yearly'] as const).map((value) => {
      const active = cycle === value;
      return (
        <button
          key={value}
          type="button"
          onClick={() => {
            if (!disabled) onChange(value);
          }}
          disabled={disabled}
          className="px-3 py-1.5 rounded-md text-xs font-semibold transition-colors capitalize disabled:cursor-not-allowed disabled:opacity-60"
          style={{
            background: active ? '#4583C4' : 'transparent',
            color: active ? '#fff' : 'var(--dash-text-muted)',
          }}
        >
          {value}
        </button>
      );
    })}
  </div>
);

const SummaryTile: React.FC<{
  label: string;
  value: string;
  detail?: string;
  icon: React.ComponentType<{ size?: number; style?: React.CSSProperties }>;
  children?: React.ReactNode;
}> = ({ label, value, detail, icon: Icon, children }) => (
  <div
    className="rounded-xl p-4 min-h-[116px] flex flex-col justify-between"
    style={{ background: 'var(--dash-card)', border: '1px solid var(--dash-border)' }}
  >
    <div className="flex items-center justify-between gap-3">
      <p
        className="text-[11px] font-semibold uppercase tracking-wider"
        style={{ color: 'var(--dash-text-dim)' }}
      >
        {label}
      </p>
      <Icon size={15} style={{ color: 'var(--dash-text-dim)' }} />
    </div>
    <div>
      <div className="flex items-center gap-2">
        <p className="text-2xl font-bold" style={{ color: 'var(--dash-text)' }}>
          {value}
        </p>
        {children}
      </div>
      {detail ? (
        <p className="text-xs mt-1" style={{ color: 'var(--dash-text-muted)' }}>
          {detail}
        </p>
      ) : null}
    </div>
  </div>
);

const BillingSummary: React.FC<{
  sub: SubscriptionPayload | null | undefined;
  billing: BillingDetailPayload | null | undefined;
}> = ({ sub, billing }) => {
  const isGranted = Boolean(sub?.isGranted ?? billing?.isGranted);
  const seats = billing?.currentSeatCount ?? sub?.seatCount ?? 0;
  const includedSeatLimit = billing?.includedSeatLimit ?? sub?.includedSeatLimit ?? 10;
  const includedSeatCount =
    billing?.includedSeatCount ??
    sub?.includedSeatCount ??
    splitIncludedSeats(seats, includedSeatLimit).included;
  const billableSeatCount =
    billing?.billableSeatCount ??
    sub?.billableSeatCount ??
    splitIncludedSeats(seats, includedSeatLimit).billable;
  const planName = sub?.planName ?? billing?.planName ?? 'No plan';
  const trialDays = trialDaysLeft(sub?.trialEndsAt ?? null);
  const nextDate = billing?.nextBillingDate ?? sub?.nextBillingDate;
  const grantExpiresAt = sub?.grantExpiresAt ?? billing?.grantExpiresAt;

  return (
    <section className="space-y-3">
      <SectionHeader title="Overview" description="Current plan, spend, seats, and renewal." />
      <div className="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-4 gap-3">
        <SummaryTile
          label="Current plan"
          value={isGranted ? 'Granted' : planName}
          detail={
            isGranted
              ? grantDetail(planName, grantExpiresAt)
              : (sub?.planSlug ?? 'Select a plan to activate billing')
          }
          icon={Sparkles}
        >
          {sub ? <StatusBadge status={sub.status} /> : null}
        </SummaryTile>
        <SummaryTile
          label="Spent this month"
          value={billing ? centsToDisplay(billing.totalCents) : '$0.00'}
          detail={billing ? formatMonth(billing.billingMonth) : 'No usage yet'}
          icon={ReceiptText}
        />
        <SummaryTile
          label="Seats"
          value={seats.toLocaleString()}
          detail={
            isGranted
              ? 'Plan access granted'
              : `${includedSeatCount} of ${includedSeatLimit} included · ${billableSeatCount} billable`
          }
          icon={Users}
        />
        <SummaryTile
          label={sub?.status === 'trial' ? 'Trial ends' : 'Next bill'}
          value={
            sub?.status === 'trial' ? `${trialDays}d` : nextDate ? formatDate(nextDate) : 'Not set'
          }
          detail={
            sub?.status === 'trial' && sub.trialEndsAt ? formatDate(sub.trialEndsAt) : undefined
          }
          icon={CalendarDays}
        />
      </div>
    </section>
  );
};

const PlanAction: React.FC<{
  plan: PlanPayload;
  cycle?: BillingCycle;
  seatCount?: number;
  isCurrent: boolean;
  isCancelled?: boolean;
  isGranted?: boolean;
  currentTier?: number;
  ownerEmail: string;
  disabled?: boolean;
  disabledLabel?: string;
}> = ({
  plan,
  cycle = 'monthly',
  seatCount,
  isCurrent,
  isCancelled = false,
  isGranted,
  currentTier,
  ownerEmail,
  disabled = false,
  disabledLabel = 'Unavailable',
}) => {
  const checkout = useCheckout();

  const handleCheckout = (): void => {
    if ((isCurrent && !isCancelled) || disabled || checkout.isPending) return;
    checkout.mutate(
      {
        planSlug: plan.slug,
        billingCycle: cycle,
        seatCount,
        email: ownerEmail || undefined,
      },
      {
        onSuccess: (data) => {
          navigateToCheckoutUrl(data.checkoutUrl);
        },
        onError: () => {
          toast.error('Failed to start checkout. Please try again.');
        },
      }
    );
  };

  const actionLabel = useMemo(() => {
    if (!isCurrent && disabled) return disabledLabel;
    if (isCurrent && isGranted) return 'Granted';
    if (isCurrent && isCancelled) return 'Resubscribe';
    if (isCurrent) return 'Current plan';
    if (currentTier === undefined) return 'Choose plan';
    return plan.tierOrder > currentTier ? 'Upgrade' : 'Switch';
  }, [currentTier, disabled, disabledLabel, isCurrent, isCancelled, isGranted, plan.tierOrder]);

  const isDisabled = (isCurrent && !isCancelled) || disabled || checkout.isPending;

  return (
    <button
      type="button"
      onClick={handleCheckout}
      disabled={isDisabled}
      className="inline-flex shrink-0 items-center justify-center gap-1.5 whitespace-nowrap px-3 py-1.5 rounded-lg text-xs font-semibold transition-opacity disabled:cursor-not-allowed disabled:opacity-70"
      style={{
        background: (isCurrent && !isCancelled) || disabled ? 'var(--dash-hover)' : '#4583C4',
        color: (isCurrent && !isCancelled) || disabled ? 'var(--dash-text-muted)' : '#fff',
        border:
          (isCurrent && !isCancelled) || disabled
            ? '1px solid var(--dash-border)'
            : '1px solid transparent',
      }}
    >
      {checkout.isPending ? <Loader2 size={12} className="animate-spin" /> : null}
      {checkout.isPending ? 'Opening...' : actionLabel}
      {(!isCurrent || isCancelled) && !disabled && !checkout.isPending ? (
        <ArrowRight size={12} className="shrink-0" />
      ) : null}
    </button>
  );
};

const SuitePlanCard: React.FC<{
  plan: PlanPayload;
  cycle: BillingCycle;
  isCurrent: boolean;
  isCancelled?: boolean;
  currentTier?: number;
  seats: number;
  includedSeatLimit: number;
  ownerEmail: string;
  isGranted?: boolean;
  planActionsDisabled?: boolean;
}> = ({
  plan,
  cycle,
  isCurrent,
  isCancelled = false,
  currentTier,
  seats,
  includedSeatLimit,
  ownerEmail,
  isGranted,
  planActionsDisabled = false,
}) => {
  const price = planPrice(plan, cycle);
  const seatSplit = splitIncludedSeats(seats, includedSeatLimit);
  const cycleTotal = price * seatSplit.billable;
  const discount = cycle === 'yearly' ? discountLabel(plan) : null;
  const periodLabel = cycle === 'yearly' ? 'yr' : 'mo';

  return (
    <div
      className="rounded-xl p-4 flex flex-col gap-4"
      style={{
        background: isCurrent ? 'var(--accent-muted)' : 'var(--dash-card)',
        border: isCurrent ? '1.5px solid var(--accent-border)' : '1px solid var(--dash-border)',
      }}
    >
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <div className="flex items-center gap-2">
            <h3 className="text-base font-bold" style={{ color: 'var(--dash-text)' }}>
              {plan.name}
            </h3>
            {plan.isFeatured && !isCurrent ? (
              <span
                className="inline-flex items-center gap-1 text-[10px] font-bold px-2 py-0.5 rounded-full"
                style={{ background: 'var(--accent)', color: '#fff' }}
              >
                <Sparkles size={9} />
                Popular
              </span>
            ) : null}
          </div>
          {plan.tagline ? (
            <p className="text-xs mt-1" style={{ color: 'var(--dash-text-muted)' }}>
              {plan.tagline}
            </p>
          ) : null}
        </div>
        <PlanAction
          plan={plan}
          cycle={cycle}
          seatCount={seats}
          isCurrent={isCurrent}
          isCancelled={isCancelled}
          isGranted={isGranted}
          currentTier={currentTier}
          ownerEmail={ownerEmail}
          disabled={planActionsDisabled}
          disabledLabel="Trial mode"
        />
      </div>

      <div className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <p className="text-2xl font-bold" style={{ color: 'var(--dash-text)' }}>
            {price === 0 ? 'Free' : centsToDisplay(price)}
            {price > 0 ? (
              <span className="text-xs font-medium" style={{ color: 'var(--dash-text-dim)' }}>
                /seat/{periodLabel}
              </span>
            ) : null}
          </p>
          <p className="text-xs mt-1" style={{ color: 'var(--dash-text-dim)' }}>
            {price === 0
              ? 'No subscription cost'
              : `${centsToDisplay(cycleTotal)}/${periodLabel} after ${seatSplit.included} included seat${seatSplit.included !== 1 ? 's' : ''}`}
          </p>
        </div>
        {discount ? (
          <span
            className="text-[11px] font-semibold px-2 py-1 rounded-full"
            style={{ background: 'var(--status-active-bg)', color: 'var(--status-active)' }}
          >
            {discount}
          </span>
        ) : null}
      </div>

      <div className="grid grid-cols-2 gap-2 text-xs" style={{ color: 'var(--dash-text-muted)' }}>
        <div>
          <span style={{ color: 'var(--dash-text-dim)' }}>Seats</span>
          <p className="font-semibold mt-0.5" style={{ color: 'var(--dash-text)' }}>
            {plan.maxSeats ? `Up to ${plan.maxSeats}` : 'Unlimited'}
          </p>
        </div>
        <div>
          <span style={{ color: 'var(--dash-text-dim)' }}>Locations</span>
          <p className="font-semibold mt-0.5" style={{ color: 'var(--dash-text)' }}>
            {plan.maxLocations ? `Up to ${plan.maxLocations}` : 'Unlimited'}
          </p>
        </div>
      </div>

      {plan.features.length > 0 ? (
        <ul className="space-y-1.5">
          {plan.features.slice(0, 4).map((feature) => (
            <li
              key={feature}
              className="flex items-start gap-2 text-xs"
              style={{ color: 'var(--dash-text-muted)' }}
            >
              <Check size={12} className="mt-0.5 shrink-0" style={{ color: 'var(--accent)' }} />
              <span>{feature}</span>
            </li>
          ))}
        </ul>
      ) : null}

      <div className="flex flex-wrap gap-1.5 pt-1">
        {plan.apps.map((app) => (
          <AppPill key={app} appId={app} />
        ))}
      </div>
    </div>
  );
};

const SoloPlanCard: React.FC<{
  plan: PlanPayload;
  isCurrent: boolean;
  isGranted?: boolean;
  ownerEmail: string;
  planActionsDisabled?: boolean;
}> = ({ plan, isCurrent, isGranted, ownerEmail, planActionsDisabled = false }) => {
  const app = plan.apps[0] ?? plan.slug;

  return (
    <div
      className="rounded-xl p-4 flex flex-col gap-3"
      style={{
        background: isCurrent ? 'var(--accent-muted)' : 'var(--dash-card)',
        border: isCurrent ? '1.5px solid var(--accent-border)' : '1px solid var(--dash-border)',
      }}
    >
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <h3 className="text-sm font-bold" style={{ color: 'var(--dash-text)' }}>
            {plan.name}
          </h3>
          <p className="text-xs mt-1" style={{ color: 'var(--dash-text-muted)' }}>
            {plan.tagline}
          </p>
        </div>
        <PlanAction
          plan={plan}
          isCurrent={isCurrent}
          isGranted={isGranted}
          ownerEmail={ownerEmail}
          disabled={planActionsDisabled}
          disabledLabel="Trial mode"
        />
      </div>
      <div className="flex items-center justify-between gap-3">
        <p className="text-lg font-bold" style={{ color: 'var(--dash-text)' }}>
          {plan.pricePerSeatCents === 0 ? 'Free' : centsToDisplay(plan.pricePerSeatCents)}
        </p>
        <AppPill appId={app} />
      </div>
    </div>
  );
};

const PlansSection: React.FC<{
  sub: SubscriptionPayload | null | undefined;
  billing: BillingDetailPayload | null | undefined;
  plans: PlanPayload[];
  loading: boolean;
}> = ({ sub, billing, plans, loading }) => {
  const [cycle, setCycle] = useState<BillingCycle>('monthly');
  const ownerEmail = useUserStore((s) => s.user?.email ?? '');
  const suitePlans = useMemo(
    () => plans.filter((plan) => !plan.isSolo).sort((a, b) => a.tierOrder - b.tierOrder),
    [plans]
  );
  const soloPlans = useMemo(
    () => plans.filter((plan) => plan.isSolo).sort((a, b) => a.tierOrder - b.tierOrder),
    [plans]
  );
  const currentTier = suitePlans.find((plan) => plan.slug === sub?.planSlug)?.tierOrder;
  const seats = billing?.currentSeatCount ?? sub?.seatCount ?? 1;
  const includedSeatLimit = billing?.includedSeatLimit ?? sub?.includedSeatLimit ?? 10;
  const isGranted = Boolean(sub?.isGranted ?? billing?.isGranted);
  const isCancelled = sub?.status === 'cancelled' || billing?.status === 'cancelled';
  const planActionsDisabled = sub?.status === 'trial' || billing?.status === 'trial';

  return (
    <>
      <section className="space-y-3">
        <SectionHeader
          title="Plan suite"
          description="Current plan selected, with monthly and yearly upgrade options."
          action={<CycleToggle cycle={cycle} onChange={setCycle} disabled={planActionsDisabled} />}
        />
        {planActionsDisabled ? (
          <div
            className="rounded-xl px-4 py-3 flex items-start gap-2 text-xs"
            style={{
              background: 'var(--status-trial-bg)',
              color: 'var(--status-trial)',
              border: '1px solid var(--dash-border)',
            }}
          >
            <AlertCircle size={14} className="mt-0.5 shrink-0" />
            <span>
              Plan changes are disabled while this account is in trial. Pricing is still being
              finalized for this test workspace.
            </span>
          </div>
        ) : null}
        {loading ? (
          <div className="grid grid-cols-1 xl:grid-cols-3 gap-3">
            {[1, 2, 3].map((i) => (
              <div
                key={i}
                className="h-72 rounded-xl animate-pulse"
                style={{ background: 'var(--dash-card)', border: '1px solid var(--dash-border)' }}
              />
            ))}
          </div>
        ) : (
          <div className="grid grid-cols-1 xl:grid-cols-3 gap-3">
            {suitePlans.map((plan) => (
              <SuitePlanCard
                key={plan.id}
                plan={plan}
                cycle={cycle}
                isCurrent={plan.slug === sub?.planSlug}
                isCancelled={isCancelled}
                isGranted={isGranted && plan.slug === sub?.planSlug}
                currentTier={currentTier}
                seats={seats}
                includedSeatLimit={includedSeatLimit}
                ownerEmail={ownerEmail}
                planActionsDisabled={planActionsDisabled}
              />
            ))}
          </div>
        )}
      </section>

      <section className="space-y-3">
        <SectionHeader
          title="Single app plans"
          description="Standalone app plans for teams that only need one workspace."
        />
        {loading ? (
          <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-3">
            {[1, 2, 3].map((i) => (
              <div
                key={i}
                className="h-32 rounded-xl animate-pulse"
                style={{ background: 'var(--dash-card)', border: '1px solid var(--dash-border)' }}
              />
            ))}
          </div>
        ) : soloPlans.length > 0 ? (
          <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-3">
            {soloPlans.map((plan) => (
              <SoloPlanCard
                key={plan.id}
                plan={plan}
                isCurrent={plan.slug === sub?.planSlug}
                isGranted={isGranted && plan.slug === sub?.planSlug}
                ownerEmail={ownerEmail}
                planActionsDisabled={planActionsDisabled}
              />
            ))}
          </div>
        ) : (
          <div
            className="rounded-xl p-5 text-sm"
            style={{ background: 'var(--dash-card)', border: '1px solid var(--dash-border)' }}
          >
            <p style={{ color: 'var(--dash-text-muted)' }}>No single app plans are available.</p>
          </div>
        )}
      </section>
    </>
  );
};

const PaymentMethodsSection: React.FC<{ sub: SubscriptionPayload | null | undefined }> = ({
  sub,
}) => {
  const ownerEmail = useUserStore((s) => s.user?.email ?? '');
  const checkout = useCheckout();
  const hasPayment = !!sub?.paddleSubscriptionId;

  const handleAddPayment = (): void => {
    if (!sub) {
      toast.info('Choose a plan first');
      return;
    }
    checkout.mutate(
      { planSlug: sub.planSlug, email: ownerEmail || undefined },
      {
        onSuccess: (data) => {
          navigateToCheckoutUrl(data.checkoutUrl);
        },
        onError: () => toast.error('Failed to start checkout. Please try again.'),
      }
    );
  };

  return (
    <section className="space-y-3">
      <SectionHeader title="Payment methods" description="Cards and billing collection." />
      <div
        className="rounded-xl p-5 flex flex-col sm:flex-row sm:items-center gap-4"
        style={{ background: 'var(--dash-card)', border: '1px solid var(--dash-border)' }}
      >
        <div
          className="h-11 w-11 rounded-xl flex items-center justify-center shrink-0"
          style={{ background: 'var(--dash-hover)' }}
        >
          <WalletCards size={18} style={{ color: 'var(--dash-text-muted)' }} />
        </div>
        <div className="min-w-0 flex-1">
          <p className="text-sm font-semibold" style={{ color: 'var(--dash-text)' }}>
            {hasPayment ? 'Payment method on file' : 'No payment method'}
          </p>
          <p className="text-xs mt-0.5" style={{ color: 'var(--dash-text-muted)' }}>
            {hasPayment
              ? 'Your card is managed securely by Paddle.'
              : 'Add a payment method when you activate or change a paid plan.'}
          </p>
        </div>
        {hasPayment ? (
          <span
            className="inline-flex shrink-0 items-center justify-center whitespace-nowrap px-3 py-1.5 rounded-lg text-xs font-semibold"
            style={{ background: 'var(--dash-hover)', color: 'var(--dash-text-muted)' }}
          >
            Managed by Paddle
          </span>
        ) : (
          <button
            type="button"
            onClick={handleAddPayment}
            disabled={checkout.isPending}
            className="inline-flex shrink-0 items-center justify-center gap-1.5 whitespace-nowrap px-3 py-1.5 rounded-lg text-xs font-semibold disabled:opacity-60"
            style={{ background: '#4583C4', color: '#fff' }}
          >
            {checkout.isPending ? (
              <Loader2 size={12} className="animate-spin" />
            ) : (
              <CreditCard size={12} />
            )}
            {checkout.isPending ? 'Opening...' : 'Add payment method'}
          </button>
        )}
      </div>
    </section>
  );
};

function seatLineDetail(item: BillingLineItem, isGranted?: boolean): string {
  if (isGranted) {
    return `${item.seatCount} active seat${item.seatCount !== 1 ? 's' : ''} · plan access granted`;
  }
  const split = splitIncludedSeats(item.seatCount, item.includedSeatLimit ?? 0);
  const included = item.includedSeatCount ?? split.included;
  const billable = item.billableSeatCount ?? split.billable;
  return `${included} included · ${billable} billable at ${centsToDisplay(item.pricePerSeatCents)}/seat/mo`;
}

function addonLineDetail(item: BillingLineItem): string {
  const base = `${centsToDisplay(item.unitPriceCents ?? 0)}/mo addon`;
  if (!item.isGranted) return base;
  if (!item.grantExpiresAt) return `${base} · granted`;
  return `${base} · granted until ${formatDate(item.grantExpiresAt)}`;
}

const CurrentChargesSection: React.FC<{
  billing: BillingDetailPayload | null | undefined;
  loading: boolean;
  onRefresh: () => void;
}> = ({ billing, loading, onRefresh }) => {
  const { t } = useTranslation();
  const isGranted = Boolean(billing?.isGranted);
  return (
    <section className="space-y-3">
      <SectionHeader title="Current charges" description="Plan seats and active paid addons." />
      <div
        className="rounded-xl overflow-hidden"
        style={{ background: 'var(--dash-card)', border: '1px solid var(--dash-border)' }}
      >
        {loading ? (
          <div className="p-5 space-y-3">
            {[1, 2, 3].map((i) => (
              <div
                key={i}
                className="h-12 rounded-lg animate-pulse"
                style={{ background: 'var(--dash-hover)' }}
              />
            ))}
          </div>
        ) : !billing || billing.lineItems.length === 0 ? (
          <EmptyState
            icon={ReceiptText}
            title={t('billing.emptyNoUsageTitle')}
            action={{ label: t('common.refresh'), onClick: onRefresh }}
            size="compact"
          />
        ) : (
          <div>
            {isGranted || billing.includedSeatLimit > 0 ? (
              <div
                className="grid grid-cols-1 sm:grid-cols-[1fr_auto] gap-2 sm:gap-4 px-5 py-4 items-center"
                style={{ borderBottom: '1px solid var(--dash-border)' }}
              >
                <div>
                  <p className="text-sm font-medium" style={{ color: 'var(--dash-text)' }}>
                    {isGranted ? 'Granted access' : 'Free tier allowance'}
                  </p>
                  <p className="text-xs mt-0.5" style={{ color: 'var(--dash-text-dim)' }}>
                    {isGranted
                      ? `${billing.planName} plan access is covered by your grant`
                      : `${billing.includedSeatCount} of ${billing.includedSeatLimit} seats included this month`}
                  </p>
                </div>
                <p
                  className="text-sm font-semibold sm:text-right"
                  style={{ color: 'var(--status-active)' }}
                >
                  {billing.seatDiscountCents > 0
                    ? `Saves ${centsToDisplay(billing.seatDiscountCents)}`
                    : 'Included'}
                </p>
              </div>
            ) : null}
            {billing.lineItems.map((item, index) => {
              const isAddon = item.type === 'addon';
              return (
                <div
                  key={`${item.type ?? 'seat'}-${item.addonId ?? index}-${item.periodStart}`}
                  className="grid grid-cols-1 sm:grid-cols-[1fr_auto] gap-2 sm:gap-4 px-5 py-4 items-center"
                  style={{ borderBottom: '1px solid var(--dash-border)' }}
                >
                  <div>
                    <p className="text-sm font-medium" style={{ color: 'var(--dash-text)' }}>
                      {isAddon
                        ? (item.description ?? 'Paid addon')
                        : `${item.seatCount} active seat${item.seatCount !== 1 ? 's' : ''}`}
                    </p>
                    <p className="text-xs mt-0.5" style={{ color: 'var(--dash-text-dim)' }}>
                      {isAddon ? addonLineDetail(item) : seatLineDetail(item, isGranted)}
                    </p>
                  </div>
                  <p
                    className="text-sm font-semibold sm:text-right"
                    style={{ color: 'var(--dash-text)' }}
                  >
                    {centsToDisplay(item.amountCents)}
                  </p>
                </div>
              );
            })}
            <div className="px-5 py-4 flex items-center justify-between gap-4">
              <p className="text-sm font-semibold" style={{ color: 'var(--dash-text)' }}>
                Estimated total
              </p>
              <p className="text-lg font-bold" style={{ color: 'var(--dash-text)' }}>
                {centsToDisplay(billing.totalCents)}
              </p>
            </div>
          </div>
        )}
      </div>
    </section>
  );
};

const InvoiceHistorySection: React.FC<{
  history: BillingHistoryMonth[];
  loading: boolean;
  onRefresh: () => void;
}> = ({ history, loading, onRefresh }) => {
  const { t } = useTranslation();
  return (
    <section className="space-y-3">
      <SectionHeader
        title="Invoice history"
        description="Monthly billing totals and invoice status."
      />
      <div
        className="rounded-xl overflow-hidden"
        style={{ background: 'var(--dash-card)', border: '1px solid var(--dash-border)' }}
      >
        {loading ? (
          <div className="p-5 space-y-3">
            {[1, 2, 3].map((i) => (
              <div
                key={i}
                className="h-12 rounded-lg animate-pulse"
                style={{ background: 'var(--dash-hover)' }}
              />
            ))}
          </div>
        ) : history.length === 0 ? (
          <EmptyState
            icon={Download}
            title={t('billing.emptyNoInvoicesTitle')}
            description={t('billing.emptyNoInvoicesDescription')}
            action={{ label: t('common.refresh'), onClick: onRefresh }}
          />
        ) : (
          <div>
            <div
              className="hidden sm:grid grid-cols-[1fr_auto_auto] gap-4 px-5 py-2.5 text-[11px] font-semibold uppercase tracking-wider"
              style={{
                color: 'var(--dash-text-dim)',
                borderBottom: '1px solid var(--dash-border)',
              }}
            >
              <span>Month</span>
              <span className="text-right">Status</span>
              <span className="text-right">Total</span>
            </div>
            {history.map((month, idx) => {
              const lineItems = month.lineItems ?? [];
              const seatItems = lineItems.filter((item: BillingLineItem) => item.type !== 'addon');
              const periodCount = seatItems.length;
              const averageSeats =
                periodCount > 0
                  ? Math.round(
                      seatItems.reduce((sum, item) => sum + item.seatCount, 0) / periodCount
                    )
                  : 0;
              const addonCount = lineItems.length - seatItems.length;

              return (
                <div
                  key={month.month || idx}
                  className="grid grid-cols-1 sm:grid-cols-[1fr_auto_auto] gap-2 sm:gap-4 px-5 py-4 items-center"
                  style={{ borderBottom: '1px solid var(--dash-border)' }}
                >
                  <div>
                    <p className="text-sm font-medium" style={{ color: 'var(--dash-text)' }}>
                      {formatMonth(month.month)}
                    </p>
                    <p className="text-xs mt-0.5" style={{ color: 'var(--dash-text-dim)' }}>
                      {averageSeats} avg seat{averageSeats !== 1 ? 's' : ''} · {periodCount} period
                      {periodCount !== 1 ? 's' : ''}
                      {addonCount > 0 ? ` · ${addonCount} addon${addonCount !== 1 ? 's' : ''}` : ''}
                    </p>
                  </div>
                  <span
                    className="w-fit sm:justify-self-end text-[10px] font-semibold px-2 py-0.5 rounded-full capitalize"
                    style={{
                      background:
                        month.status === 'paid'
                          ? 'var(--status-active-bg)'
                          : 'var(--status-trial-bg)',
                      color:
                        month.status === 'paid' ? 'var(--status-active)' : 'var(--status-trial)',
                    }}
                  >
                    {month.status}
                  </span>
                  <p
                    className="text-sm font-semibold sm:text-right"
                    style={{ color: 'var(--dash-text)' }}
                  >
                    {centsToDisplay(month.totalCents)}
                  </p>
                </div>
              );
            })}
          </div>
        )}
      </div>
    </section>
  );
};

const Billing: React.FC = (): React.JSX.Element => {
  const { data: sub, isLoading: subLoading } = useMySubscription();
  const { data: billing, isLoading: billingLoading, refetch: refetchBilling } = useBillingDetail();
  const { data: plans = [], isLoading: plansLoading } = useListPlans();
  const {
    data: history = [],
    isLoading: historyLoading,
    refetch: refetchHistory,
  } = useBillingHistory();

  return (
    <>
      <HelmetContainer title="Billing · rtylr" />
      <div className="w-full space-y-8">
        <div>
          <h1 className="text-xl font-bold" style={{ color: 'var(--dash-text)' }}>
            Billing
          </h1>
          <p className="text-sm mt-0.5" style={{ color: 'var(--dash-text-muted)' }}>
            Plans, payment methods, and invoice history.
          </p>
        </div>

        {subLoading || billingLoading ? (
          <section className="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-4 gap-3">
            {[1, 2, 3, 4].map((i) => (
              <div
                key={i}
                className="h-28 rounded-xl animate-pulse"
                style={{ background: 'var(--dash-card)', border: '1px solid var(--dash-border)' }}
              />
            ))}
          </section>
        ) : (
          <BillingSummary sub={sub} billing={billing} />
        )}

        <CurrentChargesSection
          billing={billing}
          loading={billingLoading}
          onRefresh={() => void refetchBilling()}
        />
        <PlansSection sub={sub} billing={billing} plans={plans} loading={plansLoading} />
        <PaymentMethodsSection sub={sub} />
        <InvoiceHistorySection
          history={history}
          loading={historyLoading}
          onRefresh={() => void refetchHistory()}
        />
      </div>
    </>
  );
};

export default Billing;
