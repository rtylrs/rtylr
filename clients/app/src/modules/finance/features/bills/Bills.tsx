import React, { useEffect, useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { motion } from 'motion/react';
import { Plus, CheckCircle, Clock, Receipt } from 'lucide-react';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { EmptyState } from '@voxire/ui';
import PageHeader from '../../shared/components/PageHeader';
import MoneyDisplay from '../../shared/components/MoneyDisplay';
import { SkeletonRow } from '../../shared/components/Skeleton';
import Modal from '../../shared/components/Modal';
import CurrencyInput from '../../shared/components/CurrencyInput';
import { BILLS_PAGE_SIZE, useBills } from './useBills';
import { useCreateBill, useUpdateBill } from '../../shared/hooks/queries/useBills';
import type { UpdateBillBody } from '@voxire/voxie/finance/bill/v1/types';
import { useErpReceivedPurchaseOrders } from '../../shared/hooks/queries/useErpReceivedPurchaseOrders';
import { useSettings } from '../../shared/lib/settings';
import {
  cardStyle,
  ghostBtn,
  primaryBtn,
  ACCENT,
  inputStyle,
  labelStyle,
} from '../../shared/lib/tokens';
import { minorToMajor } from '../../shared/lib/money';
import { todayIso } from '../../shared/lib/date';

const STATUS_STYLE: Record<string, React.CSSProperties> = {
  pending: { background: 'var(--finance-neutral-soft)', color: 'var(--finance-text-dim)' },
  paid: { background: 'var(--finance-success-soft)', color: 'var(--finance-success)' },
  overdue: { background: 'var(--finance-danger-soft)', color: 'var(--finance-danger)' },
  scheduled: { background: 'var(--finance-info-soft)', color: 'var(--finance-info)' },
};

interface NewBillForm {
  vendor: string;
  amount: number;
  dueDate: string;
  description: string;
}

const NewBillModal: React.FC<{ open: boolean; onClose: () => void }> = ({ open, onClose }) => {
  const { t } = useTranslation('financeCommon');
  const baseCurrency = useSettings((s) => s.currency);
  const createBill = useCreateBill();
  const [form, setForm] = useState<NewBillForm>({
    vendor: '',
    amount: 0,
    dueDate: todayIso(),
    description: '',
  });

  const handleSubmit = (): void => {
    if (!form.vendor.trim()) {
      toast.error(t('financeBills.vendorRequired'));
      return;
    }
    if (form.amount <= 0) {
      toast.error(t('financeBills.amountPositive'));
      return;
    }
    createBill.mutate(
      {
        // One-off bills have no ERP supplier ID.
        supplierName: form.vendor.trim(),
        amount: minorToMajor(form.amount, baseCurrency),
        currency: baseCurrency ?? 'USD',
        dueDate: form.dueDate,
        status: 'pending',
        description: form.description || form.vendor,
      },
      {
        onSuccess: () => {
          onClose();
          setForm({
            vendor: '',
            amount: 0,
            dueDate: todayIso(),
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
      title={t('financeBills.new')}
      width={440}
      footer={
        <>
          <button type="button" style={ghostBtn} onClick={onClose} disabled={createBill.isPending}>
            {t('financeCommon.cancel')}
          </button>
          <button
            type="button"
            style={{ ...primaryBtn, opacity: createBill.isPending ? 0.6 : 1 }}
            onClick={handleSubmit}
            disabled={createBill.isPending}
          >
            <CheckCircle size={13} />
            {createBill.isPending ? t('financeCommon.saving') : t('financeBills.create')}
          </button>
        </>
      }
    >
      <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
          <label style={labelStyle}>{t('financeBills.vendor')}</label>
          <input
            value={form.vendor}
            onChange={(e) => setForm((f) => ({ ...f, vendor: e.target.value }))}
            placeholder={t('financeBills.vendorPlaceholder')}
            style={inputStyle}
          />
        </div>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
          <label style={labelStyle}>{t('financeBills.amount')}</label>
          <CurrencyInput
            value={form.amount}
            onChange={(v) => setForm((f) => ({ ...f, amount: v }))}
          />
        </div>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
          <label style={labelStyle}>{t('financeBills.dueDate')}</label>
          <input
            type="date"
            value={form.dueDate}
            onChange={(e) => setForm((f) => ({ ...f, dueDate: e.target.value }))}
            style={inputStyle}
          />
        </div>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
          <label style={labelStyle}>{t('financeBills.description')}</label>
          <input
            value={form.description}
            onChange={(e) => setForm((f) => ({ ...f, description: e.target.value }))}
            placeholder={t('financeBills.descriptionPlaceholder')}
            style={inputStyle}
          />
        </div>
      </div>
    </Modal>
  );
};

const Bills: React.FC = () => {
  const { t } = useTranslation('financeCommon');
  const { data = [], isLoading, isError, refetch } = useBills();
  // ERP->FIN-03: received purchase orders are a payables commitment Finance
  // cannot see today. Hidden entirely when ERP is absent; never zero-filled.
  const { data: receivedPos } = useErpReceivedPurchaseOrders();
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [showCreate, setShowCreate] = useState(false);
  const [page, setPage] = useState(1);

  const maxPage = Math.max(1, Math.ceil(data.length / BILLS_PAGE_SIZE));
  const visibleBills = data.slice((page - 1) * BILLS_PAGE_SIZE, page * BILLS_PAGE_SIZE);
  const hasMore = page < maxPage;

  useEffect(() => {
    setPage((current) => Math.min(current, maxPage));
  }, [maxPage]);

  const outstanding = data.filter((b) => b.status === 'pending' || b.status === 'overdue');
  const totalDue = outstanding.reduce((s: number, b) => s + (b.amount ?? 0), 0);

  const toggle = (id: string): void =>
    setSelected((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });

  const updateBill = useUpdateBill();

  /**
   * A positive int64 as the wire carries it. Anything else means "no supplier",
   * and the contract is explicit that a caller must not invent one: an empty
   * string fails the int64 pattern before the request leaves the browser.
   */
  const supplierIdOrNone = (value: unknown): string | undefined =>
    typeof value === 'string' && /^[1-9][0-9]*$/.test(value) ? value : undefined;

  /**
   * The stored record, with status moved to paid and nothing else changed.
   *
   * UpdateBillBody marks currency, dueDate, status and description @required,
   * so every one of them has to be sent on every update. Sending a substitute
   * for a value we failed to read back does not satisfy the contract, it
   * overwrites the record: description is where this app keeps the vendor name.
   */
  const paidBody = (b: (typeof data)[number]): UpdateBillBody => {
    const supplierId = supplierIdOrNone(b.supplier);
    return {
      ...(supplierId ? { supplierId } : {}),
      ...(typeof b.amount === 'number' ? { amount: b.amount } : {}),
      currency: b.currency ?? '',
      dueDate: b.dueDate ?? '',
      status: 'paid',
      description: b.description ?? '',
    };
  };

  const payAll = async (): Promise<void> => {
    const ids = Array.from(selected);
    const results = await Promise.all(
      ids.map(async (id) => {
        const b = data.find((x) => x.id === id);
        if (!b) return { id, ok: false };
        try {
          await updateBill.mutateAsync({ id, body: paidBody(b) });
          return { id, ok: true };
        } catch {
          // useUpdateBill already surfaces the per-item reason.
          return { id, ok: false };
        }
      })
    );

    const paid = results.filter((r) => r.ok).map((r) => r.id);
    const failed = results.length - paid.length;

    // Only the rows that actually moved leave the selection, so a retry does
    // not require re-picking the ones that failed.
    setSelected((prev) => {
      const next = new Set(prev);
      paid.forEach((id) => next.delete(id));
      return next;
    });

    if (failed === 0) {
      toast.success(t('financeBills.markedPaid', { count: paid.length }));
    } else if (paid.length === 0) {
      toast.error(t('financeBills.markPaidFailed', { count: failed }));
    } else {
      toast.error(t('financeBills.markPaidPartial', { paid: paid.length, failed }));
    }
  };

  return (
    <motion.div
      initial={{ opacity: 0, y: 4 }}
      animate={{ opacity: 1, y: 0 }}
      style={{ display: 'flex', flexDirection: 'column', gap: 20 }}
    >
      <HelmetContainer title={`${t('financeBills.title')} · rtylr`} />
      <PageHeader
        title={t('financeBills.title')}
        subtitle={
          isError ? (
            t('financeErrors.generic')
          ) : isLoading ? (
            t('financeCommon.loadingLabel')
          ) : (
            <>
              {t('financeBills.outstanding', { count: outstanding.length })} <>&middot;</>{' '}
              <MoneyDisplay amount={totalDue} /> {t('financeBills.total')}
            </>
          )
        }
        actions={
          <>
            {selected.size > 0 && (
              <button type="button" style={primaryBtn} onClick={() => void payAll()}>
                <CheckCircle size={13} /> {t('financeBills.paySelected', { count: selected.size })}
              </button>
            )}
            <button
              type="button"
              style={ghostBtn}
              aria-label={t('financeBills.new')}
              onClick={() => setShowCreate(true)}
            >
              <Plus size={13} /> {t('financeBills.new')}
            </button>
          </>
        }
      />

      <NewBillModal open={showCreate} onClose={() => setShowCreate(false)} />

      {receivedPos && receivedPos.length > 0 && (
        <div
          style={{
            ...cardStyle,
            padding: 16,
            border: '1px solid var(--finance-info)',
          }}
          role="region"
          aria-label={t('finance.bills.receivedFromErp')}
        >
          <div style={{ fontSize: 13, fontWeight: 700, color: 'var(--hub-text)' }}>
            {t('finance.bills.receivedFromErp')} ({receivedPos.length})
          </div>
          <div style={{ fontSize: 12, color: 'var(--hub-text-muted)', marginTop: 2 }}>
            {t('finance.bills.receivedFromErpHint')}
          </div>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 4, marginTop: 10 }}>
            {receivedPos.slice(0, 5).map((po) => (
              <div
                key={po.id}
                style={{
                  display: 'flex',
                  justifyContent: 'space-between',
                  fontSize: 12,
                  color: 'var(--hub-text)',
                }}
              >
                <span>{po.poNumber || po.id}</span>
                <span style={{ fontWeight: 600 }}>
                  <MoneyDisplay amount={po.total} />
                </span>
              </div>
            ))}
            {receivedPos.length > 5 && (
              <div style={{ fontSize: 11, color: 'var(--hub-text-dim)' }}>
                +{receivedPos.length - 5}
              </div>
            )}
          </div>
        </div>
      )}

      {isLoading ? (
        <div style={{ ...cardStyle, padding: 20 }}>
          <SkeletonRow />
        </div>
      ) : isError ? (
        <div role="alert" style={{ ...cardStyle, padding: 20 }}>
          <p style={{ margin: '0 0 12px', fontSize: 13, color: 'var(--hub-text)' }}>
            {t('financeErrors.generic')}
          </p>
          <button type="button" style={ghostBtn} onClick={() => void refetch()}>
            {t('financeCommon.refresh')}
          </button>
        </div>
      ) : data.length === 0 ? (
        <EmptyState
          icon={Receipt}
          title={t('financeBills.empty')}
          description={t('financeBills.emptyDescription')}
          action={{
            label: t('financeBills.new'),
            onClick: () => setShowCreate(true),
          }}
        />
      ) : (
        <>
          <div style={{ ...cardStyle, overflow: 'hidden' }}>
            <table style={{ width: '100%', borderCollapse: 'collapse' }}>
              <thead>
                <tr style={{ borderBottom: '1px solid var(--hub-border)' }}>
                  {[
                    '',
                    t('financeBills.number'),
                    t('financeBills.vendor'),
                    t('financeBills.category'),
                    t('financeBills.amount'),
                    t('financeBills.due'),
                    t('financeBills.status'),
                    '',
                  ].map((h, i) => (
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
                  ))}
                </tr>
              </thead>
              <tbody>
                {visibleBills.map((b) => (
                  <tr key={b.id} style={{ borderBottom: '1px solid var(--hub-border)' }}>
                    <td style={{ padding: '12px 16px', width: 32 }}>
                      <input
                        type="checkbox"
                        checked={selected.has(b.id)}
                        onChange={() => toggle(b.id)}
                        disabled={b.status === 'paid'}
                        style={{ accentColor: ACCENT }}
                      />
                    </td>
                    <td style={{ padding: '12px 16px', fontSize: 13, fontWeight: 600 }}>{b.id}</td>
                    <td style={{ padding: '12px 16px', fontSize: 13 }}>{b.vendor}</td>
                    <td
                      style={{ padding: '12px 16px', fontSize: 12, color: 'var(--hub-text-muted)' }}
                    >
                      {b.category}
                    </td>
                    <td style={{ padding: '12px 16px', fontSize: 13, fontWeight: 600 }}>
                      <MoneyDisplay amount={b.amount ?? 0} />
                    </td>
                    <td
                      style={{
                        padding: '12px 16px',
                        fontSize: 12,
                        color:
                          b.status === 'overdue'
                            ? 'var(--finance-danger)'
                            : 'var(--hub-text-muted)',
                      }}
                    >
                      {b.dueDate}
                    </td>
                    <td style={{ padding: '12px 16px' }}>
                      <span
                        style={{
                          fontSize: 11,
                          fontWeight: 600,
                          padding: '3px 10px',
                          borderRadius: 20,
                          textTransform: 'capitalize',
                          ...STATUS_STYLE[b.status ?? ''],
                        }}
                      >
                        {b.status}
                      </span>
                    </td>
                    <td style={{ padding: '12px 16px' }}>
                      {b.status === 'pending' || b.status === 'overdue' ? (
                        <button
                          type="button"
                          style={{
                            ...ghostBtn,
                            padding: '4px 10px',
                            opacity: 0.6,
                            cursor: 'not-allowed',
                          }}
                          disabled
                          title={t('financeBills.schedulingSoon')}
                        >
                          <Clock size={12} /> {t('financeBills.schedule')}
                        </button>
                      ) : (
                        <span style={{ fontSize: 11, color: 'var(--hub-text-dim)' }}>-</span>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          {(page > 1 || hasMore) && (
            <div className="flex items-center justify-between px-1 py-3">
              <p className="text-xs" style={{ color: 'var(--hub-text-dim)' }}>
                Page {page}
              </p>
              <div className="flex items-center gap-2">
                <button
                  type="button"
                  onClick={() => setPage((current) => Math.max(1, current - 1))}
                  disabled={page <= 1}
                  className="px-3 py-1.5 rounded-md text-xs font-medium transition-colors disabled:opacity-30"
                  style={{ background: 'var(--hub-hover)', color: 'var(--hub-text-muted)' }}
                >
                  {t('financeCommon.prev')}
                </button>
                <button
                  type="button"
                  onClick={() => setPage((current) => current + 1)}
                  disabled={!hasMore}
                  className="px-3 py-1.5 rounded-md text-xs font-medium transition-colors disabled:opacity-30"
                  style={{ background: 'var(--hub-hover)', color: 'var(--hub-text-muted)' }}
                >
                  {t('financeCommon.next')}
                </button>
              </div>
            </div>
          )}
        </>
      )}
    </motion.div>
  );
};

export default Bills;
