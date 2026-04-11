import React from 'react';
import { motion } from 'motion/react';
import { Plus, ArrowRight, CheckCircle } from 'lucide-react';
import { toast } from 'sonner';
import PageHeader from '../../shared/components/PageHeader';
import MoneyDisplay from '../../shared/components/MoneyDisplay';
import EmptyState from '../../shared/components/EmptyState';
import { SkeletonRow } from '../../shared/components/Skeleton';
import { useQuotes } from './useQuotes';
import { cardStyle, ghostBtn, primaryBtn } from '../../shared/lib/tokens';

const STATUS_STYLE: Record<string, React.CSSProperties> = {
  Draft: { background: 'rgba(113,113,122,0.15)', color: '#a1a1aa' },
  Sent: { background: 'rgba(59,130,246,0.12)', color: '#3b82f6' },
  Accepted: { background: 'rgba(16,185,129,0.12)', color: '#10b981' },
  Declined: { background: 'rgba(239,68,68,0.12)', color: '#ef4444' },
  Expired: { background: 'rgba(245,158,11,0.12)', color: '#f59e0b' },
};

const Quotes: React.FC = () => {
  const { data = [], isLoading } = useQuotes();
  const total = data.reduce((s: any, q: any) => s + q.amount, 0);

  return (
    <motion.div
      initial={{ opacity: 0, y: 4 }}
      animate={{ opacity: 1, y: 0 }}
      style={{ display: 'flex', flexDirection: 'column', gap: 20 }}
    >
      <PageHeader
        title="Quotes & Estimates"
        subtitle={
          <>
            {data.length} quotes <>&middot;</> <MoneyDisplay amount={total} /> total
          </>
        }
        actions={
          <button style={primaryBtn}>
            <Plus size={13} /> New Quote
          </button>
        }
      />

      {isLoading ? (
        <div style={{ ...cardStyle, padding: 20 }}>
          <SkeletonRow />
        </div>
      ) : data.length === 0 ? (
        <EmptyState
          title="No quotes yet"
          description="Send estimates to clients and convert them to invoices when accepted."
        />
      ) : (
        <div style={{ ...cardStyle, overflow: 'hidden' }}>
          <table style={{ width: '100%', borderCollapse: 'collapse' }}>
            <thead>
              <tr style={{ borderBottom: '1px solid var(--hub-border)' }}>
                {['Quote #', 'Client', 'Amount', 'Status', 'Issued', 'Expires', 'Actions'].map(
                  (h) => (
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
                  )
                )}
              </tr>
            </thead>
            <tbody>
              {data.map((q: any) => (
                <tr key={q.id} style={{ borderBottom: '1px solid var(--hub-border)' }}>
                  <td
                    style={{
                      padding: '12px 16px',
                      fontSize: 13,
                      fontWeight: 600,
                      color: 'var(--hub-text)',
                    }}
                  >
                    {q.id}
                  </td>
                  <td style={{ padding: '12px 16px', fontSize: 13, color: 'var(--hub-text)' }}>
                    {q.client}
                  </td>
                  <td style={{ padding: '12px 16px', fontSize: 13, fontWeight: 600 }}>
                    <MoneyDisplay amount={q.amount} />
                  </td>
                  <td style={{ padding: '12px 16px' }}>
                    <span
                      style={{
                        fontSize: 11,
                        fontWeight: 600,
                        padding: '3px 10px',
                        borderRadius: 20,
                        ...STATUS_STYLE[q.status],
                      }}
                    >
                      {q.status}
                    </span>
                  </td>
                  <td
                    style={{ padding: '12px 16px', fontSize: 12, color: 'var(--hub-text-muted)' }}
                  >
                    {q.issuedDate}
                  </td>
                  <td
                    style={{ padding: '12px 16px', fontSize: 12, color: 'var(--hub-text-muted)' }}
                  >
                    {q.expiresAt}
                  </td>
                  <td style={{ padding: '12px 16px' }}>
                    {q.status === 'Accepted' ? (
                      <button
                        style={{
                          ...ghostBtn,
                          padding: '4px 10px',
                          color: '#10b981',
                          borderColor: 'rgba(16,185,129,0.3)',
                        }}
                        onClick={() => toast.success(`Converted ${q.id} to invoice`)}
                      >
                        <ArrowRight size={12} /> Convert
                      </button>
                    ) : q.status === 'Sent' ? (
                      <button
                        style={{ ...ghostBtn, padding: '4px 10px' }}
                        onClick={() => toast.success(`Marked ${q.id} as accepted`)}
                      >
                        <CheckCircle size={12} /> Mark Accepted
                      </button>
                    ) : (
                      <button style={{ ...ghostBtn, padding: '4px 10px' }}>View</button>
                    )}
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

export default Quotes;
