import React, { useState } from 'react';
import { pageTitle } from '../../shared/lib/pageTitle';
import { HelmetContainer } from '@voxire/components';
import { motion } from 'motion/react';
import { Play, Pause, Repeat, AlertTriangle } from 'lucide-react';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { getOneKitErrorMessage } from '@voxire/api';
import { EmptyState } from '@voxire/ui';
import { useUserStore } from '../../../../shared/stores/userStore';
import PageHeader from '../../shared/components/PageHeader';
import { SkeletonRow } from '../../shared/components/Skeleton';
import Modal from '../../shared/components/Modal';
import {
  useRecurringInvoices,
  useSetRecurringStatus,
  useCreateRecurringInvoice,
} from './useRecurringInvoices';
import {
  cardStyle,
  ghostBtn,
  primaryBtn,
  ACCENT,
  inputStyle,
  labelStyle,
} from '../../shared/lib/tokens';
import { canFinancePermission } from '../../shared/lib/financePermissions';

// The stored value is the frequency; only its label is translated. A second
// copy of these four words lived here and a third in the options below.
const FREQUENCIES = ['weekly', 'monthly', 'quarterly', 'yearly'] as const;

interface NewRecurringForm {
  templateInvoiceId: number;
  frequency: string;
  nextInvoiceDate: string;
}

import { useInvoices } from '../invoices/useInvoices';
import { todayIso } from '../../shared/lib/date';

const NewRecurringModal: React.FC<{ open: boolean; onClose: () => void }> = ({ open, onClose }) => {
  const { t } = useTranslation('financeCommon');
  const create = useCreateRecurringInvoice();
  const [templatePage, setTemplatePage] = useState(1);
  const {
    data: invoicesData,
    isFetching: invoicesFetching,
    isError: invoicesError,
    refetch: refetchInvoices,
  } = useInvoices({ page: templatePage, pageSize: 100 });
  const invoiceList = invoicesData?.invoices ?? [];

  const emptyForm = (): NewRecurringForm => ({
    templateInvoiceId: 0,
    frequency: 'monthly',
    nextInvoiceDate: todayIso(),
  });
  const [form, setForm] = useState<NewRecurringForm>(emptyForm);

  const handleSubmit = (): void => {
    if (form.templateInvoiceId <= 0) {
      toast.error(t('financeRecurringInvoices.templateRequired'));
      return;
    }
    create.mutate(
      {
        templateInvoiceId: String(form.templateInvoiceId),
        frequency: form.frequency,
        nextInvoiceDate: form.nextInvoiceDate,
        isActive: true,
      },
      {
        onSuccess: () => {
          toast.success(t('financeRecurringInvoices.created'));
          onClose();
          setTemplatePage(1);
          setForm(emptyForm());
        },
        onError: (error) =>
          toast.error(getOneKitErrorMessage(error, 'Failed to create recurring invoice')),
      }
    );
  };

  return (
    <Modal
      open={open}
      onClose={onClose}
      title={t('financeRecurringInvoices.new')}
      width={420}
      footer={
        <>
          <button type="button" style={ghostBtn} onClick={onClose} disabled={create.isPending}>
            {t('financeCommon.cancel')}
          </button>
          <button
            type="button"
            style={{ ...primaryBtn, opacity: create.isPending ? 0.6 : 1 }}
            onClick={handleSubmit}
            disabled={create.isPending}
          >
            {create.isPending ? 'Saving…' : 'Create'}
          </button>
        </>
      }
    >
      <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
          <label style={labelStyle}>{t('financeRecurringInvoices.templateInvoice')}</label>
          <select
            value={form.templateInvoiceId || ''}
            onChange={(e) =>
              setForm((f) => ({ ...f, templateInvoiceId: Number(e.target.value) || 0 }))
            }
            style={inputStyle}
            disabled={invoicesFetching || invoicesError}
          >
            <option value="">{t('financeRecurringInvoices.templatePlaceholder')}</option>
            {form.templateInvoiceId > 0 &&
              !invoiceList.some(
                (invoice) => String(invoice.id) === String(form.templateInvoiceId)
              ) && (
                <option
                  value={form.templateInvoiceId}
                >{`Invoice #${form.templateInvoiceId}`}</option>
              )}
            {invoiceList.map((inv) => (
              <option key={inv.id} value={inv.id}>
                {inv.number ? `Invoice #${inv.number}` : `Invoice ${inv.id}`} — {inv.client} ($
                {(inv.amount ?? 0).toFixed(2)})
              </option>
            ))}
          </select>
          {invoicesError ? (
            <button
              type="button"
              style={{ ...ghostBtn, alignSelf: 'flex-end' }}
              onClick={() => void refetchInvoices()}
            >
              {t('financeCommon.refresh')}
            </button>
          ) : invoicesFetching && invoiceList.length === 0 ? (
            <span style={{ fontSize: 12, color: 'var(--hub-text-dim)' }}>
              {t('financeCommon.loading')}
            </span>
          ) : null}
          {(templatePage > 1 || invoicesData?.hasMore) && (
            <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 8 }}>
              <button
                type="button"
                style={ghostBtn}
                onClick={() => setTemplatePage((page) => Math.max(1, page - 1))}
                disabled={templatePage === 1 || invoicesFetching}
              >
                {t('financeCommon.prev')}
              </button>
              <button
                type="button"
                style={ghostBtn}
                onClick={() => setTemplatePage((page) => page + 1)}
                disabled={!invoicesData?.hasMore || invoicesFetching}
              >
                {t('financeCommon.next')}
              </button>
            </div>
          )}
        </div>
        <div style={{ display: 'flex', gap: 12 }}>
          <div style={{ flex: 1, display: 'flex', flexDirection: 'column', gap: 6 }}>
            <label style={labelStyle}>{t('financeRecurringInvoices.frequencyLabel')}</label>
            <select
              value={form.frequency}
              onChange={(e) => setForm((f) => ({ ...f, frequency: e.target.value }))}
              style={inputStyle}
            >
              {FREQUENCIES.map((f) => (
                <option key={f} value={f}>
                  {t(`financeRecurringInvoices.frequency.${f}`)}
                </option>
              ))}
            </select>
          </div>
          <div style={{ flex: 1, display: 'flex', flexDirection: 'column', gap: 6 }}>
            <label style={labelStyle}>{t('financeRecurringInvoices.nextInvoiceDate')}</label>
            <input
              type="date"
              value={form.nextInvoiceDate}
              onChange={(e) => setForm((f) => ({ ...f, nextInvoiceDate: e.target.value }))}
              style={inputStyle}
            />
          </div>
        </div>
      </div>
    </Modal>
  );
};

