import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { pageTitle } from '../../shared/lib/pageTitle';
import { HelmetContainer } from '@voxire/components';
import { getRouteApi, useNavigate } from '@tanstack/react-router';
import { useUserStore } from '@voxire/auth';

const routeApi = getRouteApi('/layout/invoices/$id');
import { motion } from 'motion/react';
import { Send, Copy, XCircle, CreditCard, Printer } from 'lucide-react';
import { toast } from 'sonner';
import { getOneKitErrorMessage } from '@voxire/api';
import { useInvoiceDetail, useDeleteInvoice } from './useInvoices';
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
import { canFinancePermission } from '../../shared/lib/financePermissions';

const InvoiceDetail: React.FC = () => {
  const { t } = useTranslation();
  const { id } = routeApi.useParams();
  const navigate = useNavigate();
  const { data: invoice, isLoading } = useInvoiceDetail(id);
  const deleteInvoice = useDeleteInvoice();
  const [showSend, setShowSend] = useState(false);
  const [showPayment, setShowPayment] = useState(false);
  const [showVoid, setShowVoid] = useState(false);
  const role = useUserStore((state) => state.user?.role ?? state.role);
  const canSendInvoice = canFinancePermission(role, 'finance:invoices:send');

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
  const balanceDue = invoice.amount - (invoice.paid ?? 0);

  // clientId is not available in the current Invoice payload; using empty string until the API provides it.
  const clientObj = { id: invoice.id, name: invoice.client, email: invoice.clientEmail };

  return (
    <motion.div
      initial={{ opacity: 0, y: 4 }}
      animate={{ opacity: 1, y: 0 }}
      style={{ display: 'flex', flexDirection: 'column', gap: 20 }}
    >
      <HelmetContainer title={pageTitle(t('common.pageTitles.invoiceDetail'))} />
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
            <button type="button" style={ghostBtn} onClick={() => window.print()}>
              <Printer size={13} /> Print
            </button>
            <DownloadPdfButton invoiceId={invoice.id} />
            <button
              type="button"
              style={ghostBtn}
              onClick={() => {
                navigate({ to: `/invoices/new?clone=${invoice.id}` });
                toast.success('Cloning invoice');
              }}
            >
              <Copy size={13} /> Clone
            </button>
            {invoice.status !== 'Paid' && (
              <button type="button" style={ghostBtn} onClick={() => setShowPayment(true)}>
                <CreditCard size={13} /> Record Payment
              </button>
            )}
            {canSendInvoice && (
              <button type="button" style={primaryBtn} onClick={() => setShowSend(true)}>
                <Send size={13} /> Send
              </button>
            )}
            {canSendInvoice && invoice.status !== 'Paid' && (
              <button
                type="button"
                style={{
                  ...ghostBtn,
                  color: 'var(--finance-danger)',
                  borderColor: 'var(--finance-danger-border)',
                }}
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
          items={(invoice.items ?? []).map((it, idx) => ({
            id: String(idx),
            description: it.description ?? '',
            quantity: it.quantity ?? 0,
            unitPrice: it.unitPrice ?? 0,
            taxRate: it.taxRate ?? 0,
          }))}
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
            {invoice.history?.map((h, i) => (
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

      <SendInvoiceModal open={showSend} onClose={() => setShowSend(false)} invoiceId={invoice.id} />
      <RecordPaymentModal
        open={showPayment}
        onClose={() => setShowPayment(false)}
        invoiceId={invoice.id}
        amount={balanceDue}
        currency={invoice.currency}
      />
      <ConfirmDialog
        open={showVoid}
        onClose={() => setShowVoid(false)}
        onConfirm={() =>
          deleteInvoice.mutate(invoice.id, {
            onSuccess: () => {
              toast.success(`Invoice ${invoice.id} voided`);
              setShowVoid(false);
              navigate({ to: '/invoices' });
            },
            onError: (error) =>
              toast.error(getOneKitErrorMessage(error, `Failed to void invoice ${invoice.id}`)),
          })
        }
        title="Void invoice"
        message={`Are you sure you want to void ${invoice.id}? This cannot be undone.`}
        confirmLabel={deleteInvoice.isPending ? 'Voiding...' : 'Void'}
        destructive
      />
    </motion.div>
  );
};

export default InvoiceDetail;
