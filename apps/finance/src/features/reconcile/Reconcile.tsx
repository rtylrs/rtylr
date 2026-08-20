import React, { useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { useUserStore } from '@voxire/auth';
import { motion } from 'motion/react';
import { Link2, CheckCheck } from 'lucide-react';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { getOneKitErrorMessage, unwrap } from '@voxire/api';
import { EmptyState } from '@voxire/ui';
import PageHeader from '../../shared/components/PageHeader';
import MoneyDisplay from '../../shared/components/MoneyDisplay';
import { useBankTransactions } from '../../shared/hooks/useBankAccounts';
import { useListInvoices } from '../../shared/hooks/queries/useInvoices';
import { cardStyle, ghostBtn, primaryBtn, ACCENT } from '../../shared/lib/tokens';
import { safe } from '../../shared/lib/api';
import { canFinancePermission } from '../../shared/lib/financePermissions';
import { findExactMatches } from './reconcileHelper';

interface Candidate {
  id: string;
  label: string;
  amount: number;
}

const Reconcile: React.FC = () => {
  const { t } = useTranslation();
  const { data: transactions = [], refetch: refetchTransactions } = useBankTransactions();
  const { data: invoicesResp } = useListInvoices({ page: 1, pageSize: 200, status: 'sent' });
  const candidates: Candidate[] = (invoicesResp?.invoices ?? []).map((inv) => ({
    id: String(inv.id),
    label: `${inv.invoiceNumber} - ${inv.clientName}`,
    amount: inv.total ?? 0,
  }));
  const [matches, setMatches] = useState<Record<string, string>>({});
  const role = useUserStore((state) => state.user?.role ?? state.role);
  const canReconcile = canFinancePermission(role, 'finance:reconcile');

  const unmatched = transactions.filter((t) => !t.matched && !matches[t.id ?? '']);

  const [autoMatching, setAutoMatching] = useState(false);

  const match = async (txId: string, candidateId: string): Promise<void> => {
    if (!canReconcile) return;
    try {
      await unwrap(
        safe.post('/finance/reconcile/match', {
          transactionId: txId,
          targetType: 'invoice',
          targetId: candidateId,
        })
      );
      setMatches((prev) => ({ ...prev, [txId]: candidateId }));
      toast.success(`Matched transaction to ${candidateId}`);
    } catch (error) {
      toast.error(getOneKitErrorMessage(error, 'Failed to match transaction'));
    }
  };

  const handleAutoMatch = async (): Promise<void> => {
    if (!canReconcile) return;
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
        await unwrap(
          safe.post('/finance/reconcile/match', {
            transactionId: item.transactionId,
            targetType: 'invoice',
            targetId: item.candidate.id,
          })
        );
      }
      setMatches((prev) => {
        const next = { ...prev };
        for (const item of exactMatches) next[item.transactionId] = item.candidate.id;
        return next;
      });
      toast.success('Auto-match completed');
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
      <HelmetContainer title="Reconciliation · rtylr" />
      <PageHeader
        title="Reconciliation"
        subtitle={`${unmatched.length} transactions awaiting match`}
        actions={
          canReconcile ? (
            <button style={primaryBtn} onClick={handleAutoMatch} disabled={autoMatching}>
              <CheckCheck size={13} /> {autoMatching ? 'Matching…' : 'Auto-match all'}
            </button>
          ) : undefined
        }
      />

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
            {canReconcile && (
              <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap' }}>
                {candidates
                  .filter(
                    (c) =>
                      Math.abs(c.amount) === Math.abs(t.amount ?? 0) ||
                      Math.abs(Math.abs(c.amount) - Math.abs(t.amount ?? 0)) < 500
                  )
                  .map((c) => (
                    <button
                      key={c.id}
                      style={{ ...ghostBtn, padding: '6px 12px' }}
                      onClick={() => match(t.id ?? '', c.id)}
                    >
                      <Link2 size={12} /> {c.label}
                    </button>
                  ))}
              </div>
            )}
          </div>
        ))}
        {unmatched.length === 0 && (
          <EmptyState
            icon={CheckCheck}
            title={t('reconcile.emptyTitle')}
            description={t('reconcile.emptyDescription')}
            action={{
              label: t('common.refresh'),
              onClick: () => void refetchTransactions(),
            }}
          />
        )}
      </div>
    </motion.div>
  );
};

export default Reconcile;
