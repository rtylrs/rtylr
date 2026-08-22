import React, { useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { Link, useNavigate } from '@tanstack/react-router';
import { useQueryClient } from '@tanstack/react-query';
import { motion } from 'motion/react';
import { Plus, Send, CheckCircle, Eye, FileText, Download, FileSpreadsheet } from 'lucide-react';
import { toast } from 'sonner';
import { getOneKitErrorMessage } from '@voxire/api';
import { useTranslation } from 'react-i18next';
import { EmptyState } from '@voxire/ui';
import { Invoice } from '../../shared/services/financeApi';
import {
  useInvoices,
  useUpdateInvoiceStatus,
  sendInvoicesByIds,
  fetchAllInvoicesForExport,
  MAX_EXPORT_PAGES,
} from './useInvoices';
import PageHeader from '../../shared/components/PageHeader';
import MoneyDisplay from '../../shared/components/MoneyDisplay';
import { SkeletonRow } from '../../shared/components/Skeleton';
import InvoiceStatusFilter, {
  InvoiceFilterValue,
} from '../../shared/components/InvoiceStatusFilter';
import DownloadPdfButton from '../../shared/components/DownloadPdfButton';
import SendInvoiceModal from '../../shared/components/SendInvoiceModal';
import RecordPaymentModal from '../../shared/components/RecordPaymentModal';
import ConfirmDialog from '../../shared/components/ConfirmDialog';
import {
  exportQuickbooksInvoices,
  exportXeroInvoices,
  InvoiceExportOptions,
} from '../../shared/lib/exports';
import { cardStyle, ghostBtn, primaryBtn, ACCENT } from '../../shared/lib/tokens';

const statusStyle: Record<Invoice['status'], React.CSSProperties> = {
  Draft: { background: 'var(--finance-neutral-soft)', color: 'var(--finance-text-dim)' },
  Sent: { background: 'var(--finance-info-soft)', color: 'var(--finance-info)' },
  Paid: { background: 'var(--finance-accent-soft)', color: 'var(--finance-accent)' },
  Overdue: { background: 'var(--finance-danger-soft)', color: 'var(--finance-danger)' },
  Cancelled: { background: 'var(--finance-neutral-soft)', color: 'var(--finance-text-dim)' },
};

const Invoices: React.FC = () => {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const [activeTab, setActiveTab] = useState<InvoiceFilterValue>('All');
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [showBulkSend, setShowBulkSend] = useState(false);
  const [bulkSending, setBulkSending] = useState(false);
  const [sendTarget, setSendTarget] = useState<Invoice | null>(null);
  const [payTarget, setPayTarget] = useState<Invoice | null>(null);
  const [exporting, setExporting] = useState(false);
  const [page, setPage] = useState(1);
  const { data, isLoading: loading } = useInvoices({ page });
  const updateStatus = useUpdateInvoiceStatus();
  const navigate = useNavigate();
  const invoices: Invoice[] = data?.invoices ?? [];
  const hasMore = data?.hasMore ?? false;

  // H-1: the row "Paid" action records an actual payment through the
  // RecordPaymentModal flow; only after the payment settles does the invoice
  // status flip to Paid (payment recording already invalidates ['invoice', id]).
  const handlePaymentRecorded = (): void => {
    if (!payTarget) return;
    const target = payTarget;
    setPayTarget(null);
    updateStatus.mutate(
      { id: target.id, status: 'Paid' },
      {
        onSuccess: () => toast.success(t('invoices.markedPaid', { id: target.id })),
        onError: (error) =>
          toast.error(
            getOneKitErrorMessage(error, t('invoices.markPaidFailed', { id: target.id }))
          ),
      }
    );
  };

  const counts: Partial<Record<InvoiceFilterValue, number>> = {
    All: invoices.length,
    Draft: invoices.filter((i) => i.status === 'Draft').length,
    Sent: invoices.filter((i) => i.status === 'Sent').length,
    Paid: invoices.filter((i) => i.status === 'Paid').length,
    Overdue: invoices.filter((i) => i.status === 'Overdue').length,
    Cancelled: invoices.filter((i) => i.status === 'Cancelled').length,
  };

  const filtered =
    activeTab === 'All' ? invoices : invoices.filter((inv) => inv.status === activeTab);
  const totalFiltered = filtered.reduce((s, inv) => s + inv.amount, 0);

  const toggle = (id: string) =>
    setSelected((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });

  const runBulkSend = async (): Promise<void> => {
    const ids = Array.from(selected);
    setBulkSending(true);
    try {
      const result = await sendInvoicesByIds(ids);
      if (result.succeeded.length > 0 && result.failed.length === 0) {
        toast.success(t('invoices.bulkSent', { count: result.succeeded.length }));
      } else if (result.succeeded.length > 0) {
        toast.error(
          t('invoices.bulkPartial', {
            succeeded: result.succeeded.length,
            failed: result.failed.length,
          })
        );
      } else {
        toast.error(t('invoices.bulkSendFailed', { count: ids.length }));
      }
      if (result.succeeded.length > 0) {
        const succeededSet = new Set(result.succeeded);
        setSelected((prev) => new Set([...prev].filter((id) => !succeededSet.has(id))));
        await queryClient.invalidateQueries({ queryKey: ['invoices'] });
        for (const id of result.succeeded) {
          await queryClient.invalidateQueries({ queryKey: ['invoice', id] });
        }
      }
    } finally {
      setBulkSending(false);
      setShowBulkSend(false);
    }
  };

  const runExport = async (format: 'quickbooks' | 'xero'): Promise<void> => {
    setExporting(true);
    try {
      const result = await fetchAllInvoicesForExport();
      const payload = result.invoices.map((i) => ({
        id: i.id ?? '',
        client: i.clientName ?? '',
        amount: i.total ?? 0,
        issuedDate: i.issueDate ?? '',
        dueDate: i.dueDate ?? '',
        status: i.status ?? '',
      }));
      const options: InvoiceExportOptions = result.truncated
        ? { truncatedPages: result.pagesFetched }
        : {};
      if (format === 'quickbooks') exportQuickbooksInvoices(payload, options);
      else exportXeroInvoices(payload, options);
      if (result.truncated) {
        toast.info(t('invoices.exportCapped', { pages: MAX_EXPORT_PAGES, count: payload.length }));
      }
    } catch (error) {
      toast.error(getOneKitErrorMessage(error, t('errors.generic')));
    } finally {
      setExporting(false);
    }
  };

  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      transition={{ duration: 0.2 }}
      style={{ display: 'flex', flexDirection: 'column', gap: 20 }}
    >
      <HelmetContainer title="Invoices · rtylr" />
      <PageHeader
        title="Invoices"
        subtitle={
          <>
            {filtered.length} invoice{filtered.length !== 1 ? 's' : ''} <>&middot;</>{' '}
            <MoneyDisplay amount={totalFiltered} /> total
          </>
        }
        actions={
          <>
            {selected.size > 0 && (
              <button type="button" style={ghostBtn} onClick={() => setShowBulkSend(true)}>
                <Send size={13} /> {t('invoices.sendSelected', { count: selected.size })}
              </button>
            )}
            <button
              type="button"
              style={ghostBtn}
              onClick={() => void runExport('quickbooks')}
              disabled={exporting}
            >
              <FileSpreadsheet size={13} /> QuickBooks
            </button>
            <button
              type="button"
              style={ghostBtn}
              onClick={() => void runExport('xero')}
              disabled={exporting}
            >
              <Download size={13} /> Xero
            </button>
            <Link to="/invoices/recurring" style={{ ...ghostBtn, textDecoration: 'none' }}>
              Recurring
            </Link>
            <button
              type="button"
              style={primaryBtn}
              onClick={() => navigate({ to: '/invoices/new' })}
            >
              <Plus size={15} /> New Invoice
            </button>
          </>
        }
      />

      {/* Badge counts come from one page only - hide them while more pages exist. */}
      <InvoiceStatusFilter
        value={activeTab}
        onChange={setActiveTab}
        counts={hasMore ? undefined : counts}
      />

      {loading ? (
        <div style={{ ...cardStyle, padding: 20 }}>
          <SkeletonRow />
        </div>
      ) : filtered.length === 0 ? (
        <EmptyState
          icon={FileText}
          title={t('invoices.emptyTitle')}
          description={t('invoices.emptyDescription')}
          action={{
            label: t('invoices.emptyCta'),
            onClick: () => navigate({ to: '/invoices/new' }),
          }}
        />
      ) : (
        <div style={{ ...cardStyle, overflowX: 'auto' }}>
          <table style={{ width: '100%', minWidth: 720, borderCollapse: 'collapse' }}>
            <thead>
              <tr style={{ borderBottom: '1px solid var(--hub-border)' }}>
                {['', 'Invoice #', 'Client', 'Amount', 'Status', 'Due Date', 'Actions'].map(
                  (h, i) => (
                    <th
                      key={i}
                      style={{
                        padding: '10px 16px',
                        textAlign: 'start',
                        fontSize: 11,
                        fontWeight: 600,
                        textTransform: 'uppercase',
                        letterSpacing: '0.07em',
                        color: 'var(--hub-text-dim)',
                      }}
                    >
                      {h}
                    </th>
                  )
                )}
              </tr>
            </thead>
            <tbody>
              {filtered.map((inv, i) => (
                <tr
                  key={inv.id}
                  style={{
                    borderBottom: i < filtered.length - 1 ? '1px solid var(--hub-border)' : 'none',
                  }}
                  onMouseEnter={(e) => (e.currentTarget.style.background = 'var(--hub-hover)')}
                  onMouseLeave={(e) => (e.currentTarget.style.background = 'transparent')}
                >
                  <td style={{ padding: '12px 16px', width: 32 }}>
                    <input
                      type="checkbox"
                      checked={selected.has(inv.id)}
                      onChange={() => toggle(inv.id)}
                      aria-label={t('invoices.selectInvoice', { id: inv.number ?? inv.id })}
                      style={{ accentColor: ACCENT }}
                    />
                  </td>
                  <td style={{ padding: '12px 16px', fontSize: 13, fontWeight: 600 }}>
                    <Link
                      to="/invoices/$id"
                      params={{ id: String(inv.id) }}
                      style={{
                        display: 'flex',
                        alignItems: 'center',
                        gap: 8,
                        color: 'var(--hub-text)',
                        textDecoration: 'none',
                      }}
                    >
                      <FileText size={13} style={{ color: 'var(--hub-text-dim)' }} />
                      {inv.number ?? inv.id}
                    </Link>
                  </td>
                  <td style={{ padding: '12px 16px', fontSize: 13, color: 'var(--hub-text)' }}>
                    {inv.client}
                  </td>
                  <td
                    style={{
                      padding: '12px 16px',
                      fontSize: 13,
                      fontWeight: 600,
                      color: 'var(--hub-text)',
                    }}
                  >
                    <MoneyDisplay amount={inv.amount} />
                  </td>
                  <td style={{ padding: '12px 16px' }}>
                    <span
                      style={{
                        padding: '3px 10px',
                        borderRadius: 20,
                        fontSize: 11,
                        fontWeight: 600,
                        ...statusStyle[inv.status],
                      }}
                    >
                      {inv.status}
                    </span>
                  </td>
                  <td
                    style={{
                      padding: '12px 16px',
                      fontSize: 12,
                      color:
                        inv.status === 'Overdue'
                          ? 'var(--finance-danger)'
                          : 'var(--hub-text-muted)',
                    }}
                  >
                    {inv.dueDate}
                  </td>
                  <td style={{ padding: '12px 16px' }}>
                    <div style={{ display: 'flex', gap: 6 }}>
                      <Link
                        to="/invoices/$id"
                        params={{ id: String(inv.id) }}
                        style={{ ...ghostBtn, padding: '4px 10px', textDecoration: 'none' }}
                      >
                        <Eye size={12} /> View
                      </Link>
                      <DownloadPdfButton invoiceId={inv.id} label="PDF" />
                      {inv.status === 'Draft' && (
                        <button
                          type="button"
                          style={{
                            ...ghostBtn,
                            padding: '4px 10px',
                            color: 'var(--finance-info)',
                            borderColor: 'var(--finance-info-border)',
                          }}
                          onClick={() => setSendTarget(inv)}
                          aria-label={`${t('invoices.sendAria')} ${inv.number ?? inv.id}`}
                        >
                          <Send size={12} /> Send
                        </button>
                      )}
                      {(inv.status === 'Sent' || inv.status === 'Overdue') && (
                        <button
                          type="button"
                          style={{
                            ...ghostBtn,
                            padding: '4px 10px',
                            color: 'var(--finance-accent)',
                            borderColor: 'var(--finance-accent-border)',
                          }}
                          onClick={() => setPayTarget(inv)}
                          aria-label={`${t('invoices.recordPaymentAria')} ${inv.number ?? inv.id}`}
                        >
                          <CheckCircle size={12} /> Paid
                        </button>
                      )}
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {(page > 1 || hasMore) && (
        <div className="flex items-center justify-between px-1 py-3">
          <p className="text-xs" style={{ color: 'var(--hub-text-dim)' }}>
            Page {page}
          </p>
          <div className="flex items-center gap-2">
            <button
              type="button"
              onClick={() => setPage((p) => Math.max(1, p - 1))}
              disabled={page <= 1}
              className="px-3 py-1.5 rounded-md text-xs font-medium transition-colors disabled:opacity-30"
              style={{ background: 'var(--hub-hover)', color: 'var(--hub-text-muted)' }}
            >
              Prev
            </button>
            <button
              type="button"
              onClick={() => setPage((p) => p + 1)}
              disabled={!hasMore}
              className="px-3 py-1.5 rounded-md text-xs font-medium transition-colors disabled:opacity-30"
              style={{ background: 'var(--hub-hover)', color: 'var(--hub-text-muted)' }}
            >
              Next
            </button>
          </div>
        </div>
      )}

      <ConfirmDialog
        open={showBulkSend}
        onClose={() => setShowBulkSend(false)}
        onConfirm={() => void runBulkSend()}
        title={t('invoices.bulkSendTitle', { count: selected.size })}
        message={t('invoices.bulkSendMessage', { count: selected.size })}
        confirmLabel={bulkSending ? t('invoices.sending') : t('invoices.confirmSend')}
        pending={bulkSending}
      />

      <SendInvoiceModal
        open={sendTarget !== null}
        onClose={() => setSendTarget(null)}
        invoiceId={sendTarget?.id ?? ''}
      />

      <RecordPaymentModal
        open={payTarget !== null}
        onClose={() => setPayTarget(null)}
        invoiceId={payTarget?.id ?? ''}
        amount={(payTarget?.amount ?? 0) - (payTarget?.paid ?? 0)}
        currency={payTarget?.currency}
        onRecorded={handlePaymentRecorded}
      />
    </motion.div>
  );
};

export default Invoices;
