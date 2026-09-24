import React, { useEffect, useRef, useState } from 'react';
import { pageTitle } from '../../shared/lib/pageTitle';
import { HelmetContainer } from '@voxire/components';
import { Link, useNavigate } from '@tanstack/react-router';
import { motion } from 'motion/react';
import { Plus, Send, CheckCircle, Eye, FileText, Download, FileSpreadsheet } from 'lucide-react';
import { toast } from 'sonner';
import {
  createOneKitClient,
  createRestRouteClient,
  getOneKitErrorMessage,
  unwrap,
} from '@voxire/api';
import { useTranslation } from 'react-i18next';
import { EmptyState } from '@voxire/ui';
import { InvoicesServiceClient } from '@voxire/voxie/finance/invoices/v1/client';
import type { InvoiceSummary } from '@voxire/voxie/finance/invoices/v1/types';
import { Invoice } from '../../shared/services/financeApi';
import { useCreateInvoice, useInvoices } from './useInvoices';
import PageHeader from '../../shared/components/PageHeader';
import { useSettings } from '../../shared/lib/settings';
import { useUnbilledBillableEntries } from '../../shared/hooks/queries/useUnbilledBillableEntries';
import type { UnbilledBillableEntry } from '../../shared/hooks/queries/useUnbilledBillableEntries';
import MoneyDisplay from '../../shared/components/MoneyDisplay';
import { SkeletonRow } from '../../shared/components/Skeleton';
import InvoiceStatusFilter, {
  InvoiceFilterValue,
} from '../../shared/components/InvoiceStatusFilter';
import DownloadPdfButton from '../../shared/components/DownloadPdfButton';
import RecordPaymentModal from '../../shared/components/RecordPaymentModal';
import SendInvoiceModal from '../../shared/components/SendInvoiceModal';
import { exportQuickbooksInvoices, exportXeroInvoices } from '../../shared/lib/exports';
import type { InvoiceExport } from '../../shared/lib/exports';
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

let invoiceExportClient: InvoicesServiceClient | undefined;
const getInvoiceExportClient = (): InvoicesServiceClient => {
  invoiceExportClient ??= createOneKitClient(InvoicesServiceClient, 'finance');
  return invoiceExportClient;
};

const isPositiveInvoiceId = (id: string): boolean => /^[1-9]\d*$/.test(id);

const toExportRow = (invoice: InvoiceSummary): InvoiceExport => ({
  id: invoice.invoiceNumber?.trim() || invoice.id || '',
  client: invoice.clientName ?? '',
  amount: invoice.total ?? 0,
  issuedDate: invoice.issueDate ?? '',
  dueDate: invoice.dueDate ?? '',
  status: invoice.status ?? '',
});

const fetchAllInvoiceExports = async (status?: string): Promise<InvoiceExport[]> => {
  const rows: InvoiceExport[] = [];
  const pageSize = 500;
  let page = 1;

  while (true) {
    const response = await getInvoiceExportClient().listInvoices({
      status: status ?? '',
      page,
      pageSize,
    });
    const pageRows = response.invoices ?? [];
    rows.push(...pageRows.map(toExportRow));

    if (
      pageRows.length === 0 ||
      response.hasMore === false ||
      (response.hasMore === undefined && pageRows.length < pageSize)
    ) {
      return rows;
    }
    page += 1;
  }
};

let hrTimeEntriesClient: ReturnType<typeof createRestRouteClient> | undefined;
const getHrTimeEntriesClient = () => {
  hrTimeEntriesClient ??= createRestRouteClient('finance');
  return hrTimeEntriesClient;
};

interface RegisteredTimeEntry {
  id?: string | number;
  date?: string;
  description?: string;
  durationMinutes?: number;
  projectId?: string | number;
}

interface RegisteredTimeEntriesResponse {
  entries?: RegisteredTimeEntry[];
  total?: number | string;
}

interface MarkTimeEntriesResponse {
  markedCount?: number;
  skippedCount?: number;
}

