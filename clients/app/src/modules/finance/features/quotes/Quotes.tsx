import React, { useState } from 'react';
import { pageTitle } from '../../shared/lib/pageTitle';
import { HelmetContainer } from '@voxire/components';
import { motion } from 'motion/react';
import { Plus, ArrowRight, CheckCircle, FileSignature } from 'lucide-react';
import { toast } from 'sonner';
import { getOneKitErrorMessage } from '@voxire/api';
import { useTranslation } from 'react-i18next';
import { EmptyState } from '@voxire/ui';
import { useUserStore } from '../../../../shared/stores/userStore';
import PageHeader from '../../shared/components/PageHeader';
import MoneyDisplay from '../../shared/components/MoneyDisplay';
import { SkeletonRow } from '../../shared/components/Skeleton';
import Modal from '../../shared/components/Modal';
import CurrencyInput from '../../shared/components/CurrencyInput';
import { useQuotes, useCreateQuote, useConvertQuote, useUpdateQuoteStatus } from './useQuotes';
import { formatDate, toLocalIso } from '../../shared/lib/date';
import { minorToMajor } from '../../shared/lib/money';
import { useSettings } from '../../shared/lib/settings';
import { cardStyle, ghostBtn, primaryBtn, inputStyle, labelStyle } from '../../shared/lib/tokens';
import { canFinancePermission } from '../../shared/lib/financePermissions';

const STATUS_STYLE: Record<string, React.CSSProperties> = {
  Draft: { background: 'var(--finance-neutral-soft)', color: 'var(--finance-text-dim)' },
  Sent: { background: 'var(--finance-info-soft)', color: 'var(--finance-accent-blue)' },
  Accepted: { background: 'var(--finance-accent-soft)', color: 'var(--finance-accent)' },
  Converted: { background: 'var(--finance-accent-soft)', color: 'var(--finance-accent)' },
  Declined: { background: 'var(--finance-danger-soft)', color: 'var(--finance-danger)' },
  Expired: { background: 'var(--finance-warning-soft)', color: 'var(--finance-warning)' },
};

interface NewQuoteForm {
  client: string;
  amount: number;
  expiresAt: string;
  description: string;
}

const NewQuoteModal: React.FC<{ open: boolean; onClose: () => void }> = ({ open, onClose }) => {
  const { t } = useTranslation('financeCommon');
  const currency = useSettings((state) => state.currency);
  const createQuote = useCreateQuote();
  const [form, setForm] = useState<NewQuoteForm>({
    client: '',
    amount: 0,
    expiresAt: toLocalIso(new Date(Date.now() + 30 * 86400000)),
    description: '',
  });

  const handleSubmit = (e: React.FormEvent): void => {
    e.preventDefault();
    if (!form.client.trim()) {
      toast.error(t('financeQuotes.clientRequired'));
      return;
    }
    if (form.amount <= 0) {
      toast.error(t('financeQuotes.amountPositive'));
      return;
    }
    // CurrencyInput emits minor units, while the quote API stores major units.
    createQuote.mutate(
      {
        client: form.client,
        amount: minorToMajor(form.amount, currency),
        expiresAt: form.expiresAt,
        description: form.description || undefined,
      },
      {
        onSuccess: () => {
          toast.success(t('financeQuotes.created'));
          onClose();
          setForm({
            client: '',
            amount: 0,
            expiresAt: toLocalIso(new Date(Date.now() + 30 * 86400000)),
            description: '',
          });
        },
      }
    );
  };

  return (
    <Modal
      open={open}
      onClose={onClose}
      title={t('financeQuotes.newQuote')}
      width={440}
      footer={
        <>
          <button type="button" style={ghostBtn} onClick={onClose} disabled={createQuote.isPending}>
            {t('financeCommon.cancel')}
          </button>
          <button
            type="button"
            style={{ ...primaryBtn, opacity: createQuote.isPending ? 0.6 : 1 }}
            onClick={handleSubmit}
            disabled={createQuote.isPending}
          >
            <CheckCircle size={13} />
            {createQuote.isPending ? 'Saving…' : 'Create Quote'}
          </button>
        </>
      }
    >
      <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
          <label style={labelStyle}>{t('financeQuotes.client')}</label>
          <input
            value={form.client}
            onChange={(e) => setForm((f) => ({ ...f, client: e.target.value }))}
            placeholder={t('financeQuotes.clientPlaceholder')}
            style={inputStyle}
          />
        </div>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
          <label style={labelStyle}>{t('financeQuotes.amount')}</label>
          <CurrencyInput
            value={form.amount}
            currency={currency}
            onChange={(v) => setForm((f) => ({ ...f, amount: v }))}
          />
        </div>
        <div style={{ display: 'flex', gap: 12 }}>
          <div style={{ flex: 1, display: 'flex', flexDirection: 'column', gap: 6 }}>
            <label style={labelStyle}>{t('financeQuotes.expires')}</label>
            <input
              type="date"
              value={form.expiresAt}
              onChange={(e) => setForm((f) => ({ ...f, expiresAt: e.target.value }))}
              style={inputStyle}
            />
          </div>
        </div>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
          <label style={labelStyle}>{t('financeQuotes.description')}</label>
          <input
            value={form.description}
            onChange={(e) => setForm((f) => ({ ...f, description: e.target.value }))}
            placeholder={t('financeQuotes.notesPlaceholder')}
            style={inputStyle}
          />
        </div>
      </div>
    </Modal>
  );
};

