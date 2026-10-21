import React from 'react';
import { useTranslation } from 'react-i18next';
import { pageTitle } from '../../shared/lib/pageTitle';
import { HelmetContainer } from '@voxire/components';
import { useNavigate, useParams } from '@tanstack/react-router';
import { useQueryClient } from '@tanstack/react-query';
import { createOneKitClient, getOneKitErrorMessage } from '@voxire/api';
import { ExpensesServiceClient } from '@voxire/voxie/finance/expenses/v1/client';
import { motion } from 'motion/react';
import { toast } from 'sonner';
import { CheckCircle, ExternalLink, Trash2, XCircle } from 'lucide-react';
import PageHeader from '../../shared/components/PageHeader';
import { SkeletonRow } from '../../shared/components/Skeleton';
import ReceiptUpload from '../../shared/components/ReceiptUpload';
import { useExpenseById } from './useExpenseById';
import { useDeleteExpense, useExpenseCategories } from './useExpenses';
import ConfirmDialog from '../../shared/components/ConfirmDialog';
import { formatMoney, type Currency } from '../../shared/lib/money';
import { formatDate } from '../../shared/lib/date';
import { useSettings } from '../../shared/lib/settings';
import { expenseStatusStyle } from './expenseDisplay';
import { useApproveExpense, useRejectExpense } from '../../shared/hooks/queries/useExpenses';
import { cardStyle, ghostBtn, primaryBtn } from '../../shared/lib/tokens';
import { canFinancePermission } from '../../shared/lib/financePermissions';

const expensesClient = createOneKitClient(ExpensesServiceClient, 'finance');

