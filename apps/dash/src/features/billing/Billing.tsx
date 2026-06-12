import React from 'react';
import { HelmetContainer } from '@voxire/components';
import {
  AlertCircle,
  ArrowRight,
  Check,
  CheckCircle2,
  Clock,
  CreditCard,
  Download,
  Loader2,
  Sparkles,
  Users,
  XCircle,
  Zap,
} from 'lucide-react';
import { toast } from 'sonner';
import { APP_CONFIG, APP_URLS } from '@voxire/config';
import type { AppId } from '@voxire/config';
import { Breadcrumbs, SettingSection } from '../../shared/components/dash';
import { useUserStore } from '../../shared/stores/userStore';
import {
  useBillingDetail,
  useBillingHistory,
  useCheckout,
  useListPlans,
  useMySubscription,
  centsToDisplay,
  formatPeriodRange,
  trialDaysLeft,
  type BillingLineItem,
  type PlanPayload,
  type SubscriptionPayload,
} from '../../shared/hooks/queries/useSubscription';

function formatMonth(iso: string): string {
  const d = new Date(iso);
  return d.toLocaleDateString('en-US', { month: 'long', year: 'numeric', timeZone: 'UTC' });
}

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

const CurrentPlanCard: React.FC<{ sub: SubscriptionPayload }> = ({ sub }) => {
  const { data: billing } = useBillingDetail();
  const daysLeft = trialDaysLeft(sub.trialEndsAt);
  const isTrial = sub.status === 'trial';
  const estimatedMonthly = sub.seatCount * sub.pricePerSeatCents;

  return (
    <SettingSection
      title="Subscription"
      description="Your active plan, seat usage, and included apps."
    >
      <div className="space-y-5">
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
          <div className="flex items-end justify-between gap-4">
            <div>
              <p className="text-2xl font-bold" style={{ color: 'var(--dash-text)' }}>
                {centsToDisplay(billing?.totalCents ?? estimatedMonthly)}
              </p>
              <p className="text-xs mt-0.5" style={{ color: 'var(--dash-text-dim)' }}>
                {sub.seatCount} seat{sub.seatCount !== 1 ? 's' : ''} ×{' '}
                {centsToDisplay(sub.pricePerSeatCents)}/seat · billed monthly
              </p>
            </div>
            {isTrial ? (
              <div
                className="text-xs font-semibold px-2.5 py-1 rounded-lg"
                style={{ background: 'var(--status-trial-bg)', color: 'var(--status-trial)' }}
              >
                {daysLeft}d left in trial
              </div>
            ) : null}
          </div>
        </div>

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

const PlanCard: React.FC<{
  plan: PlanPayload;
  isCurrent: boolean;
  currentSeats: number;
  ownerEmail: string;
  compact?: boolean;
}> = ({ plan, isCurrent, currentSeats, ownerEmail, compact }) => {
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
      {plan.isFeatured && !isCurrent ? (
        <span
          className="absolute -top-2 right-3 text-[10px] font-bold px-2 py-0.5 rounded-full inline-flex items-center gap-1"
          style={{ background: 'var(--accent)', color: '#fff' }}
        >
          <Sparkles size={9} />
          Popular
        </span>
      ) : null}

      <div className="flex items-start justify-between gap-2">
        <div className="min-w-0">
          <p className="text-sm font-semibold" style={{ color: 'var(--dash-text)' }}>
            {plan.name}
          </p>
          <p className="text-xs mt-0.5" style={{ color: 'var(--dash-text-dim)' }}>
            {priceLabel}
          </p>
          {!compact && plan.tagline ? (
            <p className="text-xs mt-1.5" style={{ color: 'var(--dash-text-muted)' }}>
              {plan.tagline}
            </p>
          ) : null}
        </div>
        {isCurrent ? (
          <span
            className="text-[10px] font-bold px-1.5 py-0.5 rounded shrink-0"
            style={{ background: 'var(--accent)', color: '#fff' }}
          >
            Current
          </span>
        ) : null}
      </div>

      {!compact && plan.features && plan.features.length > 0 ? (
        <ul className="space-y-1.5 pt-1">
          {plan.features.slice(0, 5).map((feature, i) => (
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
              <span>{feature}</span>
            </li>
          ))}
        </ul>
      ) : null}

      <div className="flex flex-wrap gap-1 pt-1">
        {plan.apps.map((app) => (
          <AppPill key={app} appId={app} />
        ))}
      </div>

      {!isCurrent ? (
        <div className="flex items-center justify-between gap-3 pt-2">
          <p className="text-xs font-medium" style={{ color: 'var(--dash-text-dim)' }}>
            {plan.pricePerSeatCents === 0
              ? 'Free forever'
              : `~${centsToDisplay(estimatedMonthly)}/mo at ${currentSeats} seat${currentSeats !== 1 ? 's' : ''}`}
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
      ) : null}
    </div>
  );
};

const PlanSelectionSection: React.FC = () => {
  const { data: sub, isLoading: subLoading } = useMySubscription();
  const { data: plans = [], isLoading: plansLoading } = useListPlans();
  const ownerEmail = useUserStore((s) => s.user?.email ?? '');

  const tierPlans = [...plans].filter((p) => !p.isSolo).sort((a, b) => a.tierOrder - b.tierOrder);
  const soloPlans = [...plans].filter((p) => p.isSolo).sort((a, b) => a.tierOrder - b.tierOrder);

  return (
    <>
      {subLoading ? (
        <div
          className="h-40 rounded-xl animate-pulse"
          style={{ background: 'var(--dash-card)', border: '1px solid var(--dash-border)' }}
        />
      ) : sub ? (
        <CurrentPlanCard sub={sub} />
      ) : (
        <SettingSection title="Subscription" description="You don't have an active subscription.">
          <button
            className="flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-semibold"
            style={{ background: '#4583C4', color: '#fff' }}
          >
            <Zap size={13} />
            Choose a plan
          </button>
        </SettingSection>
      )}

      <SettingSection
        title="Full suite plans"
        description="Every app, one subscription. Priced per seat so the whole workspace scales together."
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
                currentSeats={sub?.seatCount ?? 1}
                ownerEmail={ownerEmail}
              />
            ))}
          </div>
        )}
      </SettingSection>

      {soloPlans.length > 0 ? (
        <SettingSection
          title="Single-app plans"
          description="Use one product on its own, then move up to the full suite when you need more."
        >
          <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
            {soloPlans.map((plan) => (
              <PlanCard
                key={plan.id}
                plan={plan}
                isCurrent={plan.slug === sub?.planSlug}
                currentSeats={sub?.seatCount ?? 1}
                ownerEmail={ownerEmail}
                compact
              />
            ))}
          </div>
        </SettingSection>
      ) : null}
    </>
  );
};

