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
  useJournal,
  useCreateJournalEntry,
  useChartOfAccounts,
} from '../../shared/hooks/useChartOfAccounts';
import { cardStyle, primaryBtn, ghostBtn, inputStyle } from '../../shared/lib/tokens';
import { todayIso } from '../../shared/lib/date';

interface LineInput {
  id: string;
  accountName: string;
  accountId: number;
  type: 'debit' | 'credit';
  amount: number;
}

interface JournalLineValues {
  debit?: number | null;
  credit?: number | null;
}

export function toJournalMinorUnits(amount: number): number | null {
  if (!Number.isFinite(amount) || amount < 0) return null;

  const cents = amount * 100;
  const rounded = Math.round(cents);
  if (Math.abs(cents - rounded) > 1e-8 || !Number.isSafeInteger(rounded)) return null;
  return rounded;
}

export function journalTotals(lines: JournalLineValues[]) {
  let debitMinor = 0;
  let creditMinor = 0;
  let valid = true;

  for (const line of lines) {
    const debit = toJournalMinorUnits(line.debit ?? 0);
    const credit = toJournalMinorUnits(line.credit ?? 0);
    if (debit === null || credit === null || (debit === 0 && credit === 0)) {
      valid = false;
    }
    debitMinor += debit ?? 0;
    creditMinor += credit ?? 0;
  }

  return { debitMinor, creditMinor, valid };
}

