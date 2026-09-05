import React, { useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { motion } from 'motion/react';
import { Plus } from 'lucide-react';
import { toast } from 'sonner';
import { getOneKitErrorMessage } from '@voxire/api';
import { useTranslation } from 'react-i18next';
import PageHeader from '../../shared/components/PageHeader';
import MoneyDisplay from '../../shared/components/MoneyDisplay';
import { SkeletonRow } from '../../shared/components/Skeleton';
import { useJournal, useCreateJournalEntry } from '../../shared/hooks/useChartOfAccounts';
import { cardStyle, primaryBtn, ghostBtn, inputStyle } from '../../shared/lib/tokens';
import { todayIso } from '../../shared/lib/date';

interface LineInput {
  id: string;
  accountName: string;
  accountId: number;
  type: 'debit' | 'credit';
  amount: number;
}

const Journal: React.FC = () => {
  const { t } = useTranslation('journal');
  const [page, setPage] = useState(1);
  const { data: journalData, isLoading } = useJournal({ page });
  const data = journalData?.entries ?? [];
  const hasMore = journalData?.hasMore ?? false;
  const createMut = useCreateJournalEntry();
  const [showCreate, setShowCreate] = useState(false);
  const [memo, setMemo] = useState('');
  const [date, setDate] = useState(todayIso());

  const [lines, setLines] = useState<LineInput[]>([
    { id: '1', accountName: '', accountId: 0, type: 'debit', amount: 0 },
    { id: '2', accountName: '', accountId: 0, type: 'credit', amount: 0 },
  ]);

  const addLine = (type: 'debit' | 'credit'): void => {
    setLines((prev) => [
      ...prev,
      { id: String(Date.now() + Math.random()), accountName: '', accountId: 0, type, amount: 0 },
    ]);
  };

  const removeLine = (id: string): void => {
    if (lines.length <= 2) {
      toast.error('Journal entry requires at least two lines');
      return;
    }
    setLines((prev) => prev.filter((l) => l.id !== id));
  };

  const updateLine = (id: string, field: keyof LineInput, value: unknown): void => {
    setLines((prev) => prev.map((l) => (l.id === id ? { ...l, [field]: value } : l)));
  };

  const totalDebits = lines
    .filter((l) => l.type === 'debit')
    .reduce((sum, l) => sum + (l.amount || 0), 0);
  const totalCredits = lines
    .filter((l) => l.type === 'credit')
    .reduce((sum, l) => sum + (l.amount || 0), 0);

  const handleCreate = async (): Promise<void> => {
    if (!memo.trim()) {
      toast.error(t('memoRequired'));
      return;
    }
    if (Math.abs(totalDebits - totalCredits) > 0.01 || totalDebits === 0) {
      toast.error(
        `Journal entry must balance: Debits (${totalDebits}) must equal Credits (${totalCredits})`
      );
      return;
    }
    const today = todayIso();
    if (date > today) {
      toast.error(t('futureDate'));
      return;
    }
    try {
      await createMut.mutateAsync({
        date,
        description: memo.trim(),
        lines: lines.map((l) => ({
          accountId: l.accountId,
          accountName: l.accountName || (l.type === 'debit' ? 'Debit Account' : 'Credit Account'),
          debit: l.type === 'debit' ? l.amount : 0,
          credit: l.type === 'credit' ? l.amount : 0,
          description: '',
        })),
      });
      toast.success(t('created'));
      setShowCreate(false);
      setMemo('');
      setLines([
        { id: '1', accountName: '', accountId: 0, type: 'debit', amount: 0 },
        { id: '2', accountName: '', accountId: 0, type: 'credit', amount: 0 },
      ]);
    } catch (error) {
      toast.error(getOneKitErrorMessage(error, t('createFailed')));
    }
  };

  return (
    <motion.div
      initial={{ opacity: 0, y: 4 }}
      animate={{ opacity: 1, y: 0 }}
      style={{ display: 'flex', flexDirection: 'column', gap: 20 }}
    >
      <HelmetContainer title={`${t('title')} · rtylr`} />
      <PageHeader
        title={t('title')}
        subtitle={`${data.length} entries`}
        actions={
          <button type="button" style={primaryBtn} onClick={() => setShowCreate(true)}>
            <Plus size={13} /> {t('newEntry')}
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
                    {isBalanced ? t('balanced') : t('outOfBalance')}
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
                        {t('account')}
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
                        {t('debit')}
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
                        {t('credit')}
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
              width: 520,
              maxHeight: '90vh',
              overflowY: 'auto',
              display: 'flex',
              flexDirection: 'column',
              gap: 12,
            }}
          >
            <h3 style={{ fontSize: 15, fontWeight: 700, color: 'var(--hub-text)', margin: 0 }}>
              {t('newJournalEntry')}
            </h3>
            <input
              style={inputStyle}
              type="date"
              value={date}
              max={todayIso()}
              onChange={(e) => setDate(e.target.value)}
            />
            <input
              style={inputStyle}
              placeholder={t('memoPlaceholder')}
              value={memo}
              onChange={(e) => setMemo(e.target.value)}
            />
            <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
              <div
                style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}
              >
                <span style={{ fontSize: 12, fontWeight: 600, color: 'var(--hub-text)' }}>
                  Entry Lines
                </span>
                <div style={{ display: 'flex', gap: 6 }}>
                  <button
                    type="button"
                    style={{ ...ghostBtn, fontSize: 11, padding: '2px 8px' }}
                    onClick={() => addLine('debit')}
                  >
                    + Add Debit
                  </button>
                  <button
                    type="button"
                    style={{ ...ghostBtn, fontSize: 11, padding: '2px 8px' }}
                    onClick={() => addLine('credit')}
                  >
                    + Add Credit
                  </button>
                </div>
              </div>

              {lines.map((l) => (
                <div key={l.id} style={{ display: 'flex', gap: 6, alignItems: 'center' }}>
                  <span
                    style={{
                      fontSize: 10,
                      fontWeight: 700,
                      textTransform: 'uppercase',
                      padding: '2px 6px',
                      borderRadius: 4,
                      background:
                        l.type === 'debit'
                          ? 'var(--finance-info-soft)'
                          : 'var(--finance-accent-soft)',
                      color: l.type === 'debit' ? 'var(--finance-info)' : 'var(--finance-accent)',
                    }}
                  >
                    {l.type}
                  </span>
                  <input
                    style={{ ...inputStyle, flex: 2 }}
                    placeholder="Account name"
                    value={l.accountName}
                    onChange={(e) => updateLine(l.id, 'accountName', e.target.value)}
                  />
                  <input
                    style={{ ...inputStyle, flex: 1 }}
                    type="number"
                    min="0"
                    step="0.01"
                    placeholder="Amount"
                    value={l.amount || ''}
                    onChange={(e) => updateLine(l.id, 'amount', parseFloat(e.target.value) || 0)}
                  />
                  {lines.length > 2 && (
                    <button
                      type="button"
                      style={{ ...ghostBtn, padding: '2px 6px', color: 'var(--finance-danger)' }}
                      onClick={() => removeLine(l.id)}
                    >
                      ✕
                    </button>
                  )}
                </div>
              ))}

              <div
                style={{
                  display: 'flex',
                  justifyContent: 'space-between',
                  padding: '8px 0',
                  fontSize: 12,
                  fontWeight: 600,
                  borderTop: '1px solid var(--hub-border)',
                  color:
                    Math.abs(totalDebits - totalCredits) < 0.01 && totalDebits > 0
                      ? 'var(--finance-accent)'
                      : 'var(--finance-danger)',
                }}
              >
                <span>
                  Debits: ${totalDebits.toFixed(2)} | Credits: ${totalCredits.toFixed(2)}
                </span>
                <span>
                  {Math.abs(totalDebits - totalCredits) < 0.01 && totalDebits > 0
                    ? 'Balanced ✓'
                    : totalDebits === 0
                      ? 'Enter amounts'
                      : `Imbalance: $${Math.abs(totalDebits - totalCredits).toFixed(2)}`}
                </span>
              </div>
            </div>
            <div style={{ display: 'flex', gap: 8, justifyContent: 'flex-end' }}>
              <button type="button" style={ghostBtn} onClick={() => setShowCreate(false)}>
                {t('cancel')}
              </button>
              <button
                type="button"
                style={primaryBtn}
                onClick={handleCreate}
                disabled={createMut.isPending}
              >
                {createMut.isPending ? t('creating') : t('create')}
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
              type="button"
              style={{ ...ghostBtn, padding: '4px 12px', opacity: page <= 1 ? 0.3 : 1 }}
              onClick={() => setPage((p) => Math.max(1, p - 1))}
              disabled={page <= 1}
            >
              Prev
            </button>
            <button
              type="button"
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
