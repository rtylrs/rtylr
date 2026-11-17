import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { pageTitle } from '../../shared/lib/pageTitle';
import { HelmetContainer } from '@voxire/components';
import { getRouteApi, useNavigate } from '@tanstack/react-router';

const routeApi = getRouteApi('/business/finance/invoices/$id');
import { motion } from 'motion/react';
import { Send, Copy, XCircle, CreditCard, Printer, FolderKanban } from 'lucide-react';
import { toast } from 'sonner';
import { getOneKitErrorMessage } from '@voxire/api';
import { useInvoiceDetail, useSetInvoiceProject, useUpdateInvoiceStatus } from './useInvoices';
import PageHeader from '../../shared/components/PageHeader';
import MoneyDisplay from '../../shared/components/MoneyDisplay';
import PaymentBadge from '../../shared/components/PaymentBadge';
import DownloadPdfButton from '../../shared/components/DownloadPdfButton';
import InvoicePreview from '../../shared/components/InvoicePreview';
import SendInvoiceModal from '../../shared/components/SendInvoiceModal';
import RecordPaymentModal from '../../shared/components/RecordPaymentModal';
import ConfirmDialog from '../../shared/components/ConfirmDialog';
import { SkeletonRow } from '../../shared/components/Skeleton';
import {
  ACCENT,
  cardStyle,
  ghostBtn,
  inputStyle,
  labelStyle,
  primaryBtn,
} from '../../shared/lib/tokens';
import { canFinancePermission } from '../../shared/lib/financePermissions';
import { useWorkProjects, type WorkProject } from '../../shared/hooks/queries/useWorkProjects';

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
  const { t } = useTranslation('financeCommon');
  const updateStatus = useUpdateInvoiceStatus();

  return (
    <ConfirmDialog
      open
      closeOnConfirm={false}
      pending={updateStatus.isPending}
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
      title={t('financeInvoices.detail.void')}
      message={`Are you sure you want to void ${invoiceNumber}? This cannot be undone.`}
      confirmLabel={updateStatus.isPending ? 'Voiding...' : 'Void'}
      destructive
    />
  );
};

const cardHeadingStyle: React.CSSProperties = {
  fontSize: 11,
  color: 'var(--hub-text-dim)',
  textTransform: 'uppercase',
  letterSpacing: '0.08em',
  marginBottom: 10,
};

interface InvoiceProjectCardProps {
  invoiceId: string;
  /** Current Work project id; '' when the invoice is untagged. */
  projectId: string;
  projectName: string | null;
  workProjects: WorkProject[] | null | undefined;
  canChange: boolean;
}

// The Work project an invoice bills. Unlike the rest of the document the tag
// can change after the invoice is sent or paid, so a project's dashboard can
// count invoices that were issued before the project was picked.
const InvoiceProjectCard: React.FC<InvoiceProjectCardProps> = ({
  invoiceId,
  projectId,
  projectName,
  workProjects,
  canChange,
}) => {
  const { t } = useTranslation('financeCommon');
  const setProject = useSetInvoiceProject();
  const [editing, setEditing] = useState(false);
  const selectId = `invoice-project-${invoiceId}`;
  const projects = workProjects ?? [];
  // Keep the current tag selectable when the picker does not list it
  // (archived, or a project this viewer cannot see) so the select never
  // misreports it.
  const listsCurrent = projectId === '' || projects.some((project) => project.id === projectId);

  const choose = (next: string) => {
    if (next === projectId) {
      setEditing(false);
      return;
    }
    setProject.mutate(
      { id: invoiceId, projectId: next === '' ? null : next },
      {
        onSuccess: () => {
          setEditing(false);
          toast.success(t('financeInvoices.detail.projectChanged'));
        },
      }
    );
  };

  return (
    <div style={{ ...cardStyle, padding: 16 }}>
      <div style={cardHeadingStyle}>{t('financeInvoices.detail.project')}</div>
      {editing ? (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
          <label htmlFor={selectId} style={labelStyle}>
            {t('financeInvoices.detail.changeProject')}
          </label>
          <select
            id={selectId}
            value={projectId}
            disabled={setProject.isPending}
            onChange={(event) => choose(event.target.value)}
            style={inputStyle}
          >
            <option value="">{t('financeInvoices.new.noProject')}</option>
            {projects.map((project) => (
              <option key={project.id} value={project.id}>
                {project.name}
              </option>
            ))}
            {!listsCurrent && <option value={projectId}>{projectName ?? projectId}</option>}
          </select>
          {projects.length === 0 && (
            <div style={{ fontSize: 11, color: 'var(--hub-text-dim)' }}>
              {t('financeInvoices.new.noProjectsAvailable')}
            </div>
          )}
          <div style={{ display: 'flex', justifyContent: 'flex-end' }}>
            <button
              type="button"
              style={ghostBtn}
              disabled={setProject.isPending}
              onClick={() => setEditing(false)}
            >
              {t('financeCommon.cancel')}
            </button>
          </div>
        </div>
      ) : (
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            gap: 8,
            flexWrap: 'wrap',
          }}
        >
          <div
            style={{
              fontSize: 13,
              fontWeight: projectName ? 600 : 400,
              color: projectName ? 'var(--hub-text)' : 'var(--hub-text-muted)',
              minWidth: 0,
              overflowWrap: 'anywhere',
            }}
          >
            {projectName ?? t('financeInvoices.new.noProject')}
          </div>
          {canChange && (
            <button type="button" style={ghostBtn} onClick={() => setEditing(true)}>
              <FolderKanban size={13} /> {t('financeInvoices.detail.changeProject')}
            </button>
          )}
        </div>
      )}
    </div>
  );
};