const Journal: React.FC = () => {
  const { t } = useTranslation('journal');
  const [page, setPage] = useState(1);
  const { data: journalData, isLoading, isError, refetch } = useJournal({ page });
  const { data: accounts = [] } = useChartOfAccounts();
  const data = journalData?.entries ?? [];
  const hasMore = !isError && (journalData?.hasMore ?? false);
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
      toast.error(t('journal.twoLinesRequired'));
      return;
    }
    setLines((prev) => prev.filter((l) => l.id !== id));
  };

  const updateLine = (id: string, field: keyof LineInput, value: unknown): void => {
    setLines((prev) => prev.map((l) => (l.id === id ? { ...l, [field]: value } : l)));
  };

  const selectAccount = (id: string, accountId: string): void => {
    const account = accounts.find((candidate) => candidate.id === accountId);
    setLines((prev) =>
      prev.map((line) =>
        line.id === id
          ? {
              ...line,
              accountId: Number(accountId) || 0,
              accountName: account?.name ?? '',
            }
          : line
      )
    );
  };

  const formTotals = journalTotals(
    lines.map((line) => ({
      debit: line.type === 'debit' ? line.amount : 0,
      credit: line.type === 'credit' ? line.amount : 0,
    }))
  );
  const totalDebits = formTotals.debitMinor / 100;
  const totalCredits = formTotals.creditMinor / 100;
  const isBalanced =
    formTotals.valid &&
    formTotals.debitMinor > 0 &&
    formTotals.debitMinor === formTotals.creditMinor;

  const handleCreate = async (): Promise<void> => {
    if (!memo.trim()) {
      toast.error(t('journal.memoRequired'));
      return;
    }
    if (lines.some((line) => !Number.isSafeInteger(line.accountId) || line.accountId <= 0)) {
      toast.error(`${t('journal.accountName')}: account ID must be a positive integer`);
      return;
    }
    if (!formTotals.valid) {
      toast.error(t('journal.amountZero'));
      return;
    }
    if (!isBalanced) {
      toast.error(
        `Journal entry must balance: Debits (${totalDebits.toFixed(2)}) must equal Credits (${totalCredits.toFixed(2)})`
      );
      return;
    }
    const today = todayIso();
    if (date > today) {
      toast.error(t('journal.futureDate'));
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
      toast.success(t('journal.created'));
      setShowCreate(false);
      setMemo('');
      setLines([
        { id: '1', accountName: '', accountId: 0, type: 'debit', amount: 0 },
        { id: '2', accountName: '', accountId: 0, type: 'credit', amount: 0 },
      ]);
    } catch (error) {
      toast.error(getOneKitErrorMessage(error, t('journal.createFailed')));
    }
  };

  return (
    <motion.div
      initial={{ opacity: 0, y: 4 }}
      animate={{ opacity: 1, y: 0 }}
      style={{ display: 'flex', flexDirection: 'column', gap: 20 }}
    >
      <HelmetContainer title={`${t('journal.title')} · rtylr`} />
      <PageHeader
        title={t('journal.title')}
        subtitle={!isLoading && !isError ? `${data.length} entries` : undefined}
        actions={
          <button type="button" style={primaryBtn} onClick={() => setShowCreate(true)}>
            <Plus size={13} /> {t('journal.newEntry')}
          </button>
        }
      />

      {isLoading ? (
        <div style={{ ...cardStyle, padding: 20 }}>
          <SkeletonRow />
        </div>
      ) : isError ? (
        <div
          role="alert"
          style={{ ...cardStyle, padding: 20, display: 'flex', flexDirection: 'column', gap: 12 }}
        >
          <p style={{ margin: 0, fontSize: 13, color: 'var(--hub-text)' }}>{t('errors.generic')}</p>
          <button type="button" style={ghostBtn} onClick={() => void refetch?.()}>
            {t('common.refresh')}
          </button>
        </div>
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
          {data.map((e) => {
            const entryTotals =
              e.debit != null && e.credit != null
                ? {
                    debitMinor: toJournalMinorUnits(e.debit) ?? 0,
                    creditMinor: toJournalMinorUnits(e.credit) ?? 0,
                    valid:
                      toJournalMinorUnits(e.debit) !== null &&
                      toJournalMinorUnits(e.credit) !== null,
                  }
                : journalTotals(e.lines ?? []);
            const isBalanced =
              entryTotals.valid &&
              entryTotals.debitMinor > 0 &&
              entryTotals.debitMinor === entryTotals.creditMinor;
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
                      {e.description ?? e.memo ?? ''}
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
                    {isBalanced ? t('journal.balanced') : t('journal.outOfBalance')}
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
                        {t('journal.account')}
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
                        {t('journal.debit')}
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
                        {t('journal.credit')}
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
              {t('journal.newJournalEntry')}
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
              placeholder={t('journal.memoPlaceholder')}
              value={memo}
              onChange={(e) => setMemo(e.target.value)}
            />
            <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
              <div
                style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}
              >
                <span style={{ fontSize: 12, fontWeight: 600, color: 'var(--hub-text)' }}>
                  {t('journal.entryLines')}
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

              {lines.map((l, lineIndex) => (
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
                  <select
                    style={{ ...inputStyle, flex: 2 }}
                    aria-label={`${t('journal.account')} ${lineIndex + 1}`}
                    value={l.accountId || ''}
                    onChange={(e) => selectAccount(l.id, e.target.value)}
                  >
                    <option value="">{t('journal.accountName')}</option>
                    {l.accountId > 0 &&
                      !accounts.some((account) => account.id === String(l.accountId)) && (
                        <option value={l.accountId}>
                          {l.accountName || `Account ${l.accountId}`}
                        </option>
                      )}
                    {accounts.map((account) => (
                      <option key={account.id} value={account.id}>
                        {account.code ? `${account.code} · ` : ''}
                        {account.name}
                        {account.type ? ` (${account.type})` : ''}
                      </option>
                    ))}
                  </select>
                  <input
                    style={{ ...inputStyle, flex: 1 }}
                    type="number"
                    min="0.01"
                    step="0.01"
                    placeholder={t('common.amount')}
                    value={l.amount || ''}
                    onChange={(e) => updateLine(l.id, 'amount', Number(e.target.value) || 0)}
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
                  color: isBalanced ? 'var(--finance-accent)' : 'var(--finance-danger)',
                }}
              >
                <span>
                  Debits: ${totalDebits.toFixed(2)} | Credits: ${totalCredits.toFixed(2)}
                </span>
                <span>
                  {isBalanced
                    ? 'Balanced ✓'
                    : !formTotals.valid
                      ? t('journal.amountZero')
                      : `Imbalance: $${Math.abs(totalDebits - totalCredits).toFixed(2)}`}
                </span>
              </div>
            </div>
            <div style={{ display: 'flex', gap: 8, justifyContent: 'flex-end' }}>
              <button type="button" style={ghostBtn} onClick={() => setShowCreate(false)}>
                {t('journal.cancel')}
              </button>
              <button
                type="button"
                style={primaryBtn}
                onClick={handleCreate}
                disabled={createMut.isPending}
              >
                {createMut.isPending ? t('journal.creating') : t('journal.create')}
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
              {t('common.prev')}
            </button>
            <button
              type="button"
              style={{ ...ghostBtn, padding: '4px 12px', opacity: !hasMore ? 0.3 : 1 }}
              onClick={() => setPage((p) => p + 1)}
              disabled={!hasMore}
            >
              {t('common.next')}
            </button>
          </div>
        </div>
      )}
    </motion.div>
  );
};

export default Journal;
