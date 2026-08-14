import React, { useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { motion } from 'motion/react';
import { Plus, CheckCircle, Clock } from 'lucide-react';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import PageHeader from '../../shared/components/PageHeader';
import MoneyDisplay from '../../shared/components/MoneyDisplay';
import EmptyState from '../../shared/components/EmptyState';
import { SkeletonRow } from '../../shared/components/Skeleton';
import Modal from '../../shared/components/Modal';
import CurrencyInput from '../../shared/components/CurrencyInput';
import { useBills } from './useBills';
import { useCreateBill } from '../../shared/hooks/queries/useBills';
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
  amount: number;
  dueDate: string;
  description: string;
}

const NewBillModal: React.FC<{ open: boolean; onClose: () => void }> = ({ open, onClose }) => {
  const { t } = useTranslation();
  const createBill = useCreateBill();
  const [form, setForm] = useState<NewBillForm>({
    vendor: '',
    amount: 0,
    dueDate: new Date().toISOString().slice(0, 10),
    description: '',
  });

  const handleSubmit = (): void => {
    if (!form.vendor.trim()) {
      toast.error(t('bills.vendorRequired'));
      return;
    }
    if (form.amount <= 0) {
      toast.error(t('bills.amountPositive'));
      return;
    }
    createBill.mutate(
      {
        supplierId: '0',
        amount: form.amount,
        currency: 'USD',
        dueDate: form.dueDate,
        status: 'unpaid',
        description: form.description || form.vendor,
      },
      {
        onSuccess: () => {
          onClose();
          setForm({
            vendor: '',
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
          <button style={ghostBtn} onClick={onClose} disabled={createBill.isPending}>
            {t('common.cancel')}
          </button>
          <button
            style={{ ...primaryBtn, opacity: createBill.isPending ? 0.6 : 1 }}
            onClick={handleSubmit}
            disabled={createBill.isPending}
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

  const payAll = (): void => {
    toast.info(t('bills.bulkPaymentSoon'));
    setSelected(new Set());
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
              <button style={primaryBtn} onClick={payAll}>
                <CheckCircle size={13} /> Pay selected ({selected.size})
              </button>
            )}
            <button style={ghostBtn} aria-label="New Bill" onClick={() => setShowCreate(true)}>
              <Plus size={13} /> {t('bills.new')}
            </button>
          </>
        }
      />

      <NewBillModal open={showCreate} onClose={() => setShowCreate(false)} />

      {isLoading ? (
        <div style={{ ...cardStyle, padding: 20 }}>
          <SkeletonRow />
        </div>
      ) : data.length === 0 ? (
        <EmptyState title={t('bills.empty')} description={t('bills.emptyDescription')} />
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
                        style={{ ...ghostBtn, padding: '4px 10px' }}
                        onClick={() => toast.info(t('bills.schedulingSoon'))}
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
