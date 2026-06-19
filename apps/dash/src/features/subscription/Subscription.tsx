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
  Check,
  Sparkles,
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

function splitIncludedSeats(
  seats: number,
  includedLimit: number
): { included: number; billable: number } {
  const safeSeats = Math.max(0, seats);
  const safeLimit = Math.max(0, includedLimit);
  const included = Math.min(safeSeats, safeLimit);
  return { included, billable: Math.max(0, safeSeats - included) };
}

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
  const activeSeats = billing?.currentSeatCount ?? sub.seatCount;
  const includedLimit = billing?.includedSeatLimit ?? sub.includedSeatLimit ?? 10;
  const seatSplit = splitIncludedSeats(activeSeats, includedLimit);
  const billableSeats = billing?.billableSeatCount ?? sub.billableSeatCount ?? seatSplit.billable;
  const includedSeats = billing?.includedSeatCount ?? sub.includedSeatCount ?? seatSplit.included;
  const estimatedMonthly = billableSeats * sub.pricePerSeatCents;

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
                {activeSeats} active seat{activeSeats !== 1 ? 's' : ''} · {includedSeats} of{' '}
                {includedLimit} included · {billableSeats} billable
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
  includedSeatLimit: number;
  ownerEmail: string;
  compact?: boolean;
}> = ({ plan, isCurrent, currentSeats, includedSeatLimit, ownerEmail, compact }) => {
  const seatSplit = splitIncludedSeats(currentSeats, includedSeatLimit);
  const estimatedMonthly = seatSplit.billable * plan.pricePerSeatCents;
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

  const priceLabel =
    plan.pricePerSeatCents === 0 ? 'Free' : `${centsToDisplay(plan.pricePerSeatCents)}/seat/mo`;

  return (
    <div
      className="rounded-xl p-4 space-y-3 relative"
      style={{
        border:
          isCurrent || plan.isFeatured
            ? '1.5px solid var(--accent-border)'
            : '1px solid var(--dash-border)',
        background: isCurrent ? 'var(--accent-muted)' : 'var(--dash-card)',
      }}
    >
      {plan.isFeatured && !isCurrent && (
        <span
          className="absolute -top-2 right-3 text-[10px] font-bold px-2 py-0.5 rounded-full inline-flex items-center gap-1"
          style={{ background: 'var(--accent)', color: '#fff' }}
        >
          <Sparkles size={9} />
          Popular
        </span>
      )}

      <div className="flex items-start justify-between gap-2">
        <div className="min-w-0">
          <p className="text-sm font-semibold" style={{ color: 'var(--dash-text)' }}>
            {plan.name}
          </p>
          <p className="text-xs mt-0.5" style={{ color: 'var(--dash-text-dim)' }}>
            {priceLabel}
          </p>
          {!compact && plan.tagline && (
            <p className="text-xs mt-1.5" style={{ color: 'var(--dash-text-muted)' }}>
              {plan.tagline}
            </p>
          )}
        </div>
        {isCurrent && (
          <span
            className="text-[10px] font-bold px-1.5 py-0.5 rounded shrink-0"
            style={{ background: 'var(--accent)', color: '#fff' }}
          >
            Current
          </span>
        )}
      </div>

      {!compact && plan.features && plan.features.length > 0 && (
        <ul className="space-y-1.5 pt-1">
          {plan.features.slice(0, 5).map((f, i) => (
            <li
              key={i}
              className="flex items-start gap-2 text-xs"
              style={{ color: 'var(--dash-text-muted)' }}
            >
              <Check
                size={12}
                className="mt-0.5 shrink-0"
                style={{ color: 'var(--accent)' }}
                strokeWidth={3}
              />
              <span>{f}</span>
            </li>
          ))}
        </ul>
      )}

      <div className="flex flex-wrap gap-1 pt-1">
        {plan.apps.map((app) => (
          <AppPill key={app} appId={app} />
        ))}
      </div>

      {!isCurrent && (
        <div className="flex items-center justify-between pt-2">
          <p className="text-xs font-medium" style={{ color: 'var(--dash-text-dim)' }}>
            {plan.pricePerSeatCents === 0
              ? 'Free forever'
              : `~${centsToDisplay(estimatedMonthly)}/mo after ${seatSplit.included} included seat${seatSplit.included !== 1 ? 's' : ''}`}
          </p>
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
            {checkout.isPending ? 'Opening…' : `Switch to ${plan.name}`}
          </button>
        </div>
      )}
    </div>
  );
};

// ── Main ──────────────────────────────────────────────────────────────────────

const Subscription: React.FC = (): React.JSX.Element => {
  const { data: sub, isLoading: subLoading } = useMySubscription();
  const { data: plans = [], isLoading: plansLoading } = useListPlans();
  const ownerEmail = useUserStore((s) => s.user?.email ?? '');

  const tierPlans = [...plans].filter((p) => !p.isSolo).sort((a, b) => a.tierOrder - b.tierOrder);
  const soloPlans = [...plans].filter((p) => p.isSolo).sort((a, b) => a.tierOrder - b.tierOrder);
  const currentSeats = sub?.seatCount ?? 1;
  const includedSeatLimit = sub?.includedSeatLimit ?? 10;

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

      {/* Tier plans */}
      <SettingSection
        title="Full suite plans"
        description="Every app, one subscription. Priced per seat - only pay for what you use."
      >
        {plansLoading ? (
          <div className="space-y-3">
            {[1, 2, 3].map((i) => (
              <div
                key={i}
                className="h-32 rounded-xl animate-pulse"
                style={{ background: 'var(--dash-hover)' }}
              />
            ))}
          </div>
        ) : (
          <div className="space-y-3">
            {tierPlans.map((plan) => (
              <PlanCard
                key={plan.id}
                plan={plan}
                isCurrent={plan.slug === sub?.planSlug}
                currentSeats={currentSeats}
                includedSeatLimit={includedSeatLimit}
                ownerEmail={ownerEmail}
              />
            ))}
          </div>
        )}
      </SettingSection>

      {/* Solo plans */}
      {soloPlans.length > 0 && (
        <SettingSection
          title="Single-app plans"
          description="Only need one tool? Start with a solo plan - upgrade to the suite anytime."
        >
          <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
            {soloPlans.map((plan) => (
              <PlanCard
                key={plan.id}
                plan={plan}
                isCurrent={plan.slug === sub?.planSlug}
                currentSeats={currentSeats}
                includedSeatLimit={includedSeatLimit}
                ownerEmail={ownerEmail}
                compact
              />
            ))}
          </div>
        </SettingSection>
      )}
    </div>
  );
};

export default Subscription;
