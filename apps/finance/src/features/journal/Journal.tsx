import React, { useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { motion } from 'motion/react';
import { Plus } from 'lucide-react';
import { toast } from 'sonner';
import { getOneKitErrorMessage } from '@voxire/api';
import { useTranslation } from 'react-i18next';
import PageHeader from '../../shared/components/PageHeader';
import MoneyDisplay from '../../shared/components/MoneyDisplay';
import CurrencyInput from '../../shared/components/CurrencyInput';
import { SkeletonRow } from '../../shared/components/Skeleton';
import { useJournal, useCreateJournalEntry } from '../../shared/hooks/useChartOfAccounts';
import { cardStyle, primaryBtn, ghostBtn, inputStyle } from '../../shared/lib/tokens';

const Journal: React.FC = () => {
  const { t } = useTranslation('journal');
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
  const [amount, setAmount] = useState(0);

  const handleCreate = async (): Promise<void> => {
    if (!memo.trim()) {
      toast.error(t('memoRequired'));
      return;
    }
    if (amount <= 0) {
      toast.error(t('amountZero'));
      return;
    }
    const today = new Date().toISOString().slice(0, 10);
    if (date > today) {
      toast.error(t('futureDate'));
      return;
    }
    const amt = amount;
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
      toast.success(t('created'));
      setShowCreate(false);
      setMemo('');
      setAmount(0);
      setDebitAcc('');
      setDebitAccId(0);
      setCreditAcc('');
      setCreditAccId(0);
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
          <button style={primaryBtn} onClick={() => setShowCreate(true)}>
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
              width: 400,
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
              max={new Date().toISOString().split('T')[0]}
              onChange={(e) => setDate(e.target.value)}
            />
            <input
              style={inputStyle}
              placeholder={t('memoPlaceholder')}
              value={memo}
              onChange={(e) => setMemo(e.target.value)}
            />
            <input
              style={inputStyle}
              placeholder={t('debitAccPlaceholder')}
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
              placeholder={t('creditAccPlaceholder')}
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
            <CurrencyInput value={amount} onChange={setAmount} />
            <div style={{ display: 'flex', gap: 8, justifyContent: 'flex-end' }}>
              <button style={ghostBtn} onClick={() => setShowCreate(false)}>
                {t('cancel')}
              </button>
              <button style={primaryBtn} onClick={handleCreate} disabled={createMut.isPending}>
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
