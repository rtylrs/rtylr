import React from 'react';
import {
  CheckCircle2,
  Clock,
  XCircle,
  AlertCircle,
  Users,
  Zap,
  ArrowRight,
  Loader2,
} from 'lucide-react';
import { toast } from 'sonner';
import { APP_CONFIG, APP_URLS } from '@voxire/config';
import type { AppId } from '@voxire/config';
import { SettingSection, Breadcrumbs } from '../../shared/components/dash';
import { useUserStore } from '../../shared/stores/userStore';
import {
  useMySubscription,
  useListPlans,
  useBillingDetail,
  useCheckout,
  centsToDisplay,
  trialDaysLeft,
  type SubscriptionPayload,
  type PlanPayload,
} from '../../shared/hooks/queries/useSubscription';

// ── Status badge ──────────────────────────────────────────────────────────────

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

// ── App pill ──────────────────────────────────────────────────────────────────

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

  if (url) {
    return (
      <a href={url as string} target="_blank" rel="noreferrer noopener">
        {pill}
      </a>
    );
  }
  return pill;
};

// ── Current plan + seat cost ──────────────────────────────────────────────────

const CurrentPlanCard: React.FC<{ sub: SubscriptionPayload }> = ({ sub }) => {
  const { data: billing } = useBillingDetail();
  const daysLeft = trialDaysLeft(sub.trialEndsAt);
  const isTrial = sub.status === 'trial';
  const estimatedMonthly = sub.seatCount * sub.pricePerSeatCents;

  return (
    <SettingSection title="Current plan" description="Your active subscription and seat usage.">
      <div className="space-y-5">
        {/* Plan + status row */}
        <div className="flex items-start justify-between gap-4">
          <div>
            <p className="text-base font-bold" style={{ color: 'var(--dash-text)' }}>
              {sub.planName}
            </p>
            <p className="text-xs mt-0.5 font-mono" style={{ color: 'var(--dash-text-dim)' }}>
              {sub.planSlug}
            </p>
          </div>
          <StatusBadge status={sub.status} />
        </div>

        {/* Per-seat pricing block */}
        <div
          className="rounded-xl p-4 space-y-3"
          style={{ background: 'var(--dash-bg)', border: '1px solid var(--dash-border)' }}
        >
          <div className="flex items-center gap-2">
            <Users size={13} style={{ color: 'var(--dash-text-muted)' }} />
            <p className="text-xs font-semibold" style={{ color: 'var(--dash-text-muted)' }}>
              Usage this month
            </p>
          </div>
          <div className="flex items-end justify-between">
            <div>
              <p className="text-2xl font-bold" style={{ color: 'var(--dash-text)' }}>
                {centsToDisplay(billing?.totalCents ?? estimatedMonthly)}
              </p>
              <p className="text-xs mt-0.5" style={{ color: 'var(--dash-text-dim)' }}>
                {sub.seatCount} seat{sub.seatCount !== 1 ? 's' : ''} ×{' '}
                {centsToDisplay(sub.pricePerSeatCents)}/seat · billed 1st of each month
              </p>
            </div>
            {isTrial && (
              <div
                className="text-xs font-semibold px-2.5 py-1 rounded-lg"
                style={{ background: 'var(--status-trial-bg)', color: 'var(--status-trial)' }}
              >
                {daysLeft}d left in trial
              </div>
            )}
          </div>
        </div>

        {/* Included apps */}
        <div>
          <p className="text-xs font-medium mb-2" style={{ color: 'var(--dash-text-muted)' }}>
            Included apps
          </p>
          <div className="flex flex-wrap gap-1.5">
            {sub.apps.map((app) => (
              <AppPill key={app} appId={app} />
            ))}
          </div>
        </div>
      </div>
    </SettingSection>
  );
};

// ── Plan card ─────────────────────────────────────────────────────────────────

