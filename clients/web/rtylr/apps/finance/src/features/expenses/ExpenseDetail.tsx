import React from 'react';
import { useTranslation } from 'react-i18next';
import { pageTitle } from '../../shared/lib/pageTitle';
import { HelmetContainer } from '@voxire/components';
import { useParams } from '@tanstack/react-router';
import { useQueryClient } from '@tanstack/react-query';
import { useUserStore } from '@voxire/auth';
import { createOneKitClient, getOneKitErrorMessage } from '@voxire/api';
import { ExpensesServiceClient } from '@voxire/voxie/finance/expenses/v1/client';
import { motion } from 'motion/react';
import { toast } from 'sonner';
import { CheckCircle, XCircle } from 'lucide-react';
import PageHeader from '../../shared/components/PageHeader';
import MoneyDisplay from '../../shared/components/MoneyDisplay';
import { SkeletonRow } from '../../shared/components/Skeleton';
import ReceiptUpload from '../../shared/components/ReceiptUpload';
import { useExpenseById } from './useExpenseById';
import { useExpenseCategories } from './useExpenses';
import { useApproveExpense, useRejectExpense } from '../../shared/hooks/queries/useExpenses';
import { cardStyle, ghostBtn, primaryBtn } from '../../shared/lib/tokens';
import { canFinancePermission } from '../../shared/lib/financePermissions';

const expensesClient = createOneKitClient(ExpensesServiceClient, 'finance');

const statusStyle: Record<string, React.CSSProperties> = {
  pending: {
    background: 'color-mix(in srgb, var(--finance-warning) 12%, transparent)',
    color: 'var(--finance-warning)',
  },
  approved: {
    background: 'color-mix(in srgb, var(--finance-success) 12%, transparent)',
    color: 'var(--finance-success)',
  },
  rejected: {
    background: 'color-mix(in srgb, var(--finance-error) 12%, transparent)',
    color: 'var(--finance-error)',
  },
};