const fetchRegisteredUnbilledEntries = async (
  from: string,
  to: string
): Promise<UnbilledBillableEntry[]> => {
  const entries: UnbilledBillableEntry[] = [];
  const pageSize = 100;
  let offset = 0;

  while (true) {
    const response = await unwrap(
      getHrTimeEntriesClient().get<RegisteredTimeEntriesResponse | null>('/hr/time-entries', {
        params: {
          isBillable: true,
          status: 'approved',
          fromDate: from,
          toDate: to,
          limit: pageSize,
          offset,
        },
      })
    );
    const page = response?.entries ?? [];
    entries.push(
      ...page.map((entry) => ({
        id: String(entry.id ?? ''),
        date: entry.date,
        description: entry.description,
        durationMinutes: entry.durationMinutes,
        projectId: entry.projectId == null ? undefined : String(entry.projectId),
      }))
    );
    offset += page.length;

    const total = Number(response?.total ?? entries.length);
    if (page.length === 0 || !Number.isFinite(total) || offset >= total) return entries;
  }
};

const markRegisteredEntriesInvoiced = async (
  entryIds: string[],
  invoiceId: string
): Promise<MarkTimeEntriesResponse> => {
  const numericEntryIds = entryIds.map((id) => Number(id));
  const numericInvoiceId = Number(invoiceId);
  if (
    numericEntryIds.some((id) => !Number.isSafeInteger(id) || id <= 0) ||
    !Number.isSafeInteger(numericInvoiceId) ||
    numericInvoiceId <= 0
  ) {
    throw new Error('Invalid invoice or time-entry id');
  }

  return unwrap(
    getHrTimeEntriesClient().patch<MarkTimeEntriesResponse>('/hr/time-entries/mark-invoiced', {
      entryIds: numericEntryIds,
      invoiceId: numericInvoiceId,
    })
  );
};

