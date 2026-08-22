import React, { useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { motion } from 'motion/react';
import { Link2, CheckCheck } from 'lucide-react';
import { toast } from 'sonner';
import { getOneKitErrorMessage, unwrap } from '@voxire/api';
import { useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import PageHeader from '../../shared/components/PageHeader';
import MoneyDisplay from '../../shared/components/MoneyDisplay';
import ConfirmDialog from '../../shared/components/ConfirmDialog';
import { useBankTransactions } from '../../shared/hooks/useBankAccounts';
import { useListInvoices } from '../../shared/hooks/queries/useInvoices';
import { cardStyle, ghostBtn, primaryBtn, ACCENT } from '../../shared/lib/tokens';
import { financeApi } from '../../shared/services/financeApi';
import { findExactMatches } from './reconcileHelper';

/** Exact-match only: amounts within one cent count, anything looser does not. */
const EXACT_MATCH_TOLERANCE = 0.01;

interface Candidate {
  id: string;
  label: string;
  amount: number;
}

const Reconcile: React.FC = () => {
  const { t } = useTranslation('reconcile');
  const queryClient = useQueryClient();
  const { data: transactions = [] } = useBankTransactions();
  // App-canonical invoice status casing ('Sent'); neither the generated nor the
  // compat contract constrains the filter's casing, so the canonical wins.
  const { data: invoicesResp } = useListInvoices({ page: 1, pageSize: 200, status: 'Sent' });
  const candidates: Candidate[] = (invoicesResp?.invoices ?? []).map((inv) => ({
    id: String(inv.id),
    label: `${inv.invoiceNumber} - ${inv.clientName}`,
    amount: inv.total ?? 0,
  }));
  const [matches, setMatches] = useState<Record<string, string>>({});

  const unmatched = transactions.filter((tx) => !tx.matched && !matches[tx.id ?? '']);

  const [autoMatching, setAutoMatching] = useState(false);
  const [showAutoMatchConfirm, setShowAutoMatchConfirm] = useState(false);

  const invalidateMatchData = (): Promise<unknown>[] => [
    queryClient.invalidateQueries({ queryKey: ['bank-tx'] }),
    queryClient.invalidateQueries({ queryKey: ['invoices'] }),
  ];

  const match = async (txId: string, candidateId: string): Promise<void> => {
    try {
      await unwrap(financeApi.matchReconcile(txId, candidateId));
      setMatches((prev) => ({ ...prev, [txId]: candidateId }));
      await Promise.all(invalidateMatchData());
      toast.success(t('matched', { candidate: candidateId }));
    } catch (error) {
      toast.error(getOneKitErrorMessage(error, t('matchFailed')));
    }
  };

  // Applies local matches per successful item and reports partial outcomes:
  // earlier successes live on the server even when a later call fails.
  const runAutoMatch = async (): Promise<void> => {
    setAutoMatching(true);
    try {
      const usedCandidateIds = new Set(Object.values(matches));
      const exactMatches = findExactMatches(
        unmatched.map((tx) => ({
          id: tx.id ?? '',
          description: tx.description ?? '',
          amount: tx.amount ?? 0,
        })),
        candidates,
        usedCandidateIds
      );
      let linked = 0;
      let failed = 0;
      for (const item of exactMatches) {
        try {
          await unwrap(financeApi.matchReconcile(item.transactionId, item.candidate.id));
          setMatches((prev) => ({ ...prev, [item.transactionId]: item.candidate.id }));
          linked += 1;
        } catch {
          failed += 1;
        }
      }
      if (linked > 0) void Promise.all(invalidateMatchData());
      if (linked === 0 && failed === 0) {
        toast.info(t('noExactMatches'));
      } else if (failed > 0) {
        toast.warning(t('autoMatchPartial', { linked, failed }));
      } else {
        toast.success(t('autoMatchDone', { linked }));
      }
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
      <HelmetContainer title="Reconciliation · rtylr" />
      <PageHeader
        title="Reconciliation"
        subtitle={t('awaitingMatch', { count: unmatched.length })}
        actions={
          <button
            type="button"
            style={primaryBtn}
            onClick={() => setShowAutoMatchConfirm(true)}
            disabled={autoMatching}
          >
            <CheckCheck size={13} /> {autoMatching ? t('matching') : t('autoMatch')}
          </button>
        }
      />

      <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
        {unmatched.map((tx) => (
          <div key={tx.id} style={{ ...cardStyle, padding: 16 }}>
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
                  {tx.description}
                </div>
                <div style={{ fontSize: 11, color: 'var(--hub-text-dim)' }}>
                  {tx.date} <>&middot;</> {tx.category}
                </div>
              </div>
              <div
                style={{
                  fontSize: 16,
                  fontWeight: 700,
                  color: (tx.amount ?? 0) >= 0 ? ACCENT : 'var(--hub-text)',
                }}
              >
                <MoneyDisplay amount={tx.amount ?? 0} signed />
              </div>
            </div>
            <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap' }}>
              {candidates
                .filter(
                  (c) =>
                    !matches[tx.id ?? ''] &&
                    Object.values(matches).indexOf(c.id) === -1 &&
                    Math.abs(Math.abs(c.amount) - Math.abs(tx.amount ?? 0)) <= EXACT_MATCH_TOLERANCE
                )
                .map((c) => (
                  <button
                    type="button"
                    key={c.id}
                    style={{ ...ghostBtn, padding: '6px 12px' }}
                    onClick={() => match(tx.id ?? '', c.id)}
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
            {t('emptyTitle')}
          </div>
        )}
      </div>

      <ConfirmDialog
        open={showAutoMatchConfirm}
        onClose={() => setShowAutoMatchConfirm(false)}
        onConfirm={() => void runAutoMatch()}
        title={t('autoMatch')}
        message={t('autoMatchConfirmMessage', { count: unmatched.length })}
        confirmLabel={t('autoMatch')}
      />
    </motion.div>
  );
};

export default Reconcile;