const PlanCard: React.FC<{
  plan: PlanPayload;
  isCurrent: boolean;
  currentSeats: number;
  ownerEmail: string;
}> = ({ plan, isCurrent, currentSeats, ownerEmail }) => {
  const estimatedMonthly = currentSeats * plan.pricePerSeatCents;
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

  return (
    <div
      className="rounded-xl p-4 space-y-3"
      style={{
        border: isCurrent ? '1.5px solid var(--accent-border)' : '1px solid var(--dash-border)',
        background: isCurrent ? 'var(--accent-muted)' : 'var(--dash-card)',
      }}
    >
      <div className="flex items-center justify-between gap-2">
        <div>
          <p className="text-sm font-semibold" style={{ color: 'var(--dash-text)' }}>
            {plan.name}
          </p>
          <p className="text-xs mt-0.5" style={{ color: 'var(--dash-text-dim)' }}>
            {centsToDisplay(plan.pricePerSeatCents)}/seat/month
          </p>
        </div>
        <div className="text-right">
          {isCurrent ? (
            <span
              className="text-[10px] font-bold px-1.5 py-0.5 rounded"
              style={{ background: 'var(--accent)', color: '#fff' }}
            >
              Current
            </span>
          ) : (
            <p className="text-xs font-medium" style={{ color: 'var(--dash-text-dim)' }}>
              ~{centsToDisplay(estimatedMonthly)}/mo at {currentSeats} seats
            </p>
          )}
        </div>
      </div>

      <div className="flex flex-wrap gap-1">
        {plan.apps.map((app) => (
          <AppPill key={app} appId={app} />
        ))}
      </div>

      {!isCurrent && (
        <button
          onClick={handleCheckout}
          disabled={checkout.isPending}
          className="flex items-center gap-1.5 text-xs font-semibold transition-opacity hover:opacity-80 disabled:opacity-50"
          style={{ color: 'var(--accent)' }}
        >
          {checkout.isPending ? (
            <Loader2 size={11} className="animate-spin" />
          ) : (
            <ArrowRight size={11} />
          )}
          {checkout.isPending ? 'Opening checkout…' : `Switch to ${plan.name}`}
        </button>
      )}
    </div>
  );
};

// ── Main ──────────────────────────────────────────────────────────────────────

const Subscription: React.FC = (): React.JSX.Element => {
  const { data: sub, isLoading: subLoading } = useMySubscription();
  const { data: plans = [], isLoading: plansLoading } = useListPlans();
  const ownerEmail = useUserStore((s) => s.user?.email ?? s.user?.emailAddress ?? '');

  return (
    <div className="space-y-6 max-w-2xl">
      <Breadcrumbs />

      {/* Current plan */}
      {subLoading ? (
        <div
          className="h-40 rounded-xl animate-pulse"
          style={{ background: 'var(--dash-card)', border: '1px solid var(--dash-border)' }}
        />
      ) : sub ? (
        <CurrentPlanCard sub={sub} />
      ) : (
        <SettingSection title="Current plan" description="You don't have an active subscription.">
          <button
            className="flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-semibold"
            style={{ background: '#4583C4', color: '#fff' }}
          >
            <Zap size={13} />
            Choose a plan
          </button>
        </SettingSection>
      )}

      {/* All plans */}
      <SettingSection
        title="Available plans"
        description="Priced per seat. Switch plans any time, changes apply at the next billing cycle."
      >
        {plansLoading ? (
          <div className="space-y-3">
            {[1, 2, 3].map((i) => (
              <div
                key={i}
                className="h-20 rounded-xl animate-pulse"
                style={{ background: 'var(--dash-hover)' }}
              />
            ))}
          </div>
        ) : (
          <div className="space-y-3">
            {plans.map((plan) => (
              <PlanCard
                key={plan.id}
                plan={plan}
                isCurrent={plan.slug === sub?.planSlug}
                currentSeats={sub?.seatCount ?? 1}
                ownerEmail={ownerEmail}
              />
            ))}
          </div>
        )}
      </SettingSection>
    </div>
  );
};

export default Subscription;
