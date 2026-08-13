import React, { useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { motion } from 'motion/react';
import { Plus } from 'lucide-react';
import { toast } from 'sonner';
import { getOneKitErrorMessage } from '@voxire/api';
import PageHeader from '../../shared/components/PageHeader';
import MoneyDisplay from '../../shared/components/MoneyDisplay';
import { SkeletonRow } from '../../shared/components/Skeleton';
import { useJournal, useCreateJournalEntry } from '../../shared/hooks/useChartOfAccounts';
import { cardStyle, primaryBtn, ghostBtn, inputStyle } from '../../shared/lib/tokens';

const Journal: React.FC = () => {
  const [page, setPage] = useState(1);
  const { data: journalData, isLoading } = useJournal({ page });
  const data = journalData?.entries ?? [];
  const hasMore = journalData?.hasMore ?? false;
  const createMut = useCreateJournalEntry();
  const [showCreate, setShowCreate] = useState(false);
  const [memo, setMemo] = useState('');
  const [date, setDate] = useState(new Date().toISOString().slice(0, 10));
  const [debitAcc, setDebitAcc] = useState('');
  const [debitAccId, setDebitAccId] = useState(0);
  const [creditAcc, setCreditAcc] = useState('');
  const [creditAccId, setCreditAccId] = useState(0);
  const [amount, setAmount] = useState('');

  const handleCreate = async (): Promise<void> => {
    if (!memo.trim()) {
      toast.error('Memo is required');
      return;
    }
    const parsedAmount = parseFloat(amount);
    if (!amount || isNaN(parsedAmount) || parsedAmount <= 0) {
      toast.error('Amount must be greater than zero');
      return;
    }
    const today = new Date().toISOString().slice(0, 10);
    if (date > today) {
      toast.error('Date cannot be in the future');
      return;
    }
    const amt = parsedAmount;
    try {
      await createMut.mutateAsync({
        date,
        description: memo.trim(),
        lines: [
          {
            accountId: debitAccId,
            accountName: debitAcc || 'Debit Account',
            debit: amt,
            credit: 0,
            description: '',
          },
          {
            accountId: creditAccId,
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
      setDebitAcc('');
      setDebitAccId(0);
      setCreditAcc('');
      setCreditAccId(0);
    } catch (error) {
      toast.error(getOneKitErrorMessage(error, 'Failed to create journal entry'));
    }
  };

  return (
    <motion.div
      initial={{ opacity: 0, y: 4 }}
      animate={{ opacity: 1, y: 0 }}
      style={{ display: 'flex', flexDirection: 'column', gap: 20 }}
    >
      <HelmetContainer title="Journal · rtylr" />
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
            // The API returns authoritative entry totals; do not recompute accounting
            // invariants from client-side floating-point line values.
            const isBalanced = e.debit != null && e.credit != null && e.debit === e.credit;
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
                      color: isBalanced ? 'var(--hub-text-dim)' : 'var(--finance-danger)',
                    }}
                  >
                    {isBalanced ? 'Balanced' : 'Out of balance'}
                  </div>
                </div>
                <table style={{ width: '100%', borderCollapse: 'collapse' }}>
                  <thead>
                    <tr>
                      <th
                        style={{
                          textAlign: 'start',
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
                          textAlign: 'end',
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
                          textAlign: 'end',
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
                            textAlign: 'end',
                            color: l.debit ? 'var(--hub-text)' : 'var(--hub-text-dim)',
                          }}
                        >
                          {l.debit ? <MoneyDisplay amount={l.debit} /> : '-'}
                        </td>
                        <td
                          style={{
                            padding: '8px 4px',
                            fontSize: 13,
                            textAlign: 'end',
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
            background: 'var(--finance-overlay)',
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
              max={new Date().toISOString().split('T')[0]}
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
              type="number"
              min="0"
              step="1"
              placeholder="Debit account ID (optional)"
              value={debitAccId || ''}
              onChange={(e) => setDebitAccId(parseInt(e.target.value, 10) || 0)}
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
              step="1"
              placeholder="Credit account ID (optional)"
              value={creditAccId || ''}
              onChange={(e) => setCreditAccId(parseInt(e.target.value, 10) || 0)}
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

      {(page > 1 || hasMore) && (
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            padding: '8px 0',
          }}
        >
          <span style={{ fontSize: 12, color: 'var(--hub-text-dim)' }}>Page {page}</span>
          <div style={{ display: 'flex', gap: 8 }}>
            <button
              style={{ ...ghostBtn, padding: '4px 12px', opacity: page <= 1 ? 0.3 : 1 }}
              onClick={() => setPage((p) => Math.max(1, p - 1))}
              disabled={page <= 1}
            >
              Prev
            </button>
            <button
              style={{ ...ghostBtn, padding: '4px 12px', opacity: !hasMore ? 0.3 : 1 }}
              onClick={() => setPage((p) => p + 1)}
              disabled={!hasMore}
            >
              Next
            </button>
          </div>
        </div>
      )}
    </motion.div>
  );
};

export default Journal;