const RecurringInvoices: React.FC = () => {
  const { t } = useTranslation('financeCommon');
  const [page, setPage] = useState(1);
  const { data, isLoading, isError, refetch } = useRecurringInvoices({ page });
  const rows = data?.invoices ?? [];
  const hasMore = data?.hasMore ?? false;
  const activeCount = rows.reduce((count, recurring) => count + Number(recurring.isActive), 0);
  const setStatus = useSetRecurringStatus();
  const [showCreate, setShowCreate] = useState(false);
  const canManageRecurringInvoices = canFinancePermission('finance:recurring_invoices:manage');

  const togglePause = (id: string, isActive: boolean): void => {
    if (!canManageRecurringInvoices) return;
    const next = isActive ? 'paused' : 'active';
    setStatus.mutate(
      { id, status: next },
      {
        onSuccess: () =>
          toast.success(
            next === 'paused' ? 'Recurring invoice paused' : 'Recurring invoice resumed'
          ),
        onError: (error) =>
          toast.error(getOneKitErrorMessage(error, 'Failed to update recurring invoice')),
      }
    );
  };

  return (
    <motion.div
      initial={{ opacity: 0, y: 4 }}
      animate={{ opacity: 1, y: 0 }}
      style={{ display: 'flex', flexDirection: 'column', gap: 20 }}
    >
      <HelmetContainer title={pageTitle(t('financeCommon.pageTitles.recurringInvoices'))} />
      <PageHeader
        title={t('financeRecurringInvoices.title')}
        subtitle={data ? `${activeCount} active subscriptions` : undefined}
        breadcrumbs={[
          { label: t('financeNav.invoices'), to: '/finance/invoices' },
          { label: t('financeNav.recurring') },
        ]}
        actions={
          canManageRecurringInvoices && !isError ? (
            <button
              type="button"
              style={primaryBtn}
              aria-label={t('financeRecurringInvoices.new')}
              onClick={() => setShowCreate(true)}
            >
              <Repeat size={13} /> New Recurring
            </button>
          ) : undefined
        }
      />

      {canManageRecurringInvoices && (
        <NewRecurringModal open={showCreate} onClose={() => setShowCreate(false)} />
      )}

      {isLoading ? (
        <div style={{ ...cardStyle, padding: 20 }}>
          <SkeletonRow />
        </div>
      ) : isError && data === undefined ? (
        <EmptyState
          icon={AlertTriangle}
          title={t('financeErrors.generic')}
          action={{
            label: t('financeCommon.refresh'),
            onClick: () => void refetch(),
          }}
        />
      ) : rows.length === 0 ? (
        <EmptyState
          icon={Repeat}
          title={t('financeRecurringInvoices.emptyTitle')}
          description={t('financeRecurringInvoices.emptyDescription')}
          action={
            canManageRecurringInvoices
              ? {
                  label: t('financeRecurringInvoices.emptyCta'),
                  onClick: () => setShowCreate(true),
                }
              : { label: t('financeCommon.refresh'), onClick: () => void refetch() }
          }
        />
      ) : (
        <div style={{ ...cardStyle, overflow: 'hidden' }}>
          <table style={{ width: '100%', borderCollapse: 'collapse' }}>
            <thead>
              <tr style={{ borderBottom: '1px solid var(--hub-border)' }}>
                {['Template', 'Frequency', 'Next run', 'Status', 'Actions'].map((h) => (
                  <th
                    key={h}
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
                ))}
              </tr>
            </thead>
            <tbody>
              {rows.map((r) => (
                <tr key={r.id} style={{ borderBottom: '1px solid var(--hub-border)' }}>
                  <td
                    style={{
                      padding: '12px 16px',
                      fontSize: 13,
                      fontWeight: 600,
                      color: 'var(--hub-text)',
                    }}
                  >
                    {`#${r.templateInvoiceId ?? '—'}`}
                  </td>
                  <td
                    style={{ padding: '12px 16px', fontSize: 12, color: 'var(--hub-text-muted)' }}
                  >
                    {FREQUENCIES.includes(r.frequency as (typeof FREQUENCIES)[number])
                      ? t(`financeRecurringInvoices.frequency.${r.frequency}`)
                      : r.frequency}
                  </td>
                  <td
                    style={{ padding: '12px 16px', fontSize: 12, color: 'var(--hub-text-muted)' }}
                  >
                    {r.nextInvoiceDate ? r.nextInvoiceDate.slice(0, 10) : '—'}
                  </td>
                  <td style={{ padding: '12px 16px' }}>
                    <span
                      style={{
                        fontSize: 11,
                        fontWeight: 600,
                        padding: '3px 10px',
                        borderRadius: 20,
                        background: r.isActive
                          ? 'var(--finance-accent-soft)'
                          : 'var(--finance-warning-soft)',
                        color: r.isActive ? ACCENT : 'var(--finance-warning)',
                      }}
                    >
                      {r.isActive ? 'Active' : 'Paused'}
                    </span>
                  </td>
                  <td style={{ padding: '12px 16px' }}>
                    {canManageRecurringInvoices && (
                      <div style={{ display: 'flex', gap: 6 }}>
                        <button
                          type="button"
                          style={{
                            ...ghostBtn,
                            padding: '4px 10px',
                            opacity: setStatus.isPending ? 0.6 : 1,
                          }}
                          onClick={() => togglePause(r.id, r.isActive)}
                          disabled={setStatus.isPending}
                          aria-label={
                            r.isActive
                              ? `Pause recurring invoice ${r.id}`
                              : `Resume recurring invoice ${r.id}`
                          }
                        >
                          {r.isActive ? <Pause size={12} /> : <Play size={12} />}
                          {r.isActive ? 'Pause' : 'Resume'}
                        </button>
                      </div>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {!isError && (page > 1 || hasMore) && (
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            padding: '8px 0',
          }}
        >
          <span style={{ fontSize: 12, color: 'var(--hub-text-dim)' }}>
            Page {data?.page ?? page}
          </span>
          <div style={{ display: 'flex', gap: 8 }}>
            <button
              type="button"
              style={ghostBtn}
              onClick={() => setPage((current) => Math.max(1, current - 1))}
              disabled={page <= 1 || isLoading}
            >
              {t('financeCommon.prev')}
            </button>
            <button
              type="button"
              style={ghostBtn}
              onClick={() => setPage((current) => current + 1)}
              disabled={!hasMore || isLoading}
            >
              {t('financeCommon.next')}
            </button>
          </div>
        </div>
      )}
    </motion.div>
  );
};

export default RecurringInvoices;
