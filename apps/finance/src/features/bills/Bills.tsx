import React, { useState } from 'react';
import { motion } from 'motion/react';
import { Plus, CheckCircle, Clock } from 'lucide-react';
import { toast } from 'sonner';
import PageHeader from '../../shared/components/PageHeader';
import MoneyDisplay from '../../shared/components/MoneyDisplay';
import EmptyState from '../../shared/components/EmptyState';
import { SkeletonRow } from '../../shared/components/Skeleton';
import { useBills } from './useBills';
import { cardStyle, ghostBtn, primaryBtn, ACCENT } from '../../shared/lib/tokens';

const STATUS_STYLE: Record<string, React.CSSProperties> = {
  unpaid: { background: 'rgba(113,113,122,0.15)', color: '#a1a1aa' },
  paid: { background: 'rgba(16,185,129,0.12)', color: '#10b981' },
  overdue: { background: 'rgba(239,68,68,0.12)', color: '#ef4444' },
  scheduled: { background: 'rgba(59,130,246,0.12)', color: '#3b82f6' },
};

const Bills: React.FC = () => {
  const { data = [], isLoading } = useBills();
  const [selected, setSelected] = useState<Set<string>>(new Set());

  const unpaid = data.filter((b: any) => b.status === 'unpaid' || b.status === 'overdue');
  const totalDue = unpaid.reduce((s: any, b: any) => s + b.amount, 0);

  const toggle = (id: string) =>
    setSelected((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });

  const payAll = () => {
    toast.success(`Scheduled ${selected.size} bills for payment`);
    setSelected(new Set());
  };

  return (
    <motion.div
      initial={{ opacity: 0, y: 4 }}
      animate={{ opacity: 1, y: 0 }}
      style={{ display: 'flex', flexDirection: 'column', gap: 20 }}
    >
      <PageHeader
        title="Bills to Pay"
        subtitle={
          <>
            {unpaid.length} outstanding <>&middot;</> <MoneyDisplay amount={totalDue} /> total
          </>
        }
        actions={
          <>
            {selected.size > 0 && (
              <button style={primaryBtn} onClick={payAll}>
                <CheckCircle size={13} /> Pay selected ({selected.size})
              </button>
            )}
            <button style={ghostBtn}>
              <Plus size={13} /> New Bill
            </button>
          </>
        }
      />

      {isLoading ? (
        <div style={{ ...cardStyle, padding: 20 }}>
          <SkeletonRow />
        </div>
      ) : data.length === 0 ? (
        <EmptyState title="No bills" description="Track supplier invoices to plan cash outflows." />
      ) : (
        <div style={{ ...cardStyle, overflow: 'hidden' }}>
          <table style={{ width: '100%', borderCollapse: 'collapse' }}>
            <thead>
              <tr style={{ borderBottom: '1px solid var(--hub-border)' }}>
                {['', 'Bill #', 'Vendor', 'Category', 'Amount', 'Due', 'Status', ''].map((h, i) => (
                  <th
                    key={i}
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
              {data.map((b: any) => (
                <tr key={b.id} style={{ borderBottom: '1px solid var(--hub-border)' }}>
                  <td style={{ padding: '12px 16px', width: 32 }}>
                    <input
                      type="checkbox"
                      checked={selected.has(b.id)}
                      onChange={() => toggle(b.id)}
                      disabled={b.status === 'paid'}
                      style={{ accentColor: ACCENT }}
                    />
                  </td>
                  <td style={{ padding: '12px 16px', fontSize: 13, fontWeight: 600 }}>{b.id}</td>
                  <td style={{ padding: '12px 16px', fontSize: 13 }}>{b.vendor}</td>
                  <td
                    style={{ padding: '12px 16px', fontSize: 12, color: 'var(--hub-text-muted)' }}
                  >
                    {b.category}
                  </td>
                  <td style={{ padding: '12px 16px', fontSize: 13, fontWeight: 600 }}>
                    <MoneyDisplay amount={b.amount} />
                  </td>
                  <td
                    style={{
                      padding: '12px 16px',
                      fontSize: 12,
                      color: b.status === 'overdue' ? '#ef4444' : 'var(--hub-text-muted)',
                    }}
                  >
                    {b.dueDate}
                  </td>
                  <td style={{ padding: '12px 16px' }}>
                    <span
                      style={{
                        fontSize: 11,
                        fontWeight: 600,
                        padding: '3px 10px',
                        borderRadius: 20,
                        textTransform: 'capitalize',
                        ...STATUS_STYLE[b.status],
                      }}
                    >
                      {b.status}
                    </span>
                  </td>
                  <td style={{ padding: '12px 16px' }}>
                    {b.status === 'unpaid' || b.status === 'overdue' ? (
                      <button
                        style={{ ...ghostBtn, padding: '4px 10px' }}
                        onClick={() => toast.success(`Scheduled ${b.id}`)}
                      >
                        <Clock size={12} /> Schedule
                      </button>
                    ) : (
                      <span style={{ fontSize: 11, color: 'var(--hub-text-dim)' }}>-</span>
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

export default Bills;
