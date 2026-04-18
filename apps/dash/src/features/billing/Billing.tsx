import React from 'react';
import { Download, CreditCard, Clock, Zap } from 'lucide-react';
import {
  useBillingDetail,
  useBillingHistory,
  useMySubscription,
  centsToDisplay,
  formatPeriodRange,
  trialDaysLeft,
  type BillingLineItem,
} from '../../shared/hooks/queries/useSubscription';
import { Breadcrumbs } from '../../shared/components/dash';

// ── Helpers ───────────────────────────────────────────────────────────────────

function formatMonth(iso: string): string {
  const d = new Date(iso);
  return d.toLocaleDateString('en-US', { month: 'long', year: 'numeric', timeZone: 'UTC' });
}

// ── Line item row ─────────────────────────────────────────────────────────────

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
        {!item.periodEnd && (
          <span
            className="ml-2 text-[10px] font-sans font-semibold px-1.5 py-0.5 rounded-full"
            style={{ background: 'var(--status-trial-bg)', color: 'var(--status-trial)' }}
          >
            in progress
          </span>
        )}
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

// ── Current month section ─────────────────────────────────────────────────────

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
      {/* Header */}
      <div
        className="px-5 py-4 flex items-center justify-between"
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
        ) : (
          billing?.nextBillingDate && (
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
          )
        )}
      </div>

      {/* Line items */}
      <div className="px-5">
        {!billing || billing.lineItems.length === 0 ? (
          <div className="py-8 text-center">
            <p className="text-sm" style={{ color: 'var(--dash-text-muted)' }}>
              No usage recorded yet this month.
            </p>
          </div>
        ) : (
          <>
            {/* Column headers */}
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

      {/* Total */}
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

// ── Payment method section ────────────────────────────────────────────────────

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

// ── Billing history ───────────────────────────────────────────────────────────

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

// ── Main ──────────────────────────────────────────────────────────────────────

const Billing: React.FC = (): React.JSX.Element => (
  <div className="space-y-6 max-w-2xl">
    <Breadcrumbs />
    <CurrentMonthCard />
    <PaymentMethodCard />
    <BillingHistoryCard />
  </div>
);

export default Billing;
