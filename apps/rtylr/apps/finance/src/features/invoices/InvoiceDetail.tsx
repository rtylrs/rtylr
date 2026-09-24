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
import { useInvoiceDetail, useUpdateInvoiceStatus } from './useInvoices';
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

interface VoidInvoiceDialogProps {
  invoiceId: string;
  invoiceNumber: string;
  onClose: () => void;
  onVoided: () => void;
}

const VoidInvoiceDialog: React.FC<VoidInvoiceDialogProps> = ({
  invoiceId,
  invoiceNumber,
  onClose,
  onVoided,
}) => {
  const { t } = useTranslation();
  const updateStatus = useUpdateInvoiceStatus();

  return (
    <ConfirmDialog
      open
      onClose={onClose}
      onConfirm={() =>
        updateStatus.mutate(
          { id: invoiceId, status: 'Cancelled' },
          {
            onSuccess: () => {
              toast.success(`Invoice ${invoiceNumber} voided`);
              onVoided();
            },
            onError: (error) =>
              toast.error(getOneKitErrorMessage(error, `Failed to void invoice ${invoiceNumber}`)),
          }
        )
      }
      title={t('invoices.detail.void')}
      message={`Are you sure you want to void ${invoiceNumber}? This cannot be undone.`}
      confirmLabel={updateStatus.isPending ? 'Voiding...' : 'Void'}
      destructive
    />
  );
};

const InvoiceDetail: React.FC = () => {
  const { t } = useTranslation();
  const { id } = routeApi.useParams();
  const navigate = useNavigate();
  const { data: invoice, isLoading, isError, error, isFetching, refetch } = useInvoiceDetail(id);
  const [showSend, setShowSend] = useState(false);
  const [showPayment, setShowPayment] = useState(false);
  const [showVoid, setShowVoid] = useState(false);
  const role = useUserStore((state) => state.user?.role ?? state.role);
  const canSendInvoice = canFinancePermission(role, 'finance:invoices:send');

  if (isError && !invoice) {
    return (
      <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
        <PageHeader
          title={t('invoices.detail.title')}
          breadcrumbs={[{ label: t('nav.invoices'), to: '/invoices' }, { label: id ?? '' }]}
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

  if (isLoading || !invoice) {
    return (
      <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
        <PageHeader
          title={t('invoices.detail.title')}
          breadcrumbs={[{ label: t('nav.invoices'), to: '/invoices' }, { label: id ?? '' }]}
        />
        <div style={{ ...cardStyle, padding: 20 }}>
          <SkeletonRow count={4} />
        </div>
      </div>
    );
  }

  const paymentStatus =
    invoice.status === 'Paid' ? 'paid' : invoice.status === 'Overdue' ? 'overdue' : 'unpaid';
  const balanceDue = Math.max(0, invoice.amount - (invoice.paid ?? 0));
  const invoiceNumber = invoice.number?.trim() || invoice.id;

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
        title={`Invoice ${invoiceNumber}`}
        subtitle={
          <span style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <PaymentBadge status={paymentStatus} /> <MoneyDisplay amount={invoice.amount} /> total
          </span>
        }
        breadcrumbs={[{ label: t('nav.invoices'), to: '/invoices' }, { label: invoiceNumber }]}
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
                toast.success(t('invoices.detail.cloning'));
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
            {canSendInvoice && invoice.status !== 'Paid' && invoice.status !== 'Cancelled' && (
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
          invoiceNumber={invoiceNumber}
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
          discountType={invoice.discountType}
          discountAmount={invoice.discountAmount}
          subtotal={invoice.subtotal}
          tax={invoice.tax}
          total={invoice.amount}
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
              {t('invoices.detail.summary')}
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
              <span>{t('invoices.detail.totalInvoiced')}</span>
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
              <span>{t('invoices.detail.paid')}</span>
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
              <span>{t('invoices.detail.balanceDue')}</span>
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
              {t('invoices.detail.history')}
            </div>
            {invoice.history?.map((h, i) => (
              <div
                key={h.id ?? `${h.at}-${i}`}
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
              {t('invoices.detail.client')}
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
      {showVoid && (
        <VoidInvoiceDialog
          invoiceId={invoice.id}
          invoiceNumber={invoiceNumber}
          onClose={() => setShowVoid(false)}
          onVoided={() => {
            setShowVoid(false);
            navigate({ to: '/invoices' });
          }}
        />
      )}
    </motion.div>
  );
};

export default InvoiceDetail;
