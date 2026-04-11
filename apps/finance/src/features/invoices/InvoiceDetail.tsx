import React, { useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { motion } from 'motion/react';
import { Send, Copy, XCircle, CreditCard, Printer, FileText } from 'lucide-react';
import { toast } from 'sonner';
import { useInvoice } from './useInvoice';
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

const InvoiceDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { data: invoice, isLoading } = useInvoice(id);
  const [showSend, setShowSend] = useState(false);
  const [showPayment, setShowPayment] = useState(false);
  const [showVoid, setShowVoid] = useState(false);

  if (isLoading || !invoice) {
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

  const paymentStatus =
    invoice.status === 'Paid' ? 'paid' : invoice.status === 'Overdue' ? 'overdue' : 'unpaid';
  const balanceDue = invoice.amount - invoice.paid;

  const clientObj = { id: 'c1', name: invoice.client, email: invoice.clientEmail };

  return (
    <motion.div
      initial={{ opacity: 0, y: 4 }}
      animate={{ opacity: 1, y: 0 }}
      style={{ display: 'flex', flexDirection: 'column', gap: 20 }}
    >
      <PageHeader
        title={`Invoice ${invoice.id}`}
        subtitle={
          <span style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <PaymentBadge status={paymentStatus} /> <MoneyDisplay amount={invoice.amount} /> total
          </span>
        }
        breadcrumbs={[{ label: 'Invoices', to: '/invoices' }, { label: invoice.id }]}
        actions={
          <>
            <button style={ghostBtn} onClick={() => window.print()}>
              <Printer size={13} /> Print
            </button>
            <DownloadPdfButton invoiceId={invoice.id} />
            <button
              style={ghostBtn}
              onClick={() => {
                navigate(`/invoices/new?clone=${invoice.id}`);
                toast.success('Cloning invoice');
              }}
            >
              <Copy size={13} /> Clone
            </button>
            {invoice.status !== 'Paid' && (
              <button style={ghostBtn} onClick={() => setShowPayment(true)}>
                <CreditCard size={13} /> Record Payment
              </button>
            )}
            <button style={primaryBtn} onClick={() => setShowSend(true)}>
              <Send size={13} /> Send
            </button>
            {invoice.status !== 'Paid' && (
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
          invoiceNumber={invoice.id}
          client={clientObj}
          issuedDate={invoice.issuedDate}
          dueDate={invoice.dueDate}
          items={invoice.items}
          notes={invoice.notes}
          discount={invoice.discount}
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
              <MoneyDisplay amount={invoice.amount} />
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
              <MoneyDisplay amount={invoice.paid} />
            </div>
            <div
              style={{
                display: 'flex',
                justifyContent: 'space-between',
                padding: '10px 0 0',
                borderTop: '1px solid var(--hub-border)',
                fontSize: 14,
                fontWeight: 700,
                color: balanceDue > 0 ? ACCENT : 'var(--hub-text)',
              }}
            >
              <span>Balance due</span>
              <MoneyDisplay amount={balanceDue} />
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
              History
            </div>
            {invoice.history.map((h, i) => (
              <div
                key={i}
                style={{ display: 'flex', gap: 10, padding: '6px 0', alignItems: 'flex-start' }}
              >
                <div
                  style={{ width: 6, height: 6, borderRadius: 3, background: ACCENT, marginTop: 5 }}
                />
                <div>
                  <div style={{ fontSize: 12, color: 'var(--hub-text)' }}>{h.label}</div>
                  <div style={{ fontSize: 10, color: 'var(--hub-text-dim)' }}>{h.at}</div>
                </div>
              </div>
            ))}
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
              {invoice.client}
            </div>
            <div style={{ fontSize: 12, color: 'var(--hub-text-muted)' }}>
              {invoice.clientEmail}
            </div>
          </div>
        </div>
      </div>

      <SendInvoiceModal
        open={showSend}
        onClose={() => setShowSend(false)}
        invoiceId={invoice.id}
        defaultTo={invoice.clientEmail}
      />
      <RecordPaymentModal
        open={showPayment}
        onClose={() => setShowPayment(false)}
        invoiceId={invoice.id}
        amount={balanceDue}
      />
      <ConfirmDialog
        open={showVoid}
        onClose={() => setShowVoid(false)}
        onConfirm={() => toast.success(`Invoice ${invoice.id} voided`)}
        title="Void invoice"
        message={`Are you sure you want to void ${invoice.id}? This cannot be undone.`}
        confirmLabel="Void"
        destructive
      />
    </motion.div>
  );
};

export default InvoiceDetail;
