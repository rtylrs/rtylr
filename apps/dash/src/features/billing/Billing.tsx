import React, { useMemo, useState } from 'react';
import { HelmetContainer } from '@voxire/components';
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
  const date = new Date(iso);
  return date.toLocaleDateString('en-US', { month: 'long', year: 'numeric', timeZone: 'UTC' });
}

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString('en-US', {
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
  const discount = Math.round((1 - plan.annualPricePerSeatCents / plan.pricePerSeatCents) * 100);
  return discount > 0 ? `Save ${discount}%` : null;
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
  const url = APP_URLS[id];

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
    <a href={url as string} target="_blank" rel="noreferrer noopener">
      {pill}
    </a>
  );
};

const CycleToggle: React.FC<{ cycle: BillingCycle; onChange: (cycle: BillingCycle) => void }> = ({
  cycle,
  onChange,
}) => (
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
          onClick={() => onChange(value)}
          className="px-3 py-1.5 rounded-md text-xs font-semibold transition-colors capitalize"
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

  return (
    <section className="space-y-3">
      <SectionHeader title="Overview" description="Current plan, spend, seats, and renewal." />
      <div className="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-4 gap-3">
        <SummaryTile
          label="Current plan"
          value={planName}
          detail={sub?.planSlug ?? 'Select a plan to activate billing'}
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
          detail={`${includedSeatCount} of ${includedSeatLimit} included · ${billableSeatCount} billable`}
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
  isCurrent: boolean;
  currentTier?: number;
  ownerEmail: string;
}> = ({ plan, isCurrent, currentTier, ownerEmail }) => {
  const checkout = useCheckout();

  const handleCheckout = (): void => {
    checkout.mutate(
      { planSlug: plan.slug, email: ownerEmail || undefined },
      {
        onSuccess: (data) => {
          window.location.href = data.checkoutUrl;
        },
        onError: () => {
          toast.error('Failed to start checkout. Please try again.');
        },
      }
    );
  };

  const actionLabel = useMemo(() => {
    if (isCurrent) return 'Current plan';
    if (currentTier === undefined) return 'Choose plan';
    return plan.tierOrder > currentTier ? 'Upgrade' : 'Switch';
  }, [currentTier, isCurrent, plan.tierOrder]);

  return (
    <button
      type="button"
      onClick={handleCheckout}
      disabled={isCurrent || checkout.isPending}
      className="inline-flex shrink-0 items-center justify-center gap-1.5 whitespace-nowrap px-3 py-1.5 rounded-lg text-xs font-semibold transition-opacity disabled:opacity-70"
      style={{
        background: isCurrent ? 'var(--dash-hover)' : '#4583C4',
        color: isCurrent ? 'var(--dash-text-muted)' : '#fff',
        border: isCurrent ? '1px solid var(--dash-border)' : '1px solid transparent',
      }}
    >
      {checkout.isPending ? <Loader2 size={12} className="animate-spin" /> : null}
      {checkout.isPending ? 'Opening...' : actionLabel}
      {!isCurrent && !checkout.isPending ? <ArrowRight size={12} className="shrink-0" /> : null}
    </button>
  );
};

const SuitePlanCard: React.FC<{
  plan: PlanPayload;
  cycle: BillingCycle;
  isCurrent: boolean;
  currentTier?: number;
  seats: number;
  includedSeatLimit: number;
  ownerEmail: string;
}> = ({ plan, cycle, isCurrent, currentTier, seats, includedSeatLimit, ownerEmail }) => {
  const price = planPrice(plan, cycle);
  const seatSplit = splitIncludedSeats(seats, includedSeatLimit);
  const monthlyTotal = price * seatSplit.billable;
  const discount = cycle === 'yearly' ? discountLabel(plan) : null;

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
          isCurrent={isCurrent}
          currentTier={currentTier}
          ownerEmail={ownerEmail}
        />
      </div>

      <div className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <p className="text-2xl font-bold" style={{ color: 'var(--dash-text)' }}>
            {price === 0 ? 'Free' : centsToDisplay(price)}
            {price > 0 ? (
              <span className="text-xs font-medium" style={{ color: 'var(--dash-text-dim)' }}>
                /seat/mo
              </span>
            ) : null}
          </p>
          <p className="text-xs mt-1" style={{ color: 'var(--dash-text-dim)' }}>
            {price === 0
              ? 'No monthly subscription cost'
              : `${centsToDisplay(monthlyTotal)}/mo after ${seatSplit.included} included seat${seatSplit.included !== 1 ? 's' : ''}`}
            {cycle === 'yearly' && price > 0 ? ' · billed yearly' : ''}
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
  ownerEmail: string;
}> = ({ plan, isCurrent, ownerEmail }) => {
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
        <PlanAction plan={plan} isCurrent={isCurrent} ownerEmail={ownerEmail} />
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

  return (
    <>
      <section className="space-y-3">
        <SectionHeader
          title="Plan suite"
          description="Current plan selected, with monthly and yearly upgrade options."
          action={<CycleToggle cycle={cycle} onChange={setCycle} />}
        />
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
                currentTier={currentTier}
                seats={seats}
                includedSeatLimit={includedSeatLimit}
                ownerEmail={ownerEmail}
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
                ownerEmail={ownerEmail}
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
          window.location.href = data.checkoutUrl;
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

function seatLineDetail(item: BillingLineItem): string {
  const split = splitIncludedSeats(item.seatCount, item.includedSeatLimit ?? 0);
  const included = item.includedSeatCount ?? split.included;
  const billable = item.billableSeatCount ?? split.billable;
  return `${included} included · ${billable} billable at ${centsToDisplay(item.pricePerSeatCents)}/seat/mo`;
}

const CurrentChargesSection: React.FC<{
  billing: BillingDetailPayload | null | undefined;
  loading: boolean;
}> = ({ billing, loading }) => (
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
        <div className="px-5 py-8 text-sm" style={{ color: 'var(--dash-text-muted)' }}>
          No billable usage recorded this month.
        </div>
      ) : (
        <div>
          {billing.includedSeatLimit > 0 ? (
            <div
              className="grid grid-cols-1 sm:grid-cols-[1fr_auto] gap-2 sm:gap-4 px-5 py-4 items-center"
              style={{ borderBottom: '1px solid var(--dash-border)' }}
            >
              <div>
                <p className="text-sm font-medium" style={{ color: 'var(--dash-text)' }}>
                  Free tier allowance
                </p>
                <p className="text-xs mt-0.5" style={{ color: 'var(--dash-text-dim)' }}>
                  {billing.includedSeatCount} of {billing.includedSeatLimit} seats included this
                  month
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
                    {isAddon
                      ? `${centsToDisplay(item.unitPriceCents ?? 0)}/mo addon`
                      : seatLineDetail(item)}
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

const InvoiceHistorySection: React.FC<{ history: BillingHistoryMonth[]; loading: boolean }> = ({
  history,
  loading,
}) => (
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
        <div className="px-5 py-12 flex flex-col items-center text-center">
          <Download size={28} style={{ color: 'var(--dash-text-dim)', marginBottom: 12 }} />
          <p className="text-sm font-medium" style={{ color: 'var(--dash-text-muted)' }}>
            No invoices yet
          </p>
          <p className="text-xs mt-1" style={{ color: 'var(--dash-text-dim)' }}>
            Your first invoice will appear after a billing cycle closes.
          </p>
        </div>
      ) : (
        <div>
          <div
            className="hidden sm:grid grid-cols-[1fr_auto_auto] gap-4 px-5 py-2.5 text-[11px] font-semibold uppercase tracking-wider"
            style={{ color: 'var(--dash-text-dim)', borderBottom: '1px solid var(--dash-border)' }}
          >
            <span>Month</span>
            <span className="text-right">Status</span>
            <span className="text-right">Total</span>
          </div>
          {history.map((month) => {
            const seatItems = month.lineItems.filter(
              (item: BillingLineItem) => item.type !== 'addon'
            );
            const periodCount = seatItems.length;
            const averageSeats =
              periodCount > 0
                ? Math.round(seatItems.reduce((sum, item) => sum + item.seatCount, 0) / periodCount)
                : 0;
            const addonCount = month.lineItems.length - seatItems.length;

            return (
              <div
                key={month.month}
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
                    color: month.status === 'paid' ? 'var(--status-active)' : 'var(--status-trial)',
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

const Billing: React.FC = (): React.JSX.Element => {
  const { data: sub, isLoading: subLoading } = useMySubscription();
  const { data: billing, isLoading: billingLoading } = useBillingDetail();
  const { data: plans = [], isLoading: plansLoading } = useListPlans();
  const { data: history = [], isLoading: historyLoading } = useBillingHistory();

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

        <CurrentChargesSection billing={billing} loading={billingLoading} />
        <PlansSection sub={sub} billing={billing} plans={plans} loading={plansLoading} />
        <PaymentMethodsSection sub={sub} />
        <InvoiceHistorySection history={history} loading={historyLoading} />
      </div>
    </>
  );
};

export default Billing;
