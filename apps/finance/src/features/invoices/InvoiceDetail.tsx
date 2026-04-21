import React, { useMemo, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { motion } from 'motion/react';
import { Send, Copy, XCircle, CreditCard, Printer } from 'lucide-react';
import { toast } from 'sonner';
import type { InvoiceItem } from '@voxire/contracts/finance/v1/invoices';
import { useGetInvoice } from '../../shared/hooks/queries/useInvoices';
import { minorToMajor } from '../../shared/utils/money';
import type { Currency } from '../../shared/utils/money';
import PageHeader from '../../shared/components/PageHeader';
import MoneyDisplay from '../../shared/components/MoneyDisplay';
import PaymentBadge from '../../shared/components/PaymentBadge';
import DownloadPdfButton from '../../shared/components/DownloadPdfButton';
import InvoicePreview from '../../shared/components/InvoicePreview';
import SendInvoiceModal from '../../shared/components/SendInvoiceModal';
import RecordPaymentModal from '../../shared/components/RecordPaymentModal';
import ConfirmDialog from '../../shared/components/ConfirmDialog';
import { SkeletonRow } from '../../shared/components/Skeleton';
import { ACCENT, cardStyle, ghostBtn, primaryBtn } from '../../shared/lib/tokens';

type PaymentStatus = 'paid' | 'overdue' | 'unpaid';

function toPaymentStatus(raw: string): PaymentStatus {
  const lower = raw?.toLowerCase?.() ?? '';
  if (lower === 'paid') return 'paid';
  if (lower === 'overdue') return 'overdue';
  return 'unpaid';
}

const InvoiceDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const numericId = id ? Number(id) : undefined;
  const { data, isLoading } = useGetInvoice(numericId);
  const [showSend, setShowSend] = useState(false);
  const [showPayment, setShowPayment] = useState(false);
  const [showVoid, setShowVoid] = useState(false);

  const view = useMemo(() => {
    const inv = data?.invoice;
    if (!inv) return null;
    const currency = (inv.currency || 'USD') as Currency;
    const total = minorToMajor(inv.total, currency);
    const paid = minorToMajor(inv.amountPaid, currency);
    const due = minorToMajor(inv.amountDue, currency);
    const discount = minorToMajor(inv.discountAmount, currency);
    const items: InvoiceItem[] = inv.items ?? [];
    return {
      id: String(inv.id),
      number: inv.invoiceNumber,
      clientName: inv.clientName,
      clientEmail: inv.clientEmail,
      status: inv.status,
      issueDate: inv.issueDate,
      dueDate: inv.dueDate,
      currency,
      total,
      paid,
      due,
      discount,
      notes: inv.notes,
      items,
      paymentStatus: toPaymentStatus(inv.status),
    };
  }, [data]);

  if (isLoading || !view) {
    return (
      <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
        <PageHeader
          title="Invoice"
          breadcrumbs={[{ label: 'Invoices', to: '/invoices' }, { label: id ?? '' }]}
        />
        <div style={{ ...cardStyle, padding: 20 }}>
          <SkeletonRow count={4} />
        </div>
      </div>
    );
  }

  const clientObj = { id: 'c1', name: view.clientName, email: view.clientEmail };

  return (
    <motion.div
      initial={{ opacity: 0, y: 4 }}
      animate={{ opacity: 1, y: 0 }}
      style={{ display: 'flex', flexDirection: 'column', gap: 20 }}
    >
      <PageHeader
        title={`Invoice ${view.number || view.id}`}
        subtitle={
          <span style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <PaymentBadge status={view.paymentStatus} />{' '}
            <MoneyDisplay amount={view.total} currency={view.currency} /> total
          </span>
        }
        breadcrumbs={[{ label: 'Invoices', to: '/invoices' }, { label: view.number || view.id }]}
        actions={
          <>
            <button style={ghostBtn} onClick={() => window.print()}>
              <Printer size={13} /> Print
            </button>
            <DownloadPdfButton invoiceId={view.id} />
            <button
              style={ghostBtn}
              onClick={() => {
                navigate(`/invoices/new?clone=${view.id}`);
                toast.success('Cloning invoice');
              }}
            >
              <Copy size={13} /> Clone
            </button>
            {view.paymentStatus !== 'paid' && (
              <button style={ghostBtn} onClick={() => setShowPayment(true)}>
                <CreditCard size={13} /> Record Payment
              </button>
            )}
            <button style={primaryBtn} onClick={() => setShowSend(true)}>
              <Send size={13} /> Send
            </button>
            {view.paymentStatus !== 'paid' && (
              <button
                style={{ ...ghostBtn, color: '#ef4444', borderColor: 'rgba(239,68,68,0.3)' }}
                onClick={() => setShowVoid(true)}
              >
                <XCircle size={13} /> Void
              </button>
            )}
          </>
        }
      />

      <div style={{ display: 'grid', gridTemplateColumns: '2fr 1fr', gap: 20 }}>
        <InvoicePreview
          invoiceNumber={view.number || view.id}
          client={clientObj}
          issuedDate={view.issueDate}
          dueDate={view.dueDate}
          items={view.items.map((item, idx) => ({
            id: `line-${idx}`,
            description: item.description,
            quantity: item.quantity,
            unitPrice: minorToMajor(item.unitPrice, view.currency),
            taxRate: item.taxRate,
          }))}
          notes={view.notes}
          discount={view.discount}
        />

        <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
          <div style={{ ...cardStyle, padding: 16 }}>
            <div
              style={{
                fontSize: 11,
                color: 'var(--hub-text-dim)',
                textTransform: 'uppercase',
                letterSpacing: '0.08em',
                marginBottom: 10,
              }}
            >
              Summary
            </div>
            <div
              style={{
                display: 'flex',
                justifyContent: 'space-between',
                fontSize: 12,
                padding: '6px 0',
                color: 'var(--hub-text-muted)',
              }}
            >
              <span>Total invoiced</span>
              <MoneyDisplay amount={view.total} currency={view.currency} />
            </div>
            <div
              style={{
                display: 'flex',
                justifyContent: 'space-between',
                fontSize: 12,
                padding: '6px 0',
                color: 'var(--hub-text-muted)',
              }}
            >
              <span>Paid</span>
              <MoneyDisplay amount={view.paid} currency={view.currency} />
            </div>
            <div
              style={{
                display: 'flex',
                justifyContent: 'space-between',
                padding: '10px 0 0',
                borderTop: '1px solid var(--hub-border)',
                fontSize: 14,
                fontWeight: 700,
                color: view.due > 0 ? ACCENT : 'var(--hub-text)',
              }}
            >
              <span>Balance due</span>
              <MoneyDisplay amount={view.due} currency={view.currency} />
            </div>
          </div>

          <div style={{ ...cardStyle, padding: 16 }}>
            <div
              style={{
                fontSize: 11,
                color: 'var(--hub-text-dim)',
                textTransform: 'uppercase',
                letterSpacing: '0.08em',
                marginBottom: 10,
              }}
            >
              Payments
            </div>
            {(data?.payments ?? []).length === 0 ? (
              <div style={{ fontSize: 12, color: 'var(--hub-text-dim)' }}>No payments recorded</div>
            ) : (
              (data?.payments ?? []).map((p) => (
                <div
                  key={p.id}
                  style={{ display: 'flex', gap: 10, padding: '6px 0', alignItems: 'flex-start' }}
                >
                  <div
                    style={{
                      width: 6,
                      height: 6,
                      borderRadius: 3,
                      background: ACCENT,
                      marginTop: 5,
                    }}
                  />
                  <div>
                    <div style={{ fontSize: 12, color: 'var(--hub-text)' }}>
                      <MoneyDisplay
                        amount={minorToMajor(p.amount, (p.currency || view.currency) as Currency)}
                        currency={(p.currency || view.currency) as Currency}
                      />{' '}
                      {p.method && <span>· {p.method}</span>}
                    </div>
                    <div style={{ fontSize: 10, color: 'var(--hub-text-dim)' }}>
                      {p.paymentDate}
                    </div>
                  </div>
                </div>
              ))
            )}
          </div>

          <div style={{ ...cardStyle, padding: 16 }}>
            <div
              style={{
                fontSize: 11,
                color: 'var(--hub-text-dim)',
                textTransform: 'uppercase',
                letterSpacing: '0.08em',
                marginBottom: 10,
              }}
            >
              Client
            </div>
            <div style={{ fontSize: 13, fontWeight: 600, color: 'var(--hub-text)' }}>
              {view.clientName}
            </div>
            <div style={{ fontSize: 12, color: 'var(--hub-text-muted)' }}>{view.clientEmail}</div>
          </div>
        </div>
      </div>

      <SendInvoiceModal
        open={showSend}
        onClose={() => setShowSend(false)}
        invoiceId={view.id}
        defaultTo={view.clientEmail}
      />
      <RecordPaymentModal
        open={showPayment}
        onClose={() => setShowPayment(false)}
        invoiceId={view.id}
        amount={view.due}
      />
      <ConfirmDialog
        open={showVoid}
        onClose={() => setShowVoid(false)}
        onConfirm={() => toast.success(`Invoice ${view.number || view.id} voided`)}
        title="Void invoice"
        message={`Are you sure you want to void ${view.number || view.id}? This cannot be undone.`}
        confirmLabel="Void"
        destructive
      />
    </motion.div>
  );
};

export default InvoiceDetail;
