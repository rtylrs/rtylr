import React, { useEffect, useState } from 'react';
import { CheckCircle } from 'lucide-react';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { getOneKitErrorMessage } from '@voxire/api';
import Modal from './Modal';
import CurrencyInput from './CurrencyInput';
import { inputStyle, primaryBtn, ghostBtn, labelStyle } from '../lib/tokens';
import { useRecordPayment } from '../../features/invoices/useInvoices';
import { formatMoney } from '../lib/money';

interface RecordPaymentModalProps {
  open: boolean;
  onClose: () => void;
  invoiceId: string;
  amount: number;
  currency?: string | null;
  onRecorded?: (amount: number) => void;
}

// App-canonical lowercase vocabulary matching the hook default ('cash') and
// the backend's payment method strings.
const METHODS = ['cash', 'card', 'bank_transfer'] as const;

const RecordPaymentModal: React.FC<RecordPaymentModalProps> = ({
  open,
  onClose,
  invoiceId,
  amount,
  currency,
  onRecorded,
}) => {
  const { t } = useTranslation();
  const [value, setValue] = useState(amount);
  const [method, setMethod] = useState<string>(METHODS[0]);
  const [date, setDate] = useState(new Date().toISOString().slice(0, 10));
  const [reference, setReference] = useState('');
  const recordPayment = useRecordPayment();

  useEffect(() => {
    if (open) {
      setValue(amount);
      setMethod(METHODS[0]);
      setDate(new Date().toISOString().slice(0, 10));
      setReference('');
    }
  }, [open, amount]);

  // Payments must be recorded in the invoice's own currency; when it is
  // unknown the submit is disabled instead of silently defaulting to USD.
  const invoiceCurrency = typeof currency === 'string' && currency ? currency : null;
  const exceedsBalance = value > amount + 1e-9;
  const canSubmit =
    value > 0 &&
    !exceedsBalance &&
    invoiceCurrency !== null &&
    !!date &&
    !!reference.trim() &&
    !recordPayment.isPending;

  const handle = (): void => {
    if (invoiceCurrency === null) {
      toast.error(t('payments.currencyUnknown'));
      return;
    }
    if (value <= 0) {
      toast.error(t('payments.amountPositive'));
      return;
    }
    if (exceedsBalance) {
      toast.error(t('payments.exceedsBalance'));
      return;
    }
    if (!date || !reference.trim()) {
      toast.error(t('payments.dateReferenceRequired'));
      return;
    }
    recordPayment.mutate(
      {
        id: invoiceId,
        amount: value,
        currency: invoiceCurrency,
        paymentDate: date,
        method,
        reference: reference.trim(),
      },
      {
        onSuccess: () => {
          onRecorded?.(value);
          toast.success(t('payments.recorded', { amount: formatMoney(value, invoiceCurrency) }));
          onClose();
        },
        onError: (error) => toast.error(getOneKitErrorMessage(error, t('payments.recordFailed'))),
      }
    );
  };

  return (
    <Modal
      open={open}
      onClose={onClose}
      blockClose={recordPayment.isPending}
      title={t('payments.title', { id: invoiceId })}
      width={420}
      footer={
        <>
          <button
            type="button"
            style={ghostBtn}
            onClick={onClose}
            disabled={recordPayment.isPending}
          >
            {t('common.cancel')}
          </button>
          <button
            type="button"
            style={{ ...primaryBtn, opacity: canSubmit ? 1 : 0.6 }}
            onClick={handle}
            disabled={!canSubmit}
          >
            <CheckCircle size={13} />
            {recordPayment.isPending ? t('payments.recording') : t('payments.record')}
          </button>
        </>
      }
    >
      <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
        {invoiceCurrency === null ? (
          <span role="alert" style={{ fontSize: 12, color: 'var(--finance-danger)' }}>
            {t('payments.currencyUnknownHint')}
          </span>
        ) : (
          <span style={{ fontSize: 12, color: 'var(--hub-text-muted)' }}>
            {t('payments.balanceDue', { amount: formatMoney(amount, invoiceCurrency) })}
          </span>
        )}
        <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
          <label style={labelStyle} htmlFor="record-payment-amount">
            {t('payments.amount')}
          </label>
          <CurrencyInput value={value} onChange={setValue} />
          {exceedsBalance && (
            <span role="alert" style={{ fontSize: 11, color: 'var(--finance-danger)' }}>
              {t('payments.exceedsBalanceHint', {
                amount: formatMoney(amount, invoiceCurrency ?? 'USD'),
              })}
            </span>
          )}
        </div>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
          <label style={labelStyle}>{t('payments.method')}</label>
          <select
            value={method}
            onChange={(e) => setMethod(e.target.value)}
            aria-label={t('payments.method')}
            style={inputStyle}
          >
            {METHODS.map((m) => (
              <option key={m} value={m}>
                {t(`payments.methods.${m}`)}
              </option>
            ))}
          </select>
        </div>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
          <label style={labelStyle} htmlFor="record-payment-date">
            {t('payments.date')}
          </label>
          <input
            id="record-payment-date"
            type="date"
            value={date}
            onChange={(e) => setDate(e.target.value)}
            style={inputStyle}
          />
        </div>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
          <label style={labelStyle} htmlFor="record-payment-reference">
            {t('payments.reference')}
          </label>
          <input
            id="record-payment-reference"
            value={reference}
            onChange={(e) => setReference(e.target.value)}
            placeholder={t('payments.referencePlaceholder')}
            style={inputStyle}
          />
        </div>
      </div>
    </Modal>
  );
};

export default RecordPaymentModal;