const ExpenseDetail: React.FC = () => {
  const { t } = useTranslation('financeCommon');
  const { id } = useParams({ strict: false });
  const { data: expense, isLoading, isError, error, isFetching, refetch } = useExpenseById(id);
  const { data: categories = [] } = useExpenseCategories();
  const queryClient = useQueryClient();
  const approveMut = useApproveExpense();
  const rejectMut = useRejectExpense();
  const [receipt, setReceipt] = React.useState<string | null>(null);
  const [savingReceipt, setSavingReceipt] = React.useState(false);
  const canApproveExpense = canFinancePermission('finance:expenses:approve');
  const navigate = useNavigate();
  const deleteMut = useDeleteExpense();
  const [confirmDelete, setConfirmDelete] = React.useState(false);
  const homeCurrency = useSettings((st) => st.currency);

  const category = React.useMemo(
    () =>
      expense?.categoryId == null ? undefined : categories.find((c) => c.id === expense.categoryId),
    [categories, expense]
  );

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
      toast.success(t('financeExpenses.updated'));
      void queryClient.invalidateQueries({ queryKey: ['finance', 'expenses'] });
      void refetch();
    } catch (error) {
      toast.error(getOneKitErrorMessage(error, t('financeErrors.generic')));
    } finally {
      setSavingReceipt(false);
    }
  };

  if (isError && !expense) {
    return (
      <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
        <PageHeader
          title={t('financeExpenses.detailTitle')}
          breadcrumbs={[
            { label: t('financeNav.expenses'), to: '/finance/expenses' },
            { label: id ?? '' },
          ]}
        />
        <div
          role="alert"
          style={{ ...cardStyle, padding: 20, display: 'flex', flexDirection: 'column', gap: 12 }}
        >
          <p style={{ margin: 0, fontSize: 13, color: 'var(--hub-text)' }}>
            {getOneKitErrorMessage(error, t('financeErrors.generic'))}
          </p>
          <button
            type="button"
            style={{ ...ghostBtn, alignSelf: 'flex-start' }}
            disabled={isFetching}
            onClick={() => void refetch()}
          >
            {t('financeCommon.refresh')}
          </button>
        </div>
      </div>
    );
  }

  if (isLoading || !expense) {
    return (
      <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
        <PageHeader
          title={t('financeExpenses.detailTitle')}
          breadcrumbs={[
            { label: t('financeNav.expenses'), to: '/finance/expenses' },
            { label: id ?? '' },
          ]}
        />
        <div style={{ ...cardStyle, padding: 20 }}>
          <SkeletonRow />
        </div>
      </div>
    );
  }

  const currency = (expense.currency || homeCurrency) as Currency;
  const isPending = expense.status === 'pending';
  const isApproved = expense.status === 'approved';
  const title = expense.description || t('financeExpenses.detailTitle');

  return (
    <motion.div
      initial={{ opacity: 0, y: 4 }}
      animate={{ opacity: 1, y: 0 }}
      style={{ display: 'flex', flexDirection: 'column', gap: 20 }}
    >
      <HelmetContainer title={pageTitle(t('financeCommon.pageTitles.expenseDetail'))} />
      <PageHeader
        title={title}
        subtitle={
          <span style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
            <span
              style={{
                fontSize: 11,
                fontWeight: 600,
                padding: '3px 10px',
                borderRadius: 20,
                ...(expenseStatusStyle[expense.status] ?? expenseStatusStyle.pending),
              }}
            >
              {t(`financeExpenses.status.${expense.status}`, { defaultValue: expense.status })}
            </span>
            <span style={{ fontVariantNumeric: 'tabular-nums' }}>
              {formatMoney(expense.amount, currency)}
            </span>
            <span>{formatDate(expense.expenseDate)}</span>
          </span>
        }
        breadcrumbs={[
          { label: t('financeNav.expenses'), to: '/finance/expenses' },
          { label: title },
        ]}
        actions={
          <>
            {!isApproved && (
              <button
                type="button"
                style={{ ...ghostBtn, color: 'var(--finance-danger)' }}
                disabled={deleteMut.isPending}
                onClick={() => setConfirmDelete(true)}
              >
                <Trash2 size={13} /> {t('financeCommon.delete')}
              </button>
            )}
            {isPending && canApproveExpense && (
              <>
                <button
                  type="button"
                  style={{
                    ...ghostBtn,
                    color: 'var(--finance-error)',
                    borderColor: 'color-mix(in srgb, var(--finance-error) 30%, transparent)',
                  }}
                  disabled={rejectMut.isPending || approveMut.isPending}
                  onClick={() => rejectMut.mutate(expense.id)}
                >
                  <XCircle size={13} /> {t('financeExpenses.reject')}
                </button>
                <button
                  type="button"
                  style={primaryBtn}
                  disabled={approveMut.isPending || rejectMut.isPending}
                  onClick={() => approveMut.mutate(expense.id)}
                >
                  <CheckCircle size={13} /> {t('financeExpenses.approve')}
                </button>
              </>
            )}
          </>
        }
      />

      {isPending && (
        <p
          style={{
            margin: 0,
            padding: '10px 14px',
            borderRadius: 10,
            fontSize: 13,
            color: 'var(--hub-text-muted)',
            background: 'var(--finance-warning-soft)',
          }}
        >
          {t('financeExpenses.pendingExplainer')}
        </p>
      )}

      <div
        style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fit, minmax(300px, 1fr))',
          gap: 20,
        }}
      >
        <div style={{ ...cardStyle, padding: 20 }}>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
            <Field label={t('financeCommon.date')} value={formatDate(expense.expenseDate)} />
            <Field label={t('financeCommon.category')}>
              <span style={{ display: 'inline-flex', alignItems: 'center', gap: 6 }}>
                <span
                  style={{
                    width: 8,
                    height: 8,
                    borderRadius: 4,
                    background: category?.color || 'var(--finance-gray)',
                  }}
                />
                {category?.name ?? t('financeExpenses.uncategorized')}
              </span>
            </Field>
            <Field
              label={t('financeCommon.amount')}
              value={formatMoney(expense.amount, currency)}
            />
            {expense.vendor && <Field label={t('financeCommon.vendor')} value={expense.vendor} />}
            {expense.paymentMethod && (
              <Field
                label={t('financeExpenses.paymentMethod')}
                value={t(`financeExpenses.paymentMethods.${expense.paymentMethod}`, {
                  defaultValue: expense.paymentMethod,
                })}
              />
            )}
            {expense.notes && (
              <div
                style={{
                  marginTop: 10,
                  padding: 12,
                  background: 'var(--hub-hover)',
                  borderRadius: 8,
                  fontSize: 12,
                  color: 'var(--hub-text-muted)',
                  whiteSpace: 'pre-wrap',
                }}
              >
                {expense.notes}
              </div>
            )}
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
            {t('financeCommon.receipt')}
          </div>
          {isApproved ? (
            // Approved expenses are locked, so the receipt can be viewed but not changed.
            receipt ? (
              <a
                href={receipt}
                target="_blank"
                rel="noreferrer"
                style={{
                  display: 'inline-flex',
                  alignItems: 'center',
                  gap: 6,
                  fontSize: 13,
                  color: 'var(--finance-accent)',
                }}
              >
                <ExternalLink size={13} /> {t('financeExpenses.viewReceipt')}
              </a>
            ) : (
              <p style={{ margin: 0, fontSize: 13, color: 'var(--hub-text-muted)' }}>
                {t('financeExpenses.noReceipt')}
              </p>
            )
          ) : (
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
                  {t('financeCommon.saving')}
                </div>
              )}
            </div>
          )}
        </div>
      </div>

      <ConfirmDialog
        open={confirmDelete}
        onClose={() => setConfirmDelete(false)}
        onConfirm={() =>
          deleteMut.mutate(expense.id, {
            onSuccess: () => {
              toast.success(t('financeExpenses.deleted'));
              void navigate({ to: '/finance/expenses' });
            },
          })
        }
        title={t('financeExpenses.deleteTitle')}
        message={t('financeExpenses.deleteConfirm', { description: title })}
        confirmLabel={t('financeCommon.delete')}
        destructive
      />
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
