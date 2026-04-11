import React from 'react';
import { motion } from 'motion/react';
import { Plus } from 'lucide-react';
import { toast } from 'sonner';
import PageHeader from '../../shared/components/PageHeader';
import MoneyDisplay from '../../shared/components/MoneyDisplay';
import { SkeletonRow } from '../../shared/components/Skeleton';
import { useJournal } from '../../shared/hooks/useChartOfAccounts';
import { cardStyle, primaryBtn } from '../../shared/lib/tokens';

const Journal: React.FC = () => {
  const { data = [], isLoading } = useJournal();

  return (
    <motion.div
      initial={{ opacity: 0, y: 4 }}
      animate={{ opacity: 1, y: 0 }}
      style={{ display: 'flex', flexDirection: 'column', gap: 20 }}
    >
      <PageHeader
        title="Journal"
        subtitle={`${data.length} entries`}
        actions={
          <button style={primaryBtn} onClick={() => toast.success('Journal entry created')}>
            <Plus size={13} /> New entry
          </button>
        }
      />

      {isLoading ? (
        <div style={{ ...cardStyle, padding: 20 }}>
          <SkeletonRow />
        </div>
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
          {data.map((e: any) => {
            const debits = e.lines.reduce((s: any, l: any) => s + l.debit, 0);
            const credits = e.lines.reduce((s: any, l: any) => s + l.credit, 0);
            return (
              <div key={e.id} style={{ ...cardStyle, padding: 18 }}>
                <div
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'space-between',
                    marginBottom: 10,
                  }}
                >
                  <div>
                    <div style={{ fontSize: 13, fontWeight: 600, color: 'var(--hub-text)' }}>
                      {e.memo}
                    </div>
                    <div style={{ fontSize: 11, color: 'var(--hub-text-dim)' }}>
                      {e.id} <>&middot;</> {e.date}
                    </div>
                  </div>
                  <div
                    style={{
                      fontSize: 11,
                      color: debits === credits ? 'var(--hub-text-dim)' : '#ef4444',
                    }}
                  >
                    {debits === credits ? 'Balanced' : 'Out of balance'}
                  </div>
                </div>
                <table style={{ width: '100%', borderCollapse: 'collapse' }}>
                  <thead>
                    <tr>
                      <th
                        style={{
                          textAlign: 'left',
                          fontSize: 10,
                          textTransform: 'uppercase',
                          color: 'var(--hub-text-dim)',
                          padding: '6px 4px',
                        }}
                      >
                        Account
                      </th>
                      <th
                        style={{
                          textAlign: 'right',
                          fontSize: 10,
                          textTransform: 'uppercase',
                          color: 'var(--hub-text-dim)',
                          padding: '6px 4px',
                        }}
                      >
                        Debit
                      </th>
                      <th
                        style={{
                          textAlign: 'right',
                          fontSize: 10,
                          textTransform: 'uppercase',
                          color: 'var(--hub-text-dim)',
                          padding: '6px 4px',
                        }}
                      >
                        Credit
                      </th>
                    </tr>
                  </thead>
                  <tbody>
                    {e.lines.map((l: any, i: any) => (
                      <tr key={i} style={{ borderTop: '1px solid var(--hub-border)' }}>
                        <td style={{ padding: '8px 4px', fontSize: 13, color: 'var(--hub-text)' }}>
                          {l.account}
                        </td>
                        <td
                          style={{
                            padding: '8px 4px',
                            fontSize: 13,
                            textAlign: 'right',
                            color: l.debit ? 'var(--hub-text)' : 'var(--hub-text-dim)',
                          }}
                        >
                          {l.debit ? <MoneyDisplay amount={l.debit} /> : '-'}
                        </td>
                        <td
                          style={{
                            padding: '8px 4px',
                            fontSize: 13,
                            textAlign: 'right',
                            color: l.credit ? 'var(--hub-text)' : 'var(--hub-text-dim)',
                          }}
                        >
                          {l.credit ? <MoneyDisplay amount={l.credit} /> : '-'}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            );
          })}
        </div>
      )}
    </motion.div>
  );
};

export default Journal;