const LineItemRow: React.FC<{ item: BillingLineItem; dim?: boolean }> = ({ item, dim }) => (
  <div
    className="grid items-center gap-4 py-3.5"
    style={{
      gridTemplateColumns: '1fr auto auto',
      borderBottom: '1px solid var(--dash-border)',
      opacity: dim ? 0.6 : 1,
    }}
  >
    <div>
      <p className="text-sm" style={{ color: 'var(--dash-text)' }}>
        {item.seatCount} seat{item.seatCount !== 1 ? 's' : ''} ·{' '}
        {centsToDisplay(item.pricePerSeatCents)}/seat/month
      </p>
      <p className="text-xs mt-0.5 font-mono" style={{ color: 'var(--dash-text-dim)' }}>
        {formatPeriodRange(item.periodStart, item.periodEnd)}
        {!item.periodEnd ? (
          <span
            className="ml-2 text-[10px] font-sans font-semibold px-1.5 py-0.5 rounded-full"
            style={{ background: 'var(--status-trial-bg)', color: 'var(--status-trial)' }}
          >
            in progress
          </span>
        ) : null}
      </p>
    </div>
    <p
      className="text-xs font-mono text-right"
      style={{ color: 'var(--dash-text-muted)', minWidth: 72 }}
    >
      {item.seatCount} × {centsToDisplay(item.pricePerSeatCents)}
    </p>
    <p
      className="text-sm font-semibold text-right"
      style={{ color: 'var(--dash-text)', minWidth: 64 }}
    >
      {centsToDisplay(item.amountCents)}
    </p>
  </div>
);