const Invoices: React.FC = () => {
  const { t } = useTranslation();
  const [activeTab, setActiveTab] = useState<InvoiceFilterValue>('All');
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [showBulkSend, setShowBulkSend] = useState(false);
  const [bulkSendIds, setBulkSendIds] = useState<string[]>([]);
  const [bulkSendIndex, setBulkSendIndex] = useState(0);
  const bulkSendAdvance = useRef(false);
  const [sendTarget, setSendTarget] = useState<Invoice | null>(null);
  const [paymentTarget, setPaymentTarget] = useState<Invoice | null>(null);
  const [page, setPage] = useState(1);
  const statusFilter = activeTab === 'All' ? undefined : activeTab.toLowerCase();
  const {
    data,
    isLoading: loading,
    isError,
    error,
    isFetching,
    refetch,
  } = useInvoices({
    page,
    status: statusFilter,
  });
  // The billable loop's Finance half (#836): pick approved billable HR time,
  // raise one invoice for it, and the loop stamps those hours invoiced so the
  // list stops offering them. Hidden when HR is absent or unreachable.
  const rangeFrom = toLocalIso(new Date(Date.now() - 90 * 864e5));
  const rangeTo = todayIso();
  const unbilledQuery = useUnbilledBillableEntries(rangeFrom, rangeTo);
  const [registeredUnbilled, setRegisteredUnbilled] = useState<UnbilledBillableEntry[] | null>(
    null
  );
  const [fallbackAttempted, setFallbackAttempted] = useState(false);
  const [markingInvoiced, setMarkingInvoiced] = useState(false);
  const [pickedHours, setPickedHours] = useState<Set<string>>(new Set());
  const baseCurrency = useSettings((st) => st.currency);
  const createInvoice = useCreateInvoice();
  const [billToName, setBillToName] = useState('');
  const [exporting, setExporting] = useState(false);
  const exportInFlight = useRef(false);
  const navigate = useNavigate();
  const invoices: Invoice[] = data?.invoices ?? [];
  const hasMore = data?.hasMore ?? false;
  const unbilled = unbilledQuery.data ?? registeredUnbilled;

  useEffect(() => {
    setRegisteredUnbilled(null);
    setFallbackAttempted(false);
  }, [rangeFrom, rangeTo]);

  // The shared hook intentionally degrades to null; current onion exposes the
  // registered compatibility path instead of the generated HR service path.
  useEffect(() => {
    if (unbilledQuery.data != null || unbilledQuery.isFetched !== true || fallbackAttempted) {
      return;
    }

    let cancelled = false;
    setFallbackAttempted(true);
    void fetchRegisteredUnbilledEntries(rangeFrom, rangeTo)
      .then((entries) => {
        if (!cancelled) setRegisteredUnbilled(entries);
      })
      .catch(() => {
        // HR is optional for Finance; keep the existing null/failure degradation.
      });

    return () => {
      cancelled = true;
    };
  }, [fallbackAttempted, rangeFrom, rangeTo, unbilledQuery.data, unbilledQuery.isFetched]);

  const markEntriesInvoiced = async (entryIds: string[], invoiceId: string): Promise<void> => {
    setMarkingInvoiced(true);
    try {
      const result = await markRegisteredEntriesInvoiced(entryIds, invoiceId);
      setRegisteredUnbilled((current) =>
        current ? current.filter((entry) => !entryIds.includes(entry.id)) : current
      );
      void unbilledQuery.refetch?.();
      toast.success(
        t('finance.unbilled.markedDone', {
          marked: result.markedCount ?? 0,
          skipped: result.skippedCount ?? 0,
        })
      );
    } catch (markError) {
      toast.error(getOneKitErrorMessage(markError, t('errors.generic')));
    } finally {
      setMarkingInvoiced(false);
    }
  };

  const markPaid = (inv: Invoice): void => {
    setPaymentTarget(inv);
  };

  const handleTabChange = (nextTab: InvoiceFilterValue): void => {
    if (nextTab === activeTab) return;
    setActiveTab(nextTab);
    setPage(1);
    setSelected(new Set());
  };

  const startBulkSend = (): void => {
    const ids = Array.from(selected);
    if (!ids.length || ids.some((id) => !isPositiveInvoiceId(id))) {
      toast.error(t('errors.generic'));
      return;
    }
    bulkSendAdvance.current = false;
    setBulkSendIds(ids);
    setBulkSendIndex(0);
    setShowBulkSend(true);
  };

  const handleBulkSent = (): void => {
    bulkSendAdvance.current = true;
    const nextIndex = bulkSendIndex + 1;
    if (nextIndex < bulkSendIds.length) {
      setBulkSendIndex(nextIndex);
      return;
    }
    setShowBulkSend(false);
    setBulkSendIds([]);
    setSelected(new Set());
  };

  const handleBulkClose = (): void => {
    if (bulkSendAdvance.current) {
      bulkSendAdvance.current = false;
      return;
    }
    setShowBulkSend(false);
    setBulkSendIds([]);
    setBulkSendIndex(0);
  };

  const runExport = (exporter: (rows: InvoiceExport[]) => void): void => {
    if (exportInFlight.current) return;
    exportInFlight.current = true;
    setExporting(true);
    const status = activeTab === 'All' ? undefined : activeTab.toLowerCase();
    void fetchAllInvoiceExports(status)
      .then((rows) => exporter(rows))
      .catch((exportError) => toast.error(getOneKitErrorMessage(exportError, t('errors.generic'))))
      .finally(() => {
        exportInFlight.current = false;
        setExporting(false);
      });
  };

  const counts: Partial<Record<InvoiceFilterValue, number>> | undefined =
    activeTab === 'All'
      ? {
          All: invoices.length,
          Draft: invoices.filter((i) => i.status === 'Draft').length,
          Sent: invoices.filter((i) => i.status === 'Sent').length,
          Paid: invoices.filter((i) => i.status === 'Paid').length,
          Overdue: invoices.filter((i) => i.status === 'Overdue').length,
        }
      : undefined;

  const filtered =
    activeTab === 'All' ? invoices : invoices.filter((inv) => inv.status === activeTab);
  const totalFiltered = filtered.reduce((s, inv) => s + inv.amount, 0);

  const toggle = (id: string): void =>
    setSelected((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });

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
    if (!picked.length || !client || createInvoice.isPending || markingInvoiced) return;
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
          if (!inv?.id) return;
          void markEntriesInvoiced(
            picked.map((e) => e.id),
            inv.id
          );
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
      <HelmetContainer title={pageTitle(t('nav.invoices'))} />
      <PageHeader
        title={t('invoices.list.title')}
        subtitle={
          isError && data === undefined ? (
            t('errors.generic')
          ) : loading ? (
            t('common.loadingLabel')
          ) : (
            <>
              {filtered.length} invoice{filtered.length !== 1 ? 's' : ''} <>&middot;</>{' '}
              <MoneyDisplay amount={totalFiltered} /> total
            </>
          )
        }
        actions={
          <>
            {selected.size > 0 && (
              <button type="button" style={ghostBtn} onClick={startBulkSend}>
                <Send size={13} /> Send {selected.size}
              </button>
            )}
            <button
              type="button"
              style={{ ...ghostBtn, opacity: exporting ? 0.6 : 1 }}
              disabled={exporting}
              onClick={() => runExport(exportQuickbooksInvoices)}
            >
              <FileSpreadsheet size={13} /> QuickBooks
            </button>
            <button
              type="button"
              style={{ ...ghostBtn, opacity: exporting ? 0.6 : 1 }}
              disabled={exporting}
              onClick={() => runExport(exportXeroInvoices)}
            >
              <Download size={13} /> Xero
            </button>
            <Link to="/invoices/recurring" style={{ ...ghostBtn, textDecoration: 'none' }}>
              {t('invoices.list.recurring')}
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
          aria-label={t('finance.unbilled.region')}
          style={{
            ...cardStyle,
            padding: 16,
            border: '1px solid var(--finance-info)',
          }}
        >
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <div style={{ fontSize: 13, fontWeight: 700, color: 'var(--hub-text)' }}>
              {t('finance.unbilled.title')} ({unbilled.length})
            </div>
            <button
              type="button"
              style={{ ...ghostBtn, padding: '5px 12px' }}
              disabled={
                !pickedHours.size ||
                !billToName.trim() ||
                createInvoice.isPending ||
                markingInvoiced
              }
              onClick={billSelectedHours}
            >
              {createInvoice.isPending || markingInvoiced
                ? t('finance.unbilled.working')
                : t('finance.unbilled.billSelected', { count: pickedHours.size })}
            </button>
          </div>
          <div style={{ fontSize: 12, color: 'var(--hub-text-muted)', marginTop: 2 }}>
            {t('finance.unbilled.hint')}
          </div>
          <input
            type="text"
            value={billToName}
            onChange={(e) => setBillToName(e.target.value)}
            placeholder={t('finance.unbilled.billToPlaceholder')}
            aria-label={t('finance.unbilled.billToPlaceholder')}
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

      <InvoiceStatusFilter value={activeTab} onChange={handleTabChange} counts={counts} />

      {loading ? (
        <div style={{ ...cardStyle, padding: 20 }}>
          <SkeletonRow />
        </div>
      ) : isError && data === undefined ? (
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
                      <DownloadPdfButton invoiceId={inv.id} label={t('invoices.list.pdf')} />
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
                            opacity: paymentTarget?.id === inv.id ? 0.6 : 1,
                          }}
                          onClick={() => markPaid(inv)}
                          disabled={paymentTarget?.id === inv.id}
                          aria-label={`Mark invoice ${inv.id} paid`}
                        >
                          <CheckCircle size={12} /> {paymentTarget?.id === inv.id ? '...' : 'Paid'}
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
              {t('common.prev')}
            </button>
            <button
              type="button"
              onClick={() => setPage((p) => p + 1)}
              disabled={!hasMore}
              className="px-3 py-1.5 rounded-md text-xs font-medium transition-colors disabled:opacity-30"
              style={{ background: 'var(--hub-hover)', color: 'var(--hub-text-muted)' }}
            >
              {t('common.next')}
            </button>
          </div>
        </div>
      )}

      {showBulkSend && bulkSendIds[bulkSendIndex] && (
        <SendInvoiceModal
          key={bulkSendIds[bulkSendIndex]}
          open
          onClose={handleBulkClose}
          invoiceId={bulkSendIds[bulkSendIndex]}
          onSent={handleBulkSent}
        />
      )}
      <SendInvoiceModal
        open={sendTarget !== null}
        onClose={() => setSendTarget(null)}
        invoiceId={sendTarget?.id ?? ''}
      />
      {paymentTarget && (
        <RecordPaymentModal
          key={paymentTarget.id}
          open
          onClose={() => setPaymentTarget(null)}
          invoiceId={paymentTarget.id}
          amount={Math.max(0, paymentTarget.amount - (paymentTarget.paid ?? 0))}
          currency={paymentTarget.currency}
        />
      )}
    </motion.div>
  );
};

export default Invoices;
