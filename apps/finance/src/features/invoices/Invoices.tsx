import React, { useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { Link, useNavigate } from '@tanstack/react-router';
import { motion } from 'motion/react';
import { Plus, Send, CheckCircle, Eye, FileText, Download, FileSpreadsheet } from 'lucide-react';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { Invoice } from '../../shared/services/financeApi';
import { useInvoices, useUpdateInvoiceStatus } from './useInvoices';
import PageHeader from '../../shared/components/PageHeader';
import MoneyDisplay from '../../shared/components/MoneyDisplay';
import EmptyState from '../../shared/components/EmptyState';
import { SkeletonRow } from '../../shared/components/Skeleton';
import InvoiceStatusFilter, {
  InvoiceFilterValue,
} from '../../shared/components/InvoiceStatusFilter';
import DownloadPdfButton from '../../shared/components/DownloadPdfButton';
import SendInvoiceModal from '../../shared/components/SendInvoiceModal';
import { exportQuickbooksInvoices, exportXeroInvoices } from '../../shared/lib/exports';
import { cardStyle, ghostBtn, primaryBtn, ACCENT } from '../../shared/lib/tokens';

type InvoiceStatus = 'Draft' | 'Sent' | 'Paid' | 'Overdue';

const statusStyle: Record<InvoiceStatus, React.CSSProperties> = {
  Draft: { background: 'rgba(113,113,122,0.15)', color: '#a1a1aa' },
  Sent: { background: 'rgba(59,130,246,0.12)', color: '#3b82f6' },
  Paid: { background: 'rgba(16,185,129,0.12)', color: '#10b981' },
  Overdue: { background: 'rgba(239,68,68,0.12)', color: '#ef4444' },
};

const Invoices: React.FC = () => {
  const { t } = useTranslation();
  const [activeTab, setActiveTab] = useState<InvoiceFilterValue>('All');
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [showBulkSend, setShowBulkSend] = useState(false);
  const [sendTarget, setSendTarget] = useState<Invoice | null>(null);
  const [page, setPage] = useState(1);
  const { data, isLoading: loading } = useInvoices({ page });
  const updateStatus = useUpdateInvoiceStatus();
  const navigate = useNavigate();
  const invoices: Invoice[] = data?.invoices ?? [];
  const hasMore = data?.hasMore ?? false;

  const markPaid = (inv: Invoice): void => {
    updateStatus.mutate(
      { id: inv.id, status: 'Paid' },
      {
        onSuccess: () => toast.success(t('invoices.markedPaid', { id: inv.id })),
        onError: () => toast.error(t('invoices.markPaidFailed', { id: inv.id })),
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
              <button style={ghostBtn} onClick={() => setShowBulkSend(true)}>
                <Send size={13} /> Send {selected.size}
              </button>
            )}
            <button style={ghostBtn} onClick={() => exportQuickbooksInvoices(exportPayload)}>
              <FileSpreadsheet size={13} /> QuickBooks
            </button>
            <button style={ghostBtn} onClick={() => exportXeroInvoices(exportPayload)}>
              <Download size={13} /> Xero
            </button>
            <Link to="/invoices/recurring" style={{ ...ghostBtn, textDecoration: 'none' }}>
              Recurring
            </Link>
            <button style={primaryBtn} onClick={() => navigate({ to: '/invoices/new' })}>
              <Plus size={15} /> New Invoice
            </button>
          </>
        }
      />

      <InvoiceStatusFilter value={activeTab} onChange={setActiveTab} counts={counts} />

      {loading ? (
        <div style={{ ...cardStyle, padding: 20 }}>
          <SkeletonRow />
        </div>
      ) : filtered.length === 0 ? (
        <EmptyState
          title="No invoices yet"
          description="Create your first invoice to start billing clients."
          action={
            <button style={primaryBtn} onClick={() => navigate({ to: '/invoices/new' })}>
              <Plus size={13} /> New Invoice
            </button>
          }
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
                        textAlign: 'left',
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
                      color: inv.status === 'Overdue' ? '#ef4444' : 'var(--hub-text-muted)',
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
                          style={{
                            ...ghostBtn,
                            padding: '4px 10px',
                            color: '#3b82f6',
                            borderColor: 'rgba(59,130,246,0.3)',
                          }}
                          onClick={() => setSendTarget(inv)}
                          aria-label={`Send invoice ${inv.id}`}
                        >
                          <Send size={12} /> Send
                        </button>
                      )}
                      {(inv.status === 'Sent' || inv.status === 'Overdue') && (
                        <button
                          style={{
                            ...ghostBtn,
                            padding: '4px 10px',
                            color: '#10b981',
                            borderColor: 'rgba(16,185,129,0.3)',
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
              onClick={() => setPage((p) => Math.max(1, p - 1))}
              disabled={page <= 1}
              className="px-3 py-1.5 rounded-md text-xs font-medium transition-colors disabled:opacity-30"
              style={{ background: 'var(--hub-hover)', color: 'var(--hub-text-muted)' }}
            >
              Prev
            </button>
            <button
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
        defaultTo={sendTarget?.clientEmail}
      />
    </motion.div>
  );
};

export default Invoices;
