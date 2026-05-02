import React from 'react';
import { motion } from 'motion/react';
import { Play, Pause, RotateCcw, Repeat } from 'lucide-react';
import { toast } from 'sonner';
import PageHeader from '../../shared/components/PageHeader';
import MoneyDisplay from '../../shared/components/MoneyDisplay';
import EmptyState from '../../shared/components/EmptyState';
import { SkeletonRow } from '../../shared/components/Skeleton';
import { useRecurringInvoices, useSetRecurringStatus } from './useRecurringInvoices';
import { cardStyle, ghostBtn, primaryBtn, ACCENT } from '../../shared/lib/tokens';

const cadenceLabel: Record<string, string> = {
  weekly: 'Weekly',
  monthly: 'Monthly',
  quarterly: 'Quarterly',
  yearly: 'Yearly',
};

const RecurringInvoices: React.FC = () => {
  const { data = [], isLoading } = useRecurringInvoices();
  const setStatus = useSetRecurringStatus();

  const togglePause = (id: string, current: string | null | undefined, client: string): void => {
    const next = current === 'active' ? 'paused' : 'active';
    setStatus.mutate(
      { id, status: next },
      {
        onSuccess: () =>
          toast.success(next === 'paused' ? `Paused ${client}` : `Resumed ${client}`),
        onError: () => toast.error(`Failed to update ${client}`),
      }
    );
  };

  return (
    <motion.div
      initial={{ opacity: 0, y: 4 }}
      animate={{ opacity: 1, y: 0 }}
      style={{ display: 'flex', flexDirection: 'column', gap: 20 }}
    >
      <PageHeader
        title="Recurring Invoices"
        subtitle={`${data.length} active subscriptions`}
        breadcrumbs={[{ label: 'Invoices', to: '/invoices' }, { label: 'Recurring' }]}
        actions={
          <button
            style={primaryBtn}
            aria-label="New Recurring Invoice"
            onClick={() => toast.info('Create recurring invoice coming soon')}
          >
            <Repeat size={13} /> New Recurring
          </button>
        }
      />

      {isLoading ? (
        <div style={{ ...cardStyle, padding: 20 }}>
          <SkeletonRow />
        </div>
      ) : data.length === 0 ? (
        <EmptyState
          title="No recurring invoices yet"
          description="Create a subscription to automatically bill customers on a schedule."
          action={
            <button
              style={primaryBtn}
              aria-label="Create subscription"
              onClick={() => toast.info('Create recurring invoice coming soon')}
            >
              Create subscription
            </button>
          }
        />
      ) : (
        <div style={{ ...cardStyle, overflow: 'hidden' }}>
          <table style={{ width: '100%', borderCollapse: 'collapse' }}>
            <thead>
              <tr style={{ borderBottom: '1px solid var(--hub-border)' }}>
                {['Client', 'Cadence', 'Amount', 'Next run', 'Status', 'Actions'].map((h) => (
                  <th
                    key={h}
                    style={{
                      padding: '10px 16px',
                      textAlign: 'left',
                      fontSize: 11,
                      fontWeight: 600,
                      textTransform: 'uppercase',
                      letterSpacing: '0.07em',
                      color: 'var(--hub-text-dim)',
                    }}
                  >
                    {h}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody>
              {data.map((r) => (
                <tr key={r.id} style={{ borderBottom: '1px solid var(--hub-border)' }}>
                  <td
                    style={{
                      padding: '12px 16px',
                      fontSize: 13,
                      fontWeight: 600,
                      color: 'var(--hub-text)',
                    }}
                  >
                    {r.client}
                  </td>
                  <td
                    style={{ padding: '12px 16px', fontSize: 12, color: 'var(--hub-text-muted)' }}
                  >
                    {cadenceLabel[r.cadence ?? '']}
                  </td>
                  <td
                    style={{
                      padding: '12px 16px',
                      fontSize: 13,
                      fontWeight: 600,
                      color: 'var(--hub-text)',
                    }}
                  >
                    <MoneyDisplay amount={r.amount} />
                  </td>
                  <td
                    style={{ padding: '12px 16px', fontSize: 12, color: 'var(--hub-text-muted)' }}
                  >
                    {r.nextDate}
                  </td>
                  <td style={{ padding: '12px 16px' }}>
                    <span
                      style={{
                        fontSize: 11,
                        fontWeight: 600,
                        padding: '3px 10px',
                        borderRadius: 20,
                        background:
                          r.status === 'active' ? 'rgba(16,185,129,0.12)' : 'rgba(245,158,11,0.12)',
                        color: r.status === 'active' ? ACCENT : '#f59e0b',
                      }}
                    >
                      {r.status === 'active' ? 'Active' : 'Paused'}
                    </span>
                  </td>
                  <td style={{ padding: '12px 16px' }}>
                    <div style={{ display: 'flex', gap: 6 }}>
                      <button
                        style={{
                          ...ghostBtn,
                          padding: '4px 10px',
                          opacity: setStatus.isPending ? 0.6 : 1,
                        }}
                        onClick={() => togglePause(r.id, r.status, r.client)}
                        disabled={setStatus.isPending}
                        aria-label={
                          r.status === 'active'
                            ? `Pause recurring invoice ${r.client}`
                            : `Resume recurring invoice ${r.client}`
                        }
                      >
                        {r.status === 'active' ? <Pause size={12} /> : <Play size={12} />}
                        {r.status === 'active' ? 'Pause' : 'Resume'}
                      </button>
                      <button
                        style={{ ...ghostBtn, padding: '4px 10px' }}
                        aria-label={`Run invoice now for ${r.client}`}
                        onClick={() => toast.info('Run now coming soon')}
                      >
                        <RotateCcw size={12} /> Run now
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </motion.div>
  );
};

export default RecurringInvoices;
