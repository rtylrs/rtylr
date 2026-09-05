import React, { useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { Link, useNavigate } from '@tanstack/react-router';
import { motion } from 'motion/react';
import { Plus, Send, CheckCircle, Eye, FileText, Download, FileSpreadsheet } from 'lucide-react';
import { toast } from 'sonner';
import { getOneKitErrorMessage } from '@voxire/api';
import { useTranslation } from 'react-i18next';
import { EmptyState } from '@voxire/ui';
import { Invoice } from '../../shared/services/financeApi';
import { useInvoices, useUpdateInvoiceStatus } from './useInvoices';
import PageHeader from '../../shared/components/PageHeader';
import { useSettings } from '../../shared/lib/settings';
import { useCreateInvoice } from './useInvoices';
import {
  useUnbilledBillableEntries,
  useMarkEntriesInvoiced,
} from '../../shared/hooks/queries/useUnbilledBillableEntries';
import MoneyDisplay from '../../shared/components/MoneyDisplay';
import { SkeletonRow } from '../../shared/components/Skeleton';
import InvoiceStatusFilter, {
  InvoiceFilterValue,
} from '../../shared/components/InvoiceStatusFilter';
import DownloadPdfButton from '../../shared/components/DownloadPdfButton';
import SendInvoiceModal from '../../shared/components/SendInvoiceModal';
import { exportQuickbooksInvoices, exportXeroInvoices } from '../../shared/lib/exports';
import { cardStyle, ghostBtn, primaryBtn, ACCENT } from '../../shared/lib/tokens';
import { toLocalIso, todayIso } from '../../shared/lib/date';

type InvoiceStatus = 'Draft' | 'Sent' | 'Paid' | 'Overdue' | 'Cancelled';

const statusStyle: Record<InvoiceStatus, React.CSSProperties> = {
  Draft: { background: 'var(--finance-neutral-soft)', color: 'var(--finance-text-dim)' },
  Sent: { background: 'var(--finance-info-soft)', color: 'var(--finance-info)' },
  Paid: { background: 'var(--finance-accent-soft)', color: 'var(--finance-accent)' },
  Overdue: { background: 'var(--finance-danger-soft)', color: 'var(--finance-danger)' },
  Cancelled: { background: 'var(--finance-neutral-soft)', color: 'var(--finance-text-dim)' },
};

const Invoices: React.FC = () => {
  const { t } = useTranslation();
  const [activeTab, setActiveTab] = useState<InvoiceFilterValue>('All');
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [showBulkSend, setShowBulkSend] = useState(false);
  const [sendTarget, setSendTarget] = useState<Invoice | null>(null);
  const [page, setPage] = useState(1);
  const { data, isLoading: loading } = useInvoices({ page });
  // The billable loop's Finance half (#836): pick approved billable HR time,
  // raise one invoice for it, and the loop stamps those hours invoiced so the
  // list stops offering them. Hidden when HR is absent or unreachable.
  const rangeFrom = toLocalIso(new Date(Date.now() - 90 * 864e5));
  const rangeTo = todayIso();
  const { data: unbilled } = useUnbilledBillableEntries(rangeFrom, rangeTo);
  const [pickedHours, setPickedHours] = useState<Set<string>>(new Set());
  const markInvoiced = useMarkEntriesInvoiced();
  const baseCurrency = useSettings((st) => st.currency);
  const createInvoice = useCreateInvoice();
  const [billToName, setBillToName] = useState('');
  const updateStatus = useUpdateInvoiceStatus();
  const navigate = useNavigate();
  const invoices: Invoice[] = data?.invoices ?? [];
  const hasMore = data?.hasMore ?? false;

  const markPaid = (inv: Invoice): void => {
    updateStatus.mutate(
      { id: inv.id, status: 'Paid' },
      {
        onSuccess: () => toast.success(t('invoices.markedPaid', { id: inv.id })),
        onError: (error) =>
          toast.error(getOneKitErrorMessage(error, t('invoices.markPaidFailed', { id: inv.id }))),
      }
    );
  };

  const counts: Partial<Record<InvoiceFilterValue, number>> = {
    All: invoices.length,
    Draft: invoices.filter((i) => i.status === 'Draft').length,
    Sent: invoices.filter((i) => i.status === 'Sent').length,
    Paid: invoices.filter((i) => i.status === 'Paid').length,
    Overdue: invoices.filter((i) => i.status === 'Overdue').length,
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

  const exportPayload = (activeTab === 'All' ? invoices : filtered).map((i) => ({
    id: i.id,
    client: i.client,
    amount: i.amount,
    issuedDate: i.issuedDate,
    dueDate: i.dueDate,
    status: i.status,
  }));

  const toggleHour = (id: string): void =>
    setPickedHours((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });

  // One invoice for the picked hours, then the loop stamps them invoiced.
  // Pricing stays manual on purpose: rates never leave HR (#836's RBAC rule),
  // so the draft opens with the hours described and the amounts blank.
  const billSelectedHours = (): void => {
    const picked = (unbilled ?? []).filter((e) => pickedHours.has(e.id));
    const client = billToName.trim();
    if (!picked.length || !client) return;
    createInvoice.mutate(
      {
        clientName: client,
        clientEmail: '',
        issueDate: rangeTo,
        dueDate: rangeTo,
        currency: baseCurrency ?? 'USD',
        discountType: 'fixed',
        items: picked.map((e, i) => ({
          description: `${e.date ?? ''} ${e.description ?? ''}`.trim() || `Time entry ${e.id}`,
          quantity: Math.round(((e.durationMinutes ?? 0) / 60) * 100) / 100,
          unit: 'hour',
          sortOrder: String(i),
        })),
      },
      {
        onSuccess: (inv) => {
          markInvoiced.mutate({ entryIds: picked.map((e) => e.id), invoiceId: inv.id });
          setPickedHours(new Set());
        },
      }
    );
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
                <Send size={13} /> Send {selected.size}
              </button>
            )}
            <button
              type="button"
              style={ghostBtn}
              onClick={() => exportQuickbooksInvoices(exportPayload)}
            >
              <FileSpreadsheet size={13} /> QuickBooks
            </button>
            <button
              type="button"
              style={ghostBtn}
              onClick={() => exportXeroInvoices(exportPayload)}
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

      {unbilled && unbilled.length > 0 && (
        <div
          role="region"
          aria-label={t('unbilled.region')}
          style={{
            ...cardStyle,
            padding: 16,
            border: '1px solid var(--finance-info)',
          }}
        >
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <div style={{ fontSize: 13, fontWeight: 700, color: 'var(--hub-text)' }}>
              {t('unbilled.title')} ({unbilled.length})
            </div>
            <button
              type="button"
              style={{ ...ghostBtn, padding: '5px 12px' }}
              disabled={
                !pickedHours.size ||
                !billToName.trim() ||
                createInvoice.isPending ||
                markInvoiced.isPending
              }
              onClick={billSelectedHours}
            >
              {createInvoice.isPending || markInvoiced.isPending
                ? t('unbilled.working')
                : t('unbilled.billSelected', { count: pickedHours.size })}
            </button>
          </div>
          <div style={{ fontSize: 12, color: 'var(--hub-text-muted)', marginTop: 2 }}>
            {t('unbilled.hint')}
          </div>
          <input
            type="text"
            value={billToName}
            onChange={(e) => setBillToName(e.target.value)}
            placeholder={t('unbilled.billToPlaceholder')}
            aria-label={t('unbilled.billToPlaceholder')}
            style={{
              marginTop: 10,
              padding: '7px 10px',
              borderRadius: 8,
              border: '1px solid var(--hub-border)',
              background: 'var(--hub-surface)',
              color: 'var(--hub-text)',
              fontSize: 16,
              maxWidth: 320,
            }}
          />
          <div style={{ display: 'flex', flexDirection: 'column', gap: 6, marginTop: 10 }}>
            {unbilled.slice(0, 8).map((e) => (
              <label
                key={e.id}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: 8,
                  fontSize: 12,
                  color: 'var(--hub-text)',
                  cursor: 'pointer',
                }}
              >
                <input
                  type="checkbox"
                  checked={pickedHours.has(e.id)}
                  onChange={() => toggleHour(e.id)}
                  style={{ width: 16, height: 16, accentColor: ACCENT }}
                />
                <span>{e.date}</span>
                <span
                  style={{
                    flex: 1,
                    overflow: 'hidden',
                    textOverflow: 'ellipsis',
                    whiteSpace: 'nowrap',
                  }}
                >
                  {e.description || e.projectId}
                </span>
                <span>{Math.round(((e.durationMinutes ?? 0) / 60) * 100) / 100} h</span>
              </label>
            ))}
            {unbilled.length > 8 && (
              <div style={{ fontSize: 11, color: 'var(--hub-text-dim)' }}>
                +{unbilled.length - 8}
              </div>
            )}
          </div>
        </div>
      )}

      <InvoiceStatusFilter value={activeTab} onChange={setActiveTab} counts={counts} />

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
                      {inv.id}
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
                          aria-label={`Send invoice ${inv.id}`}
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
                            opacity: updateStatus.isPending ? 0.6 : 1,
                          }}
                          onClick={() => markPaid(inv)}
                          disabled={updateStatus.isPending}
                          aria-label={`Mark invoice ${inv.id} paid`}
                        >
                          <CheckCircle size={12} /> {updateStatus.isPending ? '...' : 'Paid'}
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

      <SendInvoiceModal
        open={showBulkSend}
        onClose={() => setShowBulkSend(false)}
        invoiceId={`bulk (${selected.size})`}
        onSent={() => setSelected(new Set())}
      />
      <SendInvoiceModal
        open={sendTarget !== null}
        onClose={() => setSendTarget(null)}
        invoiceId={sendTarget?.id ?? ''}
      />
    </motion.div>
  );
};

export default Invoices;