const InvoiceDetail: React.FC = () => {
  const { t } = useTranslation('financeCommon');
  const { id } = routeApi.useParams();
  const navigate = useNavigate();
  const { data: invoice, isLoading, isError, error, isFetching, refetch } = useInvoiceDetail(id);
  const { data: workProjects } = useWorkProjects();
  const [showSend, setShowSend] = useState(false);
  const [showPayment, setShowPayment] = useState(false);
  const [showVoid, setShowVoid] = useState(false);
  const canSendInvoice = canFinancePermission('finance:invoices:send');
  // Same permission as every other invoice write on the server.
  const canWriteInvoice = canFinancePermission('finance:invoices:create');

  if (isError && !invoice) {
    return (
      <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
        <PageHeader
          title={t('financeInvoices.detail.title')}
          breadcrumbs={[
            { label: t('financeNav.invoices'), to: '/finance/invoices' },
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
            style={ghostBtn}
            disabled={isFetching}
            onClick={() => void refetch()}
          >
            {t('financeCommon.refresh')}
          </button>
        </div>
      </div>
    );
  }

  if (isLoading || !invoice) {
    return (
      <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
        <PageHeader
          title={t('financeInvoices.detail.title')}
          breadcrumbs={[
            { label: t('financeNav.invoices'), to: '/finance/invoices' },
            { label: id ?? '' },
          ]}
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
  // The picker only lists active projects the caller can see, so an archived
  // or hidden project falls back to its id rather than disappearing.
  const projectName = invoice.projectId
    ? (workProjects?.find((project) => project.id === invoice.projectId)?.name ??
      t('financeInvoices.new.projectFallback', { id: invoice.projectId }))
    : null;
  // Cancelled covers void too (both fold into Cancelled); the server freezes
  // the tag on both, so the control is not offered there.
  const canChangeProject = canWriteInvoice && invoice.status !== 'Cancelled';

  return (
    <motion.div
      initial={{ opacity: 0, y: 4 }}
      animate={{ opacity: 1, y: 0 }}
      style={{ display: 'flex', flexDirection: 'column', gap: 20 }}
    >
      <HelmetContainer title={pageTitle(t('financeCommon.pageTitles.invoiceDetail'))} />
      <PageHeader
        title={`Invoice ${invoiceNumber}`}
        subtitle={
          <span style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <PaymentBadge status={paymentStatus} /> <MoneyDisplay amount={invoice.amount} /> total
          </span>
        }
        breadcrumbs={[
          { label: t('financeNav.invoices'), to: '/finance/invoices' },
          { label: invoiceNumber },
        ]}
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
                navigate({ to: '/finance/invoices/new', search: { clone: invoice.id } });
                toast.success(t('financeInvoices.detail.cloning'));
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
              {t('financeInvoices.detail.summary')}
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
              <span>{t('financeInvoices.detail.totalInvoiced')}</span>
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
              <span>{t('financeInvoices.detail.paid')}</span>
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
              <span>{t('financeInvoices.detail.balanceDue')}</span>
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
              {t('financeInvoices.detail.history')}
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
              {t('financeInvoices.detail.client')}
            </div>
            <div style={{ fontSize: 13, fontWeight: 600, color: 'var(--hub-text)' }}>
              {invoice.client}
            </div>
            <div style={{ fontSize: 12, color: 'var(--hub-text-muted)' }}>
              {invoice.clientEmail}
            </div>
          </div>

          {(projectName || canChangeProject) && (
            <InvoiceProjectCard
              invoiceId={invoice.id}
              projectId={invoice.projectId ?? ''}
              projectName={projectName}
              workProjects={workProjects}
              canChange={canChangeProject}
            />
          )}
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
            navigate({ to: '/finance/invoices' });
          }}
        />
      )}
    </motion.div>
  );
};

export default InvoiceDetail;
