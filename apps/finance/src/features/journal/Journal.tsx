import React, { useState } from 'react';
import { motion } from 'motion/react';
import { Plus } from 'lucide-react';
import { toast } from 'sonner';
import PageHeader from '../../shared/components/PageHeader';
import MoneyDisplay from '../../shared/components/MoneyDisplay';
import { SkeletonRow } from '../../shared/components/Skeleton';
import { useJournal, useCreateJournalEntry } from '../../shared/hooks/useChartOfAccounts';
import { cardStyle, primaryBtn, ghostBtn, inputStyle } from '../../shared/lib/tokens';

const Journal: React.FC = () => {
  const { data = [], isLoading } = useJournal();
  const createMut = useCreateJournalEntry();
  const [showCreate, setShowCreate] = useState(false);
  const [memo, setMemo] = useState('');
  const [date, setDate] = useState(new Date().toISOString().slice(0, 10));
  const [debitAcc, setDebitAcc] = useState('');
  const [creditAcc, setCreditAcc] = useState('');
  const [amount, setAmount] = useState('');

  const handleCreate = async (): Promise<void> => {
    if (!memo.trim() || !amount) return;
    const amt = parseFloat(amount);
    try {
      await createMut.mutateAsync({
        date,
        description: memo.trim(),
        lines: [
          {
            accountId: 0,
            accountName: debitAcc || 'Debit Account',
            debit: amt,
            credit: 0,
            description: '',
          },
          {
            accountId: 0,
            accountName: creditAcc || 'Credit Account',
            debit: 0,
            credit: amt,
            description: '',
          },
        ],
      });
      toast.success('Journal entry created');
      setShowCreate(false);
      setMemo('');
      setAmount('');
    } catch {
      toast.error('Failed to create journal entry');
    }
  };

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
          <button style={primaryBtn} onClick={() => setShowCreate(true)}>
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
          {data.map((e) => {
            const debits = (e.lines ?? []).reduce((s: number, l) => s + (l.debit ?? 0), 0);
            const credits = (e.lines ?? []).reduce((s: number, l) => s + (l.credit ?? 0), 0);
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
                    {(e.lines ?? []).map((l, i) => (
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
      {showCreate && (
        <div
          style={{
            position: 'fixed',
            inset: 0,
            background: 'rgba(0,0,0,0.5)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            zIndex: 9999,
          }}
          onClick={(e) => {
            if (e.target === e.currentTarget) setShowCreate(false);
          }}
        >
          <div
            style={{
              ...cardStyle,
              padding: 24,
              width: 400,
              display: 'flex',
              flexDirection: 'column',
              gap: 12,
            }}
          >
            <h3 style={{ fontSize: 15, fontWeight: 700, color: 'var(--hub-text)', margin: 0 }}>
              New Journal Entry
            </h3>
            <input
              style={inputStyle}
              type="date"
              value={date}
              onChange={(e) => setDate(e.target.value)}
            />
            <input
              style={inputStyle}
              placeholder="Memo / description"
              value={memo}
              onChange={(e) => setMemo(e.target.value)}
            />
            <input
              style={inputStyle}
              placeholder="Debit account name"
              value={debitAcc}
              onChange={(e) => setDebitAcc(e.target.value)}
            />
            <input
              style={inputStyle}
              placeholder="Credit account name"
              value={creditAcc}
              onChange={(e) => setCreditAcc(e.target.value)}
            />
            <input
              style={inputStyle}
              type="number"
              min="0"
              step="0.01"
              placeholder="Amount"
              value={amount}
              onChange={(e) => setAmount(e.target.value)}
            />
            <div style={{ display: 'flex', gap: 8, justifyContent: 'flex-end' }}>
              <button style={ghostBtn} onClick={() => setShowCreate(false)}>
                Cancel
              </button>
              <button style={primaryBtn} onClick={handleCreate} disabled={createMut.isPending}>
                {createMut.isPending ? 'Creating…' : 'Create'}
              </button>
            </div>
          </div>
        </div>
      )}
    </motion.div>
  );
};

export default Journal;
