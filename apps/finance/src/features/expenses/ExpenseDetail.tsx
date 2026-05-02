import React from 'react';
import { useParams } from 'react-router-dom';
import { motion } from 'motion/react';
import { CheckCircle, XCircle } from 'lucide-react';
import PageHeader from '../../shared/components/PageHeader';
import MoneyDisplay from '../../shared/components/MoneyDisplay';
import { SkeletonRow } from '../../shared/components/Skeleton';
import ReceiptUpload from '../../shared/components/ReceiptUpload';
import { useExpenseById } from './useExpenseById';
import { useApproveExpense, useRejectExpense } from '../../shared/hooks/queries/useExpenses';
import { cardStyle, ghostBtn, primaryBtn } from '../../shared/lib/tokens';

const approvalStyle: Record<string, React.CSSProperties> = {
  pending: {
    background: 'color-mix(in srgb, var(--finance-warning, #f59e0b) 12%, transparent)',
    color: 'var(--finance-warning, #f59e0b)',
  },
  approved: {
    background: 'color-mix(in srgb, var(--finance-success, #10b981) 12%, transparent)',
    color: 'var(--finance-success, #10b981)',
  },
  rejected: {
    background: 'color-mix(in srgb, var(--finance-error, #ef4444) 12%, transparent)',
    color: 'var(--finance-error, #ef4444)',
  },
};

const ExpenseDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const { data: expense, isLoading } = useExpenseById(id);
  const approveMut = useApproveExpense();
  const rejectMut = useRejectExpense();
  const [receipt, setReceipt] = React.useState<string | null>(null);

  React.useEffect(() => {
    if (expense?.receiptUrl) setReceipt(expense.receiptUrl);
  }, [expense]);

  if (isLoading || !expense) {
    return (
      <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
        <PageHeader
          title="Expense"
          breadcrumbs={[{ label: 'Expenses', to: '/expenses' }, { label: id ?? '' }]}
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
                ...approvalStyle[expense.approval],
              }}
            >
              {expense.approval}
            </span>
            <MoneyDisplay amount={expense.amount} />
          </span>
        }
        breadcrumbs={[{ label: 'Expenses', to: '/expenses' }, { label: expense.id }]}
        actions={
          <>
            <button
              style={{
                ...ghostBtn,
                color: 'var(--finance-error, #ef4444)',
                borderColor: 'color-mix(in srgb, var(--finance-error, #ef4444) 30%, transparent)',
              }}
              disabled={rejectMut.isPending}
              onClick={() => rejectMut.mutate(expense.id)}
            >
              <XCircle size={13} /> Reject
            </button>
            <button
              style={primaryBtn}
              disabled={approveMut.isPending}
              onClick={() => approveMut.mutate(expense.id)}
            >
              <CheckCircle size={13} /> Approve
            </button>
          </>
        }
      />

      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 20 }}>
        <div style={{ ...cardStyle, padding: 20 }}>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
            <Field label="Date" value={expense.date} />
            <Field label="Description" value={expense.description} />
            <Field label="Category" value={expense.category} />
            <Field label="Amount">
              <MoneyDisplay amount={expense.amount} />
            </Field>
            <Field label="Submitter" value={expense.submitter} />
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
            Receipt
          </div>
          <ReceiptUpload value={receipt} onChange={setReceipt} />
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
