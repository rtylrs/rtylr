import React from 'react';
import { CreditCard, Download, Check } from 'lucide-react';
import { Button } from '../../shared/components/ui';
import { Breadcrumbs, SettingSection, SkeletonList } from '../../shared/components/dash';
import { useBilling, useInvoices, type Invoice } from '../../shared/hooks/queries/useBilling';
import { useCheckout } from '../../shared/hooks/queries/useSubscription';
import { toast } from 'sonner';

function formatMoney(amount: number, currency: string) {
  const safeAmount = Number.isFinite(amount) ? amount : 0;
  try {
    return new Intl.NumberFormat('en-US', { style: 'currency', currency }).format(safeAmount);
  } catch {
    return `$${safeAmount.toFixed(2)}`;
  }
}

function formatDate(iso: string) {
  return new Date(iso).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  });
}

const UsageBar: React.FC<{ label: string; used: number; limit: number; unit?: string }> = ({
  label,
  used,
  limit,
  unit,
}) => {
  const safeUsed = Number.isFinite(used) ? used : 0;
  const safeLimit = Number.isFinite(limit) && limit > 0 ? limit : 0;
  const pct = safeLimit > 0 ? Math.min(100, Math.round((safeUsed / safeLimit) * 100)) : 0;
  const color = pct > 90 ? '#ef4444' : pct > 75 ? '#f59e0b' : '#10b981';
  return (
    <div className="space-y-1.5">
      <div className="flex items-center justify-between text-xs">
        <span style={{ color: 'var(--dash-text-muted)' }}>{label}</span>
        <span style={{ color: 'var(--dash-text)' }}>
          {safeUsed.toLocaleString()}
          {unit ?? ''} / {safeLimit.toLocaleString()}
          {unit ?? ''}
        </span>
      </div>
      <div
        className="h-1.5 rounded-full overflow-hidden"
        style={{ background: 'var(--dash-hover)' }}
      >
        <div className="h-full transition-all" style={{ width: `${pct}%`, background: color }} />
      </div>
    </div>
  );
};

const InvoiceRow: React.FC<{ invoice: Invoice }> = ({ invoice }) => (
  <div className="flex items-center gap-4 px-4 py-3 rounded-lg transition-colors hover:bg-[var(--dash-hover)]">
    <div className="flex-1 min-w-0">
      <p className="text-sm font-semibold" style={{ color: 'var(--dash-text)' }}>
        {invoice.number}
      </p>
      <p className="text-[11px]" style={{ color: 'var(--dash-text-dim)' }}>
        {formatDate(invoice.createdAt)}
      </p>
    </div>
    <span
      className="px-1.5 py-0.5 rounded text-[10px] font-semibold uppercase"
      style={{
        background: invoice.status === 'paid' ? 'rgba(16,185,129,0.15)' : 'rgba(245,158,11,0.15)',
        color: invoice.status === 'paid' ? '#10b981' : '#f59e0b',
      }}
    >
      {invoice.status}
    </span>
    <p className="text-sm font-semibold w-20 text-right" style={{ color: 'var(--dash-text)' }}>
      {formatMoney(invoice.amount, invoice.currency)}
    </p>
    {invoice.pdfUrl ? (
      <a
        href={invoice.pdfUrl}
        target="_blank"
        rel="noopener noreferrer"
        className="h-7 w-7 rounded-lg flex items-center justify-center transition-colors hover:bg-[var(--dash-active)]"
        style={{ color: 'var(--dash-text-dim)' }}
        aria-label="Download invoice"
      >
        <Download size={13} />
      </a>
    ) : (
      <button
        className="h-7 w-7 rounded-lg flex items-center justify-center opacity-30 cursor-not-allowed"
        style={{ color: 'var(--dash-text-dim)' }}
        disabled
        aria-label="Invoice not available"
      >
        <Download size={13} />
      </button>
    )}
  </div>
);

const PLANS = [
  { key: 'starter', label: 'Starter', price: 29, seats: 5, features: ['Hub', 'POS', '1 location'] },
  {
    key: 'pro',
    label: 'Pro',
    price: 79,
    seats: 15,
    features: ['Everything in Starter', 'All apps', '5 locations', 'Priority support'],
  },
  {
    key: 'enterprise',
    label: 'Enterprise',
    price: 0,
    seats: 0,
    features: ['Unlimited seats', 'SSO & SAML', 'Custom SLA', 'Dedicated manager'],
  },
];