const ExpenseDetail: React.FC = () => {
  const { t } = useTranslation();
  const { id } = useParams({ strict: false });
  const { data: expense, isLoading, isError, error, isFetching, refetch } = useExpenseById(id);
  const { data: categories = [] } = useExpenseCategories();
  const queryClient = useQueryClient();
  const approveMut = useApproveExpense();
  const rejectMut = useRejectExpense();
  const [receipt, setReceipt] = React.useState<string | null>(null);
  const [savingReceipt, setSavingReceipt] = React.useState(false);
  const role = useUserStore((state) => state.user?.role ?? state.role);
  const canApproveExpense = canFinancePermission(role, 'finance:expenses:approve');

  const categoryName = React.useMemo(() => {
    if (expense?.categoryId == null) return 'Uncategorized';
    return categories.find((c) => c.id === expense.categoryId)?.name ?? 'Uncategorized';
  }, [categories, expense]);

  React.useEffect(() => {
    setReceipt(expense?.receiptUrl ?? null);
  }, [expense]);

  const handleReceiptChange = async (receiptUrl: string | null): Promise<void> => {
    if (!expense || savingReceipt) return;

    setSavingReceipt(true);
    try {
      await expensesClient.updateExpense({
        id: expense.id,
        body: {
          categoryId: expense.categoryId != null ? String(expense.categoryId) : undefined,
          amount: expense.amount,
          currency: expense.currency ?? 'USD',
          description: expense.description,
          vendor: expense.vendor ?? '',
          receiptUrl: receiptUrl ?? '',
          paymentMethod: expense.paymentMethod ?? '',
          isTaxDeductible: expense.isTaxDeductible ?? false,
          taxAmount: expense.taxAmount ?? 0,
          notes: expense.notes ?? '',
          expenseDate: expense.expenseDate,
        },
      });
      setReceipt(receiptUrl);
      toast.success(t('expenses.updated'));
      void queryClient.invalidateQueries({ queryKey: ['expenses'] });
      void refetch();
    } catch (error) {
      toast.error(getOneKitErrorMessage(error, t('errors.generic')));
    } finally {
      setSavingReceipt(false);
    }
  };

  if (isError && !expense) {
    return (
      <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
        <PageHeader
          title={t('expenses.detailTitle')}
          breadcrumbs={[{ label: t('nav.expenses'), to: '/expenses' }, { label: id ?? '' }]}
        />
        <div
          role="alert"
          style={{ ...cardStyle, padding: 20, display: 'flex', flexDirection: 'column', gap: 12 }}
        >
          <p style={{ margin: 0, fontSize: 13, color: 'var(--hub-text)' }}>
            {getOneKitErrorMessage(error, t('errors.generic'))}
          </p>
          <button
            type="button"
            style={ghostBtn}
            disabled={isFetching}
            onClick={() => void refetch()}
          >
            {t('common.refresh')}
          </button>
        </div>
      </div>
    );
  }

  if (isLoading || !expense) {
    return (
      <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
        <PageHeader
          title={t('expenses.detailTitle')}
          breadcrumbs={[{ label: t('nav.expenses'), to: '/expenses' }, { label: id ?? '' }]}
        />
        <div style={{ ...cardStyle, padding: 20 }}>
          <SkeletonRow />
        </div>
      </div>
    );
  }

  return (
    <motion.div
      initial={{ opacity: 0, y: 4 }}
      animate={{ opacity: 1, y: 0 }}
      style={{ display: 'flex', flexDirection: 'column', gap: 20 }}
    >
      <HelmetContainer title={pageTitle(t('common.pageTitles.expenseDetail'))} />
      <PageHeader
        title={`Expense ${expense.id}`}
        subtitle={
          <span style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
            <span
              style={{
                fontSize: 11,
                fontWeight: 600,
                padding: '3px 10px',
                borderRadius: 20,
                textTransform: 'capitalize',
                ...(statusStyle[expense.status] ?? statusStyle.pending),
              }}
            >
              {expense.status}
            </span>
            <MoneyDisplay amount={expense.amount} />
          </span>
        }
        breadcrumbs={[{ label: t('nav.expenses'), to: '/expenses' }, { label: expense.id }]}
        actions={
          canApproveExpense ? (
            <>
              <button
                type="button"
                style={{
                  ...ghostBtn,
                  color: 'var(--finance-error)',
                  borderColor: 'color-mix(in srgb, var(--finance-error) 30%, transparent)',
                }}
                disabled={rejectMut.isPending}
                onClick={() => rejectMut.mutate(expense.id)}
              >
                <XCircle size={13} /> Reject
              </button>
              <button
                type="button"
                style={primaryBtn}
                disabled={approveMut.isPending}
                onClick={() => approveMut.mutate(expense.id)}
              >
                <CheckCircle size={13} /> Approve
              </button>
            </>
          ) : undefined
        }
      />

      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 20 }}>
        <div style={{ ...cardStyle, padding: 20 }}>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
            <Field label={t('common.date')} value={expense.expenseDate} />
            <Field label={t('common.description')} value={expense.description} />
            <Field label={t('common.category')} value={categoryName} />
            <Field label={t('common.amount')}>
              <MoneyDisplay amount={expense.amount} />
            </Field>
            <Field label={t('common.vendor')} value={expense.vendor ?? undefined} />
            <div
              style={{
                marginTop: 10,
                padding: 12,
                background: 'var(--hub-hover)',
                borderRadius: 8,
                fontSize: 12,
                color: 'var(--hub-text-muted)',
              }}
            >
              {expense.notes}
            </div>
          </div>
        </div>
        <div style={{ ...cardStyle, padding: 20 }}>
          <div
            style={{
              fontSize: 11,
              color: 'var(--hub-text-dim)',
              textTransform: 'uppercase',
              letterSpacing: '0.08em',
              marginBottom: 12,
            }}
          >
            {t('common.receipt')}
          </div>
          <div aria-busy={savingReceipt}>
            <div style={{ pointerEvents: savingReceipt ? 'none' : 'auto' }}>
              <ReceiptUpload
                value={receipt}
                onChange={(receiptUrl) => void handleReceiptChange(receiptUrl)}
              />
            </div>
            {savingReceipt && (
              <div
                role="status"
                style={{ marginTop: 8, color: 'var(--hub-text-dim)', fontSize: 12 }}
              >
                {t('common.saving')}
              </div>
            )}
          </div>
        </div>
      </div>
    </motion.div>
  );
};

const Field: React.FC<{ label: string; value?: string; children?: React.ReactNode }> = ({
  label,
  value,
  children,
}) => (
  <div
    style={{
      display: 'flex',
      justifyContent: 'space-between',
      padding: '6px 0',
      borderBottom: '1px solid var(--hub-border)',
    }}
  >
    <span style={{ fontSize: 12, color: 'var(--hub-text-muted)' }}>{label}</span>
    <span style={{ fontSize: 13, color: 'var(--hub-text)', fontWeight: 600 }}>
      {value ?? children}
    </span>
  </div>
);

export default ExpenseDetail;
