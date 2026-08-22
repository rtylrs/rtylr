import React, { useState } from 'react';
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
import ConfirmDialog from '../../shared/components/ConfirmDialog';
import { useBills } from './useBills';
import { useCreateBill, useUpdateBill } from '../../shared/hooks/queries/useBills';
import { useSettings } from '../../shared/lib/settings';
import {
  cardStyle,
  ghostBtn,
  primaryBtn,
  ACCENT,
  inputStyle,
  labelStyle,
} from '../../shared/lib/tokens';

const STATUS_STYLE: Record<string, React.CSSProperties> = {
  unpaid: { background: 'var(--finance-neutral-soft)', color: 'var(--finance-text-dim)' },
  paid: { background: 'var(--finance-accent-soft)', color: 'var(--finance-accent)' },
  overdue: { background: 'var(--finance-danger-soft)', color: 'var(--finance-danger)' },
  scheduled: { background: 'var(--finance-info-soft)', color: 'var(--finance-info)' },
};

interface NewBillForm {
  vendor: string;
  supplierId: string;
  amount: number;
  dueDate: string;
  description: string;
}

// finance.bill.v1 requires supplier_id to be a numeric id (@gt(0)); there is
// no supplier-directory client in gen/ts/finance to resolve names to ids, so
// the id must come from the user.
const isValidSupplierId = (value: string): boolean => /^[1-9][0-9]*$/.test(value.trim());