const Quotes: React.FC = () => {
  const { t } = useTranslation('financeCommon');
  const [page, setPage] = useState(1);
  const {
    data: result,
    isLoading,
    isPending,
    isError = false,
    isFetching = false,
    refetch,
  } = useQuotes({ page });
  const data = result?.quotes ?? [];
  const hasMore = result?.hasMore ?? false;
  const isPaginated = hasMore || page > 1;
  const loading = Boolean(isLoading || isPending);
  const listUnavailable = !result && !loading && !isError;
  const convertQuote = useConvertQuote();
  const updateStatus = useUpdateQuoteStatus();
  const total = data.reduce((s: number, q) => s + (q.amount ?? 0), 0);
  const [showCreate, setShowCreate] = useState(false);
  const [selectedQuote, setSelectedQuote] = useState<(typeof data)[number] | null>(null);
  const canManageQuotes = canFinancePermission('finance:quotes:manage');

  const handleConvert = (id: string): void => {
    if (!canManageQuotes) return;
    convertQuote.mutate(id, {
      onSuccess: () => toast.success(t('financeQuotes.converted')),
      onError: (error) => toast.error(getOneKitErrorMessage(error, 'Failed to convert quote')),
    });
  };

  const handleMarkAccepted = (id: string): void => {
    if (!canManageQuotes) return;
    updateStatus.mutate(
      { id, status: 'accepted' },
      { onSuccess: () => toast.success(t('financeQuotes.accepted')) }
    );
  };

  return (
    <motion.div
      initial={{ opacity: 0, y: 4 }}
      animate={{ opacity: 1, y: 0 }}
      style={{ display: 'flex', flexDirection: 'column', gap: 20 }}
    >
      <HelmetContainer title={pageTitle(t('financeCommon.pageTitles.quotes'))} />
      <PageHeader
        title={t('financeQuotes.title')}
        subtitle={
          isError || loading || listUnavailable ? undefined : (
            <>
              {data.length} quotes <>&middot;</> <MoneyDisplay amount={total} />{' '}
              {isPaginated ? t('financeCommon.subtotal') : t('financeCommon.total')}
            </>
          )
        }
        actions={
          canManageQuotes ? (
            <button
              type="button"
              style={primaryBtn}
              aria-label={t('financeQuotes.newQuote')}
              onClick={() => setShowCreate(true)}
            >
              <Plus size={13} /> New Quote
            </button>
          ) : undefined
        }
      />

      {canManageQuotes && <NewQuoteModal open={showCreate} onClose={() => setShowCreate(false)} />}

      {isError || listUnavailable ? (
        <div
          role="alert"
          style={{ ...cardStyle, padding: 20, display: 'flex', flexDirection: 'column', gap: 12 }}
        >
          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <FileSignature size={16} style={{ color: 'var(--finance-danger)' }} />
            <span style={{ fontSize: 13, fontWeight: 600, color: 'var(--hub-text)' }}>
              {t('financeErrors.generic')}
            </span>
          </div>
          <button
            type="button"
            style={{ ...ghostBtn, cursor: isFetching ? 'wait' : 'pointer' }}
            disabled={isFetching}
            onClick={() => void refetch?.()}
          >
            {t('financeCommon.refresh')}
          </button>
        </div>
      ) : loading ? (
        <div style={{ ...cardStyle, padding: 20 }}>
          <SkeletonRow />
        </div>
      ) : data.length === 0 ? (
        <EmptyState
          icon={FileSignature}
          title={t('financeQuotes.emptyTitle')}
          description={t('financeQuotes.emptyDescription')}
          action={
            canManageQuotes
              ? {
                  label: t('financeQuotes.emptyCta'),
                  onClick: () => setShowCreate(true),
                }
              : { label: t('financeCommon.refresh'), onClick: () => void refetch?.() }
          }
        />
      ) : (
        <div style={{ ...cardStyle, overflow: 'hidden' }}>
          <table style={{ width: '100%', borderCollapse: 'collapse' }}>
            <thead>
              <tr style={{ borderBottom: '1px solid var(--hub-border)' }}>
                {['Quote #', 'Client', 'Amount', 'Status', 'Issued', 'Expires', 'Actions'].map(
                  (h) => (
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
                  )
                )}
              </tr>
            </thead>
            <tbody>
              {data.map((q) => (
                <tr key={q.id} style={{ borderBottom: '1px solid var(--hub-border)' }}>
                  <td
                    style={{
                      padding: '12px 16px',
                      fontSize: 13,
                      fontWeight: 600,
                      color: 'var(--hub-text)',
                    }}
                  >
                    {q.id}
                  </td>
                  <td style={{ padding: '12px 16px', fontSize: 13, color: 'var(--hub-text)' }}>
                    {q.client}
                  </td>
                  <td style={{ padding: '12px 16px', fontSize: 13, fontWeight: 600 }}>
                    <MoneyDisplay amount={q.amount ?? 0} />
                  </td>
                  <td style={{ padding: '12px 16px' }}>
                    <span
                      style={{
                        fontSize: 11,
                        fontWeight: 600,
                        padding: '3px 10px',
                        borderRadius: 20,
                        ...STATUS_STYLE[q.status ?? ''],
                      }}
                    >
                      {q.status}
                    </span>
                  </td>
                  <td
                    style={{ padding: '12px 16px', fontSize: 12, color: 'var(--hub-text-muted)' }}
                  >
                    {formatDate(q.issuedDate ?? '')}
                  </td>
                  <td
                    style={{ padding: '12px 16px', fontSize: 12, color: 'var(--hub-text-muted)' }}
                  >
                    {q.expiresAt}
                  </td>
                  <td style={{ padding: '12px 16px' }}>
                    {q.status === 'Accepted' && canManageQuotes ? (
                      <button
                        type="button"
                        style={{
                          ...ghostBtn,
                          padding: '4px 10px',
                          color: 'var(--finance-accent)',
                          borderColor: 'var(--finance-accent-border)',
                          opacity: convertQuote.isPending ? 0.6 : 1,
                        }}
                        aria-label={`Convert quote ${q.id} to invoice`}
                        onClick={() => handleConvert(q.id)}
                        disabled={convertQuote.isPending}
                      >
                        <ArrowRight size={12} /> Convert
                      </button>
                    ) : q.status === 'Sent' && canManageQuotes ? (
                      <button
                        type="button"
                        style={{
                          ...ghostBtn,
                          padding: '4px 10px',
                          opacity: updateStatus.isPending ? 0.6 : 1,
                        }}
                        aria-label={`Mark quote ${q.id} as accepted`}
                        onClick={() => handleMarkAccepted(q.id)}
                        disabled={updateStatus.isPending}
                      >
                        <CheckCircle size={12} /> Mark Accepted
                      </button>
                    ) : (
                      <button
                        type="button"
                        style={{ ...ghostBtn, padding: '4px 10px' }}
                        aria-label={`View quote ${q.id}`}
                        onClick={() => setSelectedQuote(q)}
                      >
                        {t('financeQuotes.view')}
                      </button>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      <Modal
        open={selectedQuote !== null}
        onClose={() => setSelectedQuote(null)}
        title={
          selectedQuote ? `${t('financeQuotes.view')} ${selectedQuote.id}` : t('financeQuotes.view')
        }
        width={480}
      >
        {selectedQuote && (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', gap: 16 }}>
              <span style={{ color: 'var(--hub-text-muted)', fontSize: 12 }}>
                {t('financeQuotes.client')}
              </span>
              <strong style={{ color: 'var(--hub-text)', fontSize: 13, textAlign: 'end' }}>
                {selectedQuote.client}
              </strong>
            </div>
            <div style={{ display: 'flex', justifyContent: 'space-between', gap: 16 }}>
              <span style={{ color: 'var(--hub-text-muted)', fontSize: 12 }}>
                {t('financeCommon.amount')}
              </span>
              <MoneyDisplay amount={selectedQuote.amount ?? 0} />
            </div>
            <div style={{ display: 'flex', justifyContent: 'space-between', gap: 16 }}>
              <span style={{ color: 'var(--hub-text-muted)', fontSize: 12 }}>
                {t('financeCommon.status')}
              </span>
              <span style={{ color: 'var(--hub-text)', fontSize: 13 }}>{selectedQuote.status}</span>
            </div>
            <div style={{ display: 'flex', justifyContent: 'space-between', gap: 16 }}>
              <span style={{ color: 'var(--hub-text-muted)', fontSize: 12 }}>
                {t('financeCommon.date')}
              </span>
              <span style={{ color: 'var(--hub-text)', fontSize: 13 }}>
                {formatDate(selectedQuote.issuedDate ?? '')}
              </span>
            </div>
            <div style={{ display: 'flex', justifyContent: 'space-between', gap: 16 }}>
              <span style={{ color: 'var(--hub-text-muted)', fontSize: 12 }}>
                {t('financeQuotes.expires')}
              </span>
              <span style={{ color: 'var(--hub-text)', fontSize: 13 }}>
                {selectedQuote.expiresAt}
              </span>
            </div>
          </div>
        )}
      </Modal>

      {!isError && !loading && !listUnavailable && (page > 1 || hasMore) && (
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
              {t('financeQuotes.prev')}
            </button>
            <button
              type="button"
              onClick={() => setPage((p) => p + 1)}
              disabled={!hasMore}
              className="px-3 py-1.5 rounded-md text-xs font-medium transition-colors disabled:opacity-30"
              style={{ background: 'var(--hub-hover)', color: 'var(--hub-text-muted)' }}
            >
              {t('financeQuotes.next')}
            </button>
          </div>
        </div>
      )}
    </motion.div>
  );
};

export default Quotes;