const CurrentMonthCard: React.FC = () => {
  const { data: billing, isLoading } = useBillingDetail();
  const { data: sub } = useMySubscription();

  if (isLoading) {
    return (
      <div
        className="rounded-xl h-48 animate-pulse"
        style={{ background: 'var(--dash-card)', border: '1px solid var(--dash-border)' }}
      />
    );
  }

  const daysLeft = trialDaysLeft(sub?.trialEndsAt ?? null);
  const isTrial = sub?.status === 'trial';

  return (
    <div
      className="rounded-xl overflow-hidden"
      style={{ background: 'var(--dash-card)', border: '1px solid var(--dash-border)' }}
    >
      <div
        className="px-5 py-4 flex items-center justify-between gap-4"
        style={{ borderBottom: '1px solid var(--dash-border)' }}
      >
        <div>
          <p
            className="text-xs font-semibold uppercase tracking-widest"
            style={{ color: 'var(--dash-text-dim)' }}
          >
            {billing ? formatMonth(billing.billingMonth) : 'Current month'}
          </p>
          <p className="text-sm font-semibold mt-0.5" style={{ color: 'var(--dash-text)' }}>
            {billing?.planName ?? 'No plan'}
          </p>
        </div>

        {isTrial ? (
          <div
            className="flex items-center gap-1.5 px-3 py-1.5 rounded-full text-xs font-semibold"
            style={{ background: 'var(--status-trial-bg)', color: 'var(--status-trial)' }}
          >
            <Clock size={11} />
            Trial · {daysLeft} day{daysLeft !== 1 ? 's' : ''} left
          </div>
        ) : billing?.nextBillingDate ? (
          <p className="text-xs" style={{ color: 'var(--dash-text-muted)' }}>
            Next bill:{' '}
            <span style={{ color: 'var(--dash-text)' }}>
              {new Date(billing.nextBillingDate).toLocaleDateString('en-US', {
                month: 'short',
                day: 'numeric',
                year: 'numeric',
                timeZone: 'UTC',
              })}
            </span>
          </p>
        ) : null}
      </div>

      <div className="px-5">
        {!billing || billing.lineItems?.length === 0 ? (
          <div className="py-8 text-center">
            <p className="text-sm" style={{ color: 'var(--dash-text-muted)' }}>
              No usage recorded yet this month.
            </p>
          </div>
        ) : (
          <>
            <div
              className="grid gap-4 py-2.5 text-[11px] font-semibold uppercase tracking-wider"
              style={{
                gridTemplateColumns: '1fr auto auto',
                color: 'var(--dash-text-dim)',
                borderBottom: '1px solid var(--dash-border)',
              }}
            >
              <span>Description</span>
              <span className="text-right" style={{ minWidth: 72 }}>
                Rate
              </span>
              <span className="text-right" style={{ minWidth: 64 }}>
                Amount (USD)
              </span>
            </div>

            {billing.lineItems.map((item, i) => (
              <LineItemRow key={i} item={item} />
            ))}
          </>
        )}
      </div>

      <div
        className="px-5 py-4 flex items-center justify-between"
        style={{ borderTop: '1px solid var(--dash-border)' }}
      >
        <p className="text-sm font-semibold" style={{ color: 'var(--dash-text)' }}>
          {isTrial ? 'Estimated total (trial, not charged)' : 'Estimated total'}
        </p>
        <p className="text-xl font-bold" style={{ color: 'var(--dash-text)' }}>
          {billing ? centsToDisplay(billing.totalCents) : '$0.00'}
        </p>
      </div>
    </div>
  );
};

const PaymentMethodCard: React.FC = () => {
  const { data: sub } = useMySubscription();
  const hasPayment = !!sub?.paddleSubscriptionId;

  return (
    <div
      className="rounded-xl"
      style={{ background: 'var(--dash-card)', border: '1px solid var(--dash-border)' }}
    >
      <div className="px-5 py-4" style={{ borderBottom: '1px solid var(--dash-border)' }}>
        <p className="text-sm font-semibold" style={{ color: 'var(--dash-text)' }}>
          Payment method
        </p>
      </div>
      <div className="px-5 py-6">
        {hasPayment ? (
          <div className="flex items-center gap-3">
            <div
              className="h-10 w-10 rounded-lg flex items-center justify-center"
              style={{ background: 'var(--dash-hover)' }}
            >
              <CreditCard size={16} style={{ color: 'var(--dash-text-muted)' }} />
            </div>
            <div>
              <p className="text-sm font-medium" style={{ color: 'var(--dash-text)' }}>
                Managed by Paddle
              </p>
              <p className="text-xs mt-0.5" style={{ color: 'var(--dash-text-muted)' }}>
                Update your payment method in the billing portal
              </p>
            </div>
            <button
              className="ml-auto px-3 py-1.5 rounded-lg text-xs font-semibold transition-colors"
              style={{
                background: 'var(--dash-hover)',
                color: 'var(--dash-text)',
                border: '1px solid var(--dash-border)',
              }}
            >
              Manage
            </button>
          </div>
        ) : (
          <div className="flex flex-col items-center py-4 text-center">
            <div
              className="h-12 w-12 rounded-xl flex items-center justify-center mb-3"
              style={{ background: 'var(--dash-hover)' }}
            >
              <CreditCard size={20} style={{ color: 'var(--dash-text-dim)' }} />
            </div>
            <p className="text-sm font-medium mb-1" style={{ color: 'var(--dash-text)' }}>
              No payment method
            </p>
            <p className="text-xs mb-4" style={{ color: 'var(--dash-text-muted)' }}>
              Add a card to activate your plan after the trial.
            </p>
            <button
              className="flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-semibold"
              style={{ background: '#4583C4', color: '#fff' }}
            >
              <Zap size={13} />
              Activate plan
            </button>
          </div>
        )}
      </div>
    </div>
  );
};

