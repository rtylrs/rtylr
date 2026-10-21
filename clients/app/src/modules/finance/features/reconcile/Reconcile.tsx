import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { pageTitle } from '../../shared/lib/pageTitle';
import { HelmetContainer } from '@voxire/components';
import { motion } from 'motion/react';
import { Link2, CheckCheck } from 'lucide-react';
import { toast } from 'sonner';
import { getOneKitErrorMessage, unwrap } from '@voxire/api';
import PageHeader from '../../shared/components/PageHeader';
import MoneyDisplay from '../../shared/components/MoneyDisplay';
import { useBankAccounts, useBankTransactions } from '../../shared/hooks/useBankAccounts';
import { Link } from '@tanstack/react-router';
import { SkeletonRow } from '../../shared/components/Skeleton';
import LoadError from '../../shared/components/LoadError';
import { useListInvoices } from '../../shared/hooks/queries/useInvoices';
import { cardStyle, ghostBtn, primaryBtn, ACCENT } from '../../shared/lib/tokens';
import { financeApi } from '../../shared/services/financeApi';
import { findExactMatches } from './reconcileHelper';

interface Candidate {
  id: string;
  label: string;
  amount: number;
}

const Reconcile: React.FC = () => {
  const { t } = useTranslation('financeCommon');
  const accountsQuery = useBankAccounts();
  const transactionsQuery = useBankTransactions();
  const transactions = transactionsQuery.data ?? [];
  const loading = accountsQuery.isLoading || transactionsQuery.isLoading;
  const failed = accountsQuery.isError || transactionsQuery.isError;
  const noAccounts = !loading && !failed && (accountsQuery.data ?? []).length === 0;
  const { data: invoicesResp } = useListInvoices({ page: 1, pageSize: 200, status: 'sent' });
  const candidates: Candidate[] = (invoicesResp?.invoices ?? []).map((inv) => ({
    id: String(inv.id),
    label: `${inv.invoiceNumber} - ${inv.clientName}`,
    amount: inv.total ?? 0,
  }));
  const [matches, setMatches] = useState<Record<string, string>>({});

  const unmatched = transactions.filter((t) => !t.matched && !matches[t.id ?? '']);

  const [autoMatching, setAutoMatching] = useState(false);

  const match = async (txId: string, candidateId: string): Promise<void> => {
    try {
      await unwrap(financeApi.matchReconcile(txId, candidateId));
      setMatches((prev) => ({ ...prev, [txId]: candidateId }));
      toast.success(`Matched transaction to ${candidateId}`);
    } catch (error) {
      toast.error(getOneKitErrorMessage(error, 'Failed to match transaction'));
    }
  };

  const handleAutoMatch = async (): Promise<void> => {
    setAutoMatching(true);
    try {
      const exactMatches = findExactMatches(
        unmatched.map((t) => ({
          id: t.id ?? '',
          description: t.description ?? '',
          amount: t.amount ?? 0,
        })),
        candidates
      );
      for (const item of exactMatches) {
        await unwrap(financeApi.matchReconcile(item.transactionId, item.candidate.id));
        // Record each match as it lands. A later call failing must not hide
        // the matches the server already accepted.
        setMatches((prev) => ({ ...prev, [item.transactionId]: item.candidate.id }));
      }
      toast.success(t('financeReconcile.autoMatched'));
    } catch (error) {
      toast.error(getOneKitErrorMessage(error, 'Auto-match failed'));
    } finally {
      setAutoMatching(false);
    }
  };

  return (
    <motion.div
      initial={{ opacity: 0, y: 4 }}
      animate={{ opacity: 1, y: 0 }}
      style={{ display: 'flex', flexDirection: 'column', gap: 20 }}
    >
      <HelmetContainer title={pageTitle(t('financeCommon.pageTitles.reconciliation'))} />
      <PageHeader
        title={t('financeReconcile.title')}
        subtitle={
          loading || failed || noAccounts
            ? undefined
            : t('financeReconcile.awaitingCount', { count: unmatched.length })
        }
        actions={
          unmatched.length > 0 ? (
            <button
              type="button"
              style={primaryBtn}
              onClick={handleAutoMatch}
              disabled={autoMatching}
            >
              <CheckCheck size={13} />{' '}
              {autoMatching ? t('financeReconcile.matching') : t('financeReconcile.autoMatchAll')}
            </button>
          ) : undefined
        }
      />

      {loading ? (
        <div style={{ ...cardStyle, padding: 20 }}>
          <SkeletonRow />
        </div>
      ) : failed ? (
        <LoadError
          title={
            accountsQuery.isError
              ? t('financeBankAccounts.loadFailedTitle')
              : t('financeReconcile.loadFailedTitle')
          }
          message={t('financeReconcile.loadFailed')}
          retrying={accountsQuery.isFetching || transactionsQuery.isFetching}
          onRetry={() => {
            void accountsQuery.refetch();
            void transactionsQuery.refetch();
          }}
        />
      ) : noAccounts ? (
        <div
          style={{ ...cardStyle, padding: 24, display: 'flex', flexDirection: 'column', gap: 10 }}
        >
          <h3 style={{ margin: 0, fontSize: 15, fontWeight: 600, color: 'var(--hub-text)' }}>
            {t('financeReconcile.noAccountsTitle')}
          </h3>
          <p style={{ margin: 0, fontSize: 13, color: 'var(--hub-text-muted)', lineHeight: 1.5 }}>
            {t('financeReconcile.noAccountsDescription')}
          </p>
          <Link
            to="/finance/bank"
            style={{ fontSize: 13, fontWeight: 600, color: 'var(--finance-accent)' }}
          >
            {t('financeReconcile.goToBankFeeds')}
          </Link>
        </div>
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
          {unmatched.map((t) => (
            <div key={t.id} style={{ ...cardStyle, padding: 16 }}>
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
                    {t.description}
                  </div>
                  <div style={{ fontSize: 11, color: 'var(--hub-text-dim)' }}>
                    {t.date} <>&middot;</> {t.category}
                  </div>
                </div>
                <div
                  style={{
                    fontSize: 16,
                    fontWeight: 700,
                    color: (t.amount ?? 0) >= 0 ? ACCENT : 'var(--hub-text)',
                  }}
                >
                  <MoneyDisplay amount={t.amount ?? 0} signed />
                </div>
              </div>
              <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap' }}>
                {candidates
                  .filter(
                    (c) =>
                      Math.abs(c.amount) === Math.abs(t.amount ?? 0) ||
                      Math.abs(Math.abs(c.amount) - Math.abs(t.amount ?? 0)) < 500
                  )
                  .map((c) => (
                    <button
                      type="button"
                      key={c.id}
                      style={{ ...ghostBtn, padding: '6px 12px' }}
                      onClick={() => match(t.id ?? '', c.id)}
                    >
                      <Link2 size={12} /> {c.label}
                    </button>
                  ))}
              </div>
            </div>
          ))}
          {unmatched.length === 0 && (
            <div
              style={{
                ...cardStyle,
                padding: 30,
                textAlign: 'center',
                color: 'var(--hub-text-muted)',
                fontSize: 13,
              }}
            >
              {transactions.length === 0
                ? t('financeReconcile.noTransactions')
                : t('financeReconcile.allDone')}
            </div>
          )}
        </div>
      )}
    </motion.div>
  );
};

export default Reconcile;