const BillingHome: React.FC = () => {
  const { data: billing, isLoading } = useBilling();
  const { data: invoices = [], isLoading: invoicesLoading } = useInvoices();
  const checkout = useCheckout();

  if (isLoading || !billing) {
    return (
      <div className="space-y-5 max-w-4xl">
        <SkeletonList rows={6} withAvatar={false} />
      </div>
    );
  }

  return (
    <div className="space-y-5 max-w-4xl">
      <div className="flex items-center gap-3">
        <div
          className="h-9 w-9 rounded-xl flex items-center justify-center"
          style={{ background: 'rgba(69,131,196,0.15)' }}
        >
          <CreditCard size={16} style={{ color: '#4583C4' }} />
        </div>
        <div>
          <h1 className="text-lg font-semibold" style={{ color: 'var(--dash-text)' }}>
            Billing
          </h1>
          <p className="text-xs mt-0.5" style={{ color: 'var(--dash-text-dim)' }}>
            Subscription, usage, and invoices.
          </p>
        </div>
      </div>

      <SettingSection title="Current plan" description={`Renews ${formatDate(billing.renewsAt)}`}>
        <div className="flex items-center gap-4">
          <div
            className="flex-1 rounded-lg p-4"
            style={{ background: 'var(--dash-hover)', border: '1px solid var(--dash-border)' }}
          >
            <p
              className="text-[11px] font-semibold uppercase tracking-wider"
              style={{ color: 'var(--dash-text-dim)' }}
            >
              Plan
            </p>
            <p className="text-2xl font-bold capitalize mt-1" style={{ color: 'var(--dash-text)' }}>
              {billing.plan}
            </p>
          </div>
          <div
            className="flex-1 rounded-lg p-4"
            style={{ background: 'var(--dash-hover)', border: '1px solid var(--dash-border)' }}
          >
            <p
              className="text-[11px] font-semibold uppercase tracking-wider"
              style={{ color: 'var(--dash-text-dim)' }}
            >
              Monthly
            </p>
            <p className="text-2xl font-bold mt-1" style={{ color: 'var(--dash-text)' }}>
              {formatMoney(billing.amount, billing.currency)}
            </p>
          </div>
          <div
            className="flex-1 rounded-lg p-4"
            style={{ background: 'var(--dash-hover)', border: '1px solid var(--dash-border)' }}
          >
            <p
              className="text-[11px] font-semibold uppercase tracking-wider"
              style={{ color: 'var(--dash-text-dim)' }}
            >
              Status
            </p>
            <p
              className="text-2xl font-bold mt-1 capitalize"
              style={{ color: billing.status === 'active' ? '#10b981' : '#f59e0b' }}
            >
              {billing.status}
            </p>
          </div>
        </div>
      </SettingSection>

      <SettingSection title="Usage this month" description="Resets on the 1st of each month.">
        <div className="space-y-4">
          <UsageBar
            label="Seats"
            used={billing.usage.seats.used}
            limit={billing.usage.seats.limit}
          />
          <UsageBar
            label="Orders"
            used={billing.usage.orders.used}
            limit={billing.usage.orders.limit}
          />
          <UsageBar
            label="API calls"
            used={billing.usage.apiCalls.used}
            limit={billing.usage.apiCalls.limit}
          />
          <UsageBar
            label="Storage"
            used={billing.usage.storageGb.used}
            limit={billing.usage.storageGb.limit}
            unit=" GB"
          />
        </div>
      </SettingSection>

      <SettingSection
        title="Upgrade plan"
        description="Change your plan at any time. Prorated instantly."
      >
        <div className="grid grid-cols-3 gap-3">
          {PLANS.map((plan) => {
            const isCurrent = plan.key === billing.plan;
            return (
              <div
                key={plan.key}
                className="rounded-lg p-4 flex flex-col"
                style={{
                  background: isCurrent ? 'rgba(69,131,196,0.06)' : 'var(--dash-hover)',
                  border: `1px solid ${isCurrent ? 'rgba(69,131,196,0.3)' : 'var(--dash-border)'}`,
                }}
              >
                <div className="flex items-center justify-between mb-2">
                  <p className="text-sm font-semibold" style={{ color: 'var(--dash-text)' }}>
                    {plan.label}
                  </p>
                  {isCurrent && (
                    <span
                      className="text-[10px] font-semibold uppercase"
                      style={{ color: '#4583C4' }}
                    >
                      Current
                    </span>
                  )}
                </div>
                <p className="text-xl font-bold" style={{ color: 'var(--dash-text)' }}>
                  {plan.price === 0 ? 'Custom' : `$${plan.price}`}
                  <span className="text-xs font-normal" style={{ color: 'var(--dash-text-dim)' }}>
                    {plan.price === 0 ? '' : '/mo'}
                  </span>
                </p>
                <ul className="space-y-1 mt-3 flex-1">
                  {plan.features.map((f) => (
                    <li
                      key={f}
                      className="flex items-start gap-1.5 text-[11px]"
                      style={{ color: 'var(--dash-text-muted)' }}
                    >
                      <Check size={10} className="mt-0.5 shrink-0" style={{ color: '#10b981' }} />
                      {f}
                    </li>
                  ))}
                </ul>
                <Button
                  size="sm"
                  variant={isCurrent ? 'secondary' : 'default'}
                  className="mt-3 w-full"
                  disabled={isCurrent || checkout.isPending}
                  onClick={() => {
                    if (plan.price === 0) {
                      window.open('mailto:info@voxire.com?subject=Enterprise inquiry', '_blank');
                      return;
                    }
                    checkout.mutate(
                      { planSlug: plan.key },
                      {
                        onSuccess: (data) => {
                          window.location.href = data.checkoutUrl;
                        },
                        onError: () => toast.error('Failed to start checkout'),
                      }
                    );
                  }}
                >
                  {isCurrent
                    ? 'Current plan'
                    : plan.price === 0
                      ? 'Contact sales'
                      : checkout.isPending
                        ? 'Loading...'
                        : 'Upgrade'}
                </Button>
              </div>
            );
          })}
        </div>
      </SettingSection>

      <SettingSection title="Invoices">
        {invoicesLoading ? (
          <SkeletonList rows={3} withAvatar={false} />
        ) : (
          <div className="-mx-5">
            {invoices.map((i) => (
              <InvoiceRow key={i.id} invoice={i} />
            ))}
          </div>
        )}
      </SettingSection>
    </div>
  );
};

export default BillingHome;