const NewBillModal: React.FC<{ open: boolean; onClose: () => void }> = ({ open, onClose }) => {
  const { t } = useTranslation();
  const baseCurrency = useSettings((s) => s.currency);
  const createBill = useCreateBill();
  const [form, setForm] = useState<NewBillForm>({
    vendor: '',
    supplierId: '',
    amount: 0,
    dueDate: new Date().toISOString().slice(0, 10),
    description: '',
  });

  const supplierIdValid = isValidSupplierId(form.supplierId);
  const canSubmit =
    !!form.vendor.trim() && form.amount > 0 && supplierIdValid && !createBill.isPending;

  const handleSubmit = (): void => {
    if (!form.vendor.trim()) {
      toast.error(t('bills.vendorRequired'));
      return;
    }
    if (!supplierIdValid) {
      toast.error(t('bills.supplierIdInvalid'));
      return;
    }
    if (form.amount <= 0) {
      toast.error(t('bills.amountPositive'));
      return;
    }
    createBill.mutate(
      {
        supplierId: form.supplierId.trim(),
        amount: form.amount,
        currency: baseCurrency ?? 'USD',
        dueDate: form.dueDate,
        status: 'unpaid',
        description: form.description || form.vendor,
      },
      {
        onSuccess: () => {
          onClose();
          setForm({
            vendor: '',
            supplierId: '',
            amount: 0,
            dueDate: new Date().toISOString().slice(0, 10),
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
      title={t('bills.new')}
      width={440}
      footer={
        <>
          <button type="button" style={ghostBtn} onClick={onClose} disabled={createBill.isPending}>
            {t('common.cancel')}
          </button>
          <button
            type="button"
            style={{ ...primaryBtn, opacity: createBill.isPending || !canSubmit ? 0.6 : 1 }}
            onClick={handleSubmit}
            disabled={!canSubmit}
          >
            <CheckCircle size={13} />
            {createBill.isPending ? t('common.saving') : t('bills.create')}
          </button>
        </>
      }
    >
      <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
          <label style={labelStyle}>{t('bills.vendor')}</label>
          <input
            value={form.vendor}
            onChange={(e) => setForm((f) => ({ ...f, vendor: e.target.value }))}
            placeholder={t('bills.vendorPlaceholder')}
            style={inputStyle}
          />
        </div>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
          <label style={labelStyle} htmlFor="new-bill-supplier-id">
            {t('bills.supplierId')}
          </label>
          <input
            id="new-bill-supplier-id"
            inputMode="numeric"
            value={form.supplierId}
            onChange={(e) => setForm((f) => ({ ...f, supplierId: e.target.value }))}
            placeholder={t('bills.supplierIdPlaceholder')}
            style={{
              ...inputStyle,
              borderColor:
                form.supplierId !== '' && !supplierIdValid
                  ? 'var(--finance-danger)'
                  : inputStyle.borderColor,
            }}
            aria-label={t('bills.supplierId')}
            aria-invalid={form.supplierId !== '' && !supplierIdValid}
          />
          <span
            style={{
              fontSize: 11,
              color:
                form.supplierId !== '' && !supplierIdValid
                  ? 'var(--finance-danger)'
                  : 'var(--hub-text-dim)',
            }}
          >
            {form.supplierId !== '' && !supplierIdValid
              ? t('bills.supplierIdInvalid')
              : t('bills.supplierIdHint')}
          </span>
        </div>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
          <label style={labelStyle}>{t('bills.amount')}</label>
          <CurrencyInput
            value={form.amount}
            onChange={(v) => setForm((f) => ({ ...f, amount: v }))}
          />
        </div>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
          <label style={labelStyle}>{t('bills.dueDate')}</label>
          <input
            type="date"
            value={form.dueDate}
            onChange={(e) => setForm((f) => ({ ...f, dueDate: e.target.value }))}
            style={inputStyle}
          />
        </div>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
          <label style={labelStyle}>{t('bills.description')}</label>
          <input
            value={form.description}
            onChange={(e) => setForm((f) => ({ ...f, description: e.target.value }))}
            placeholder={t('bills.descriptionPlaceholder')}
            style={inputStyle}
          />
        </div>
      </div>
    </Modal>
  );
};

const Bills: React.FC = () => {
  const { t } = useTranslation();
  const { data = [], isLoading } = useBills();
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [showCreate, setShowCreate] = useState(false);

  const unpaid = data.filter((b) => b.status === 'unpaid' || b.status === 'overdue');
  const totalDue = unpaid.reduce((s: number, b) => s + (b.amount ?? 0), 0);

  const toggle = (id: string): void =>
    setSelected((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });

  const updateBill = useUpdateBill();
  const [showPayConfirm, setShowPayConfirm] = useState(false);
  const [paying, setPaying] = useState(false);

  const payAll = async (): Promise<void> => {
    const ids = Array.from(selected);
    setPaying(true);
    const results = await Promise.allSettled(
      ids.map(async (id) => {
        const b = data.find((x) => x.id === id);
        const supplierId = typeof b?.supplier === 'string' ? b.supplier : '';
        const dueDate = typeof b?.dueDate === 'string' ? b.dueDate : '';
        const currency = typeof b?.currency === 'string' && b.currency ? b.currency : 'USD';
        const rawDescription =
          typeof (b as Record<string, unknown>)?.description === 'string'
            ? String((b as Record<string, unknown>).description).trim()
            : '';
        // updateBill requires a non-empty description; fall back to the vendor
        // name the list view already carries for this purpose.
        const description =
          rawDescription || (typeof b?.vendor === 'string' ? b.vendor.trim() : '') || id;
        if (!/^[1-9][0-9]*$/.test(supplierId) || !dueDate) {
          throw new Error(`Bill ${id} is missing required fields`);
        }
        await updateBill.mutateAsync({
          id,
          body: {
            supplierId,
            amount: typeof b?.amount === 'number' ? b.amount : undefined,
            currency,
            dueDate,
            status: 'paid',
            description,
          },
        });
      })
    );
    setPaying(false);
    setShowPayConfirm(false);

    const succeededIds = ids.filter((_, i) => results[i].status === 'fulfilled');
    const failedCount = ids.length - succeededIds.length;

    if (succeededIds.length > 0 && failedCount === 0) {
      toast.success(t('bills.markedPaid', { count: succeededIds.length }));
    } else if (succeededIds.length > 0) {
      toast.error(t('bills.bulkPartial', { succeeded: succeededIds.length, failed: failedCount }));
    } else {
      toast.error(t('bills.bulkFailed', { count: ids.length }));
    }

    if (succeededIds.length > 0) {
      const succeededSet = new Set(succeededIds);
      setSelected((prev) => new Set([...prev].filter((id) => !succeededSet.has(id))));
    }
  };

  return (
    <motion.div
      initial={{ opacity: 0, y: 4 }}
      animate={{ opacity: 1, y: 0 }}
      style={{ display: 'flex', flexDirection: 'column', gap: 20 }}
    >
      <HelmetContainer title={`${t('bills.title')} · rtylr`} />
      <PageHeader
        title={t('bills.title')}
        subtitle={
          <>
            {t('bills.outstanding', { count: unpaid.length })} <>&middot;</>{' '}
            <MoneyDisplay amount={totalDue} /> {t('bills.total')}
          </>
        }
        actions={
          <>
            {selected.size > 0 && (
              <button
                type="button"
                style={primaryBtn}
                onClick={() => setShowPayConfirm(true)}
                disabled={paying}
              >
                <CheckCircle size={13} /> {t('bills.paySelected', { count: selected.size })}
              </button>
            )}
            <button
              type="button"
              style={ghostBtn}
              aria-label="New Bill"
              onClick={() => setShowCreate(true)}
            >
              <Plus size={13} /> {t('bills.new')}
            </button>
          </>
        }
      />

      <NewBillModal open={showCreate} onClose={() => setShowCreate(false)} />

      <ConfirmDialog
        open={showPayConfirm}
        onClose={() => setShowPayConfirm(false)}
        onConfirm={() => void payAll()}
        title={t('bills.payConfirmTitle', { count: selected.size })}
        message={t('bills.payConfirmMessage', { count: selected.size })}
        confirmLabel={paying ? t('bills.paying') : t('bills.confirmPaid')}
        pending={paying}
      />

      {isLoading ? (
        <div style={{ ...cardStyle, padding: 20 }}>
          <SkeletonRow />
        </div>
      ) : data.length === 0 ? (
        <EmptyState
          icon={Receipt}
          title={t('bills.empty')}
          description={t('bills.emptyDescription')}
          action={{
            label: t('bills.new'),
            onClick: () => setShowCreate(true),
          }}
        />
      ) : (
        <div style={{ ...cardStyle, overflow: 'hidden' }}>
          <table style={{ width: '100%', borderCollapse: 'collapse' }}>
            <thead>
              <tr style={{ borderBottom: '1px solid var(--hub-border)' }}>
                {[
                  '',
                  t('bills.number'),
                  t('bills.vendor'),
                  t('bills.category'),
                  t('bills.amount'),
                  t('bills.due'),
                  t('bills.status'),
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
              {data.map((b) => (
                <tr key={b.id} style={{ borderBottom: '1px solid var(--hub-border)' }}>
                  <td style={{ padding: '12px 16px', width: 32 }}>
                    <input
                      type="checkbox"
                      checked={selected.has(b.id)}
                      onChange={() => toggle(b.id)}
                      disabled={b.status === 'paid'}
                      aria-label={t('bills.selectBill', { id: b.id })}
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
                        b.status === 'overdue' ? 'var(--finance-danger)' : 'var(--hub-text-muted)',
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
                    {b.status === 'unpaid' || b.status === 'overdue' ? (
                      <button
                        type="button"
                        style={{
                          ...ghostBtn,
                          padding: '4px 10px',
                          opacity: 0.6,
                          cursor: 'not-allowed',
                        }}
                        disabled
                        title={t('bills.schedulingSoon')}
                      >
                        <Clock size={12} /> {t('bills.schedule')}
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
      )}
    </motion.div>
  );
};

export default Bills;
