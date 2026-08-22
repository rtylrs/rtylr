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
import {
  useChartOfAccounts,
  useJournal,
  useCreateJournalEntry,
} from '../../shared/hooks/useChartOfAccounts';
import { cardStyle, primaryBtn, ghostBtn, inputStyle } from '../../shared/lib/tokens';

interface LineInput {
  id: string;
  accountId: string;
  accountName: string;
  type: 'debit' | 'credit';
  amount: number;
  amountInvalid: boolean;
}

const DATE_PATTERN = /^\d{4}-\d{2}-\d{2}$/;

const makeInitialLines = (): LineInput[] => [
  {
    id: 'line-debit',
    accountId: '',
    accountName: '',
    type: 'debit',
    amount: 0,
    amountInvalid: false,
  },
  {
    id: 'line-credit',
    accountId: '',
    accountName: '',
    type: 'credit',
    amount: 0,
    amountInvalid: false,
  },
];

const Journal: React.FC = () => {
  const { t } = useTranslation('journal');
  const [page, setPage] = useState(1);
  const { data: journalData, isLoading } = useJournal({ page });
  const data = journalData?.entries ?? [];
  const hasMore = journalData?.hasMore ?? false;
  const totalEntries = journalData?.total;
  const createMut = useCreateJournalEntry();
  // Same source the Chart of Accounts screen uses - free-text account names
  // used to post as accountId "0", which no real account resolves to.
  const { data: accounts = [] } = useChartOfAccounts();
  const [showCreate, setShowCreate] = useState(false);
  const [memo, setMemo] = useState('');
  const [date, setDate] = useState(new Date().toISOString().slice(0, 10));

  const [lines, setLines] = useState<LineInput[]>(makeInitialLines());

  const addLine = (type: 'debit' | 'credit'): void => {
    setLines((prev) => [
      ...prev,
      {
        id: `line-${Date.now()}-${Math.random()}`,
        accountId: '',
        accountName: '',
        type,
        amount: 0,
        amountInvalid: false,
      },
    ]);
  };

  const removeLine = (id: string): void => {
    if (lines.length <= 2) {
      toast.error(t('minLines'));
      return;
    }
    setLines((prev) => prev.filter((l) => l.id !== id));
  };

  const selectLineAccount = (id: string, accountId: string): void => {
    const account = accounts.find((a) => a.id === accountId);
    setLines((prev) =>
      prev.map((l) => (l.id === id ? { ...l, accountId, accountName: account?.name ?? '' } : l))
    );
  };

  const updateLineAmount = (id: string, raw: string): void => {
    setLines((prev) =>
      prev.map((l) => {
        if (l.id !== id) return l;
        if (raw.trim() === '') return { ...l, amount: 0, amountInvalid: false };
        const parsed = Number(raw);
        // Flag junk instead of silently writing 0.
        if (!Number.isFinite(parsed) || parsed < 0) return { ...l, amountInvalid: true };
        return { ...l, amount: parsed, amountInvalid: false };
      })
    );
  };

  const totalDebits = lines
    .filter((l) => l.type === 'debit')
    .reduce((sum, l) => sum + (Number.isFinite(l.amount) ? l.amount : 0), 0);
  const totalCredits = lines
    .filter((l) => l.type === 'credit')
    .reduce((sum, l) => sum + (Number.isFinite(l.amount) ? l.amount : 0), 0);

  const dateValid = DATE_PATTERN.test(date);
  const allAccountsSelected = lines.every((l) => l.accountId !== '');
  const amountsValid = lines.every((l) => !l.amountInvalid);
  const canSubmit =
    dateValid && allAccountsSelected && amountsValid && memo.trim().length > 0 && totalDebits > 0;

  const handleCreate = async (): Promise<void> => {
    if (!memo.trim()) {
      toast.error(t('memoRequired'));
      return;
    }
    if (!dateValid) {
      toast.error(t('dateRequired'));
      return;
    }
    if (!allAccountsSelected) {
      toast.error(t('accountRequired'));
      return;
    }
    if (!amountsValid) {
      toast.error(t('amountInvalid'));
      return;
    }
    if (Math.abs(totalDebits - totalCredits) > 0.01 || totalDebits === 0) {
      toast.error(
        t('balanceRequired', {
          debits: totalDebits.toFixed(2),
          credits: totalCredits.toFixed(2),
        })
      );
      return;
    }
    const today = new Date().toISOString().slice(0, 10);
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
          accountName: l.accountName,
          debit: l.type === 'debit' ? l.amount : 0,
          credit: l.type === 'credit' ? l.amount : 0,
          description: '',
        })),
      });
      toast.success(t('created'));
      setShowCreate(false);
      setMemo('');
      setDate(new Date().toISOString().slice(0, 10));
      setLines(makeInitialLines());
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
        subtitle={
          totalEntries != null
            ? t('entriesCount', { count: totalEntries })
            : t('loadedCount', { count: data.length })
        }
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
            // JournalEntry carries no entry-level debit/credit totals on the
            // contract; balance must be computed from the entry's lines.
            const entryDebits = (e.lines ?? []).reduce((sum, l) => sum + (l.debit ?? 0), 0);
            const entryCredits = (e.lines ?? []).reduce((sum, l) => sum + (l.credit ?? 0), 0);
            const isBalanced = Math.abs(entryDebits - entryCredits) < 0.01;
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
                      {e.memo || e.description}
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
              style={{ ...inputStyle, fontSize: 16 }}
              type="date"
              value={date}
              max={new Date().toISOString().split('T')[0]}
              onChange={(e) => setDate(e.target.value)}
              aria-label={t('dateLabel')}
            />
            {!dateValid && (
              <p style={{ margin: 0, fontSize: 11, color: 'var(--finance-danger)' }}>
                {t('dateRequired')}
              </p>
            )}
            <input
              style={{ ...inputStyle, fontSize: 16 }}
              placeholder={t('memoPlaceholder')}
              value={memo}
              onChange={(e) => setMemo(e.target.value)}
              aria-label={t('memoPlaceholder')}
            />
            <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
              <div
                style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}
              >
                <span style={{ fontSize: 12, fontWeight: 600, color: 'var(--hub-text)' }}>
                  {t('entryLines')}
                </span>
                <div style={{ display: 'flex', gap: 6 }}>
                  <button
                    type="button"
                    style={{ ...ghostBtn, fontSize: 11, padding: '2px 8px' }}
                    onClick={() => addLine('debit')}
                  >
                    {t('addDebit')}
                  </button>
                  <button
                    type="button"
                    style={{ ...ghostBtn, fontSize: 11, padding: '2px 8px' }}
                    onClick={() => addLine('credit')}
                  >
                    {t('addCredit')}
                  </button>
                </div>
              </div>

              {lines.map((l) => (
                <div
                  key={l.id}
                  style={{ display: 'flex', gap: 6, alignItems: 'center', flexWrap: 'wrap' }}
                >
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
                  <select
                    style={{ ...inputStyle, flex: 2, fontSize: 16, width: 'auto' }}
                    value={l.accountId}
                    onChange={(e) => selectLineAccount(l.id, e.target.value)}
                    aria-label={`${t('selectAccount')} (${l.type})`}
                  >
                    <option value="">{t('selectAccount')}</option>
                    {accounts.map((a) => (
                      <option key={a.id} value={a.id}>
                        {(a.code ? `${a.code} · ` : '') + a.name}
                      </option>
                    ))}
                  </select>
                  <input
                    style={{
                      ...inputStyle,
                      flex: 1,
                      fontSize: 16,
                      width: 'auto',
                      ...(l.amountInvalid ? { border: '1px solid var(--finance-danger)' } : {}),
                    }}
                    type="number"
                    min="0"
                    step="0.01"
                    placeholder={t('amountPlaceholder')}
                    aria-label={`${t('amountPlaceholder')} (${l.type})`}
                    value={l.amount === 0 ? '' : String(l.amount)}
                    onChange={(e) => updateLineAmount(l.id, e.target.value)}
                  />
                  {lines.length > 2 && (
                    <button
                      type="button"
                      style={{ ...ghostBtn, padding: '2px 6px', color: 'var(--finance-danger)' }}
                      onClick={() => removeLine(l.id)}
                      aria-label={t('removeLine')}
                    >
                      ✕
                    </button>
                  )}
                  {l.amountInvalid && (
                    <span
                      role="alert"
                      style={{
                        width: '100%',
                        fontSize: 11,
                        color: 'var(--finance-danger)',
                      }}
                    >
                      {t('amountInvalid')}
                    </span>
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
                  color: canSubmit ? 'var(--finance-accent)' : 'var(--finance-danger)',
                }}
              >
                <span>
                  {t('debitsCredits', {
                    debits: totalDebits.toFixed(2),
                    credits: totalCredits.toFixed(2),
                  })}
                </span>
                <span>
                  {canSubmit
                    ? t('balancedOk')
                    : totalDebits === 0
                      ? t('enterAmounts')
                      : t('imbalance', {
                          amount: Math.abs(totalDebits - totalCredits).toFixed(2),
                        })}
                </span>
              </div>
            </div>
            <div style={{ display: 'flex', gap: 8, justifyContent: 'flex-end' }}>
              <button type="button" style={ghostBtn} onClick={() => setShowCreate(false)}>
                {t('cancel')}
              </button>
              <button
                type="button"
                style={{ ...primaryBtn, opacity: canSubmit ? 1 : 0.5 }}
                onClick={() => void handleCreate()}
                disabled={!canSubmit || createMut.isPending}
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
          <span style={{ fontSize: 12, color: 'var(--hub-text-dim)' }}>
            {t('pageLabel', { page })}
          </span>
          <div style={{ display: 'flex', gap: 8 }}>
            <button
              type="button"
              style={{ ...ghostBtn, padding: '4px 12px', opacity: page <= 1 ? 0.3 : 1 }}
              onClick={() => setPage((p) => Math.max(1, p - 1))}
              disabled={page <= 1}
            >
              {t('prev')}
            </button>
            <button
              type="button"
              style={{ ...ghostBtn, padding: '4px 12px', opacity: !hasMore ? 0.3 : 1 }}
              onClick={() => setPage((p) => p + 1)}
              disabled={!hasMore}
            >
              {t('next')}
            </button>
          </div>
        </div>
      )}
    </motion.div>
  );
};

export default Journal;