const BillingHistoryCard: React.FC = () => {
  const { data: history = [], isLoading } = useBillingHistory();

  return (
    <div
      className="rounded-xl"
      style={{ background: 'var(--dash-card)', border: '1px solid var(--dash-border)' }}
    >
      <div className="px-5 py-4" style={{ borderBottom: '1px solid var(--dash-border)' }}>
        <p className="text-sm font-semibold" style={{ color: 'var(--dash-text)' }}>
          Invoice history
        </p>
      </div>

      {isLoading ? (
        <div className="px-5 py-8 space-y-3">
          {[1, 2, 3].map((i) => (
            <div
              key={i}
              className="h-14 rounded-lg animate-pulse"
              style={{ background: 'var(--dash-hover)' }}
            />
          ))}
        </div>
      ) : history.length === 0 ? (
        <div className="px-5 py-10 flex flex-col items-center text-center">
          <Download size={28} style={{ color: 'var(--dash-text-dim)', marginBottom: 12 }} />
          <p className="text-sm font-medium" style={{ color: 'var(--dash-text-muted)' }}>
            No invoices yet
          </p>
          <p className="text-xs mt-1" style={{ color: 'var(--dash-text-dim)' }}>
            Your first invoice will appear here after the billing cycle closes.
          </p>
        </div>
      ) : (
        <div className="divide-y" style={{ '--tw-divide-opacity': 1 } as React.CSSProperties}>
          {history.map((month, i) => (
            <div key={i} className="px-5 py-3.5 flex items-center gap-4">
              <div className="flex-1">
                <p className="text-sm font-medium" style={{ color: 'var(--dash-text)' }}>
                  {formatMonth(month.month)}
                </p>
                <p className="text-xs mt-0.5" style={{ color: 'var(--dash-text-dim)' }}>
                  {month.lineItems.reduce((sum, li) => sum + li.seatCount, 0) /
                    Math.max(1, month.lineItems.length)}{' '}
                  avg seats · {month.lineItems.length} period
                  {month.lineItems.length !== 1 ? 's' : ''}
                </p>
              </div>
              <p className="text-sm font-semibold" style={{ color: 'var(--dash-text)' }}>
                {centsToDisplay(month.totalCents)}
              </p>
              <span
                className="text-[10px] font-semibold px-2 py-0.5 rounded-full"
                style={{
                  background:
                    month.status === 'paid' ? 'var(--status-active-bg)' : 'var(--status-trial-bg)',
                  color: month.status === 'paid' ? 'var(--status-active)' : 'var(--status-trial)',
                }}
              >
                {month.status}
              </span>
              <button
                aria-label="Download invoice"
                className="h-7 w-7 rounded-lg flex items-center justify-center transition-colors hover:bg-[var(--dash-hover)]"
                style={{ color: 'var(--dash-text-dim)' }}
              >
                <Download size={13} />
              </button>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

const Billing: React.FC = (): React.JSX.Element => (
  <>
    <HelmetContainer title="Billing · rtylr" />
    <div className="space-y-6 max-w-3xl">
      <Breadcrumbs />
      <PlanSelectionSection />
      <CurrentMonthCard />
      <PaymentMethodCard />
      <BillingHistoryCard />
    </div>
  </>
);

export default Billing;
