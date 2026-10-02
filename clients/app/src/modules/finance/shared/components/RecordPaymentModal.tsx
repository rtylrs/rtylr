import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { CheckCircle } from 'lucide-react';
import { toast } from 'sonner';
import { getOneKitErrorMessage } from '@voxire/api';
import Modal from './Modal';
import CurrencyInput from './CurrencyInput';
import { inputStyle, primaryBtn, ghostBtn, labelStyle } from '../lib/tokens';
import { useRecordPayment } from '../../features/invoices/useInvoices';
import { formatMoney, majorToMinor, minorToMajor, type Currency } from '../lib/money';
import { todayIso } from '../lib/date';

interface RecordPaymentModalProps {
  open: boolean;
  onClose: () => void;
  invoiceId: string;
  amount: number;
  currency?: string | null;
  onRecorded?: (amount: number) => void;
}

const METHODS = [
  'Bank Transfer',
  'Credit Card',
  'Cash',
  'Check',
  'ACH',
  'Wire',
  'PayPal',
  'Stripe',
];

const RecordPaymentModal: React.FC<RecordPaymentModalProps> = ({
  open,
  onClose,
  invoiceId,
  amount,
  currency,
  onRecorded,
}) => {
  const { t } = useTranslation('financeCommon');
  const resolvedCurrency = (currency || 'USD') as Currency;
  const [value, setValue] = useState(() => majorToMinor(amount, resolvedCurrency));
  const [method, setMethod] = useState(METHODS[0]);
  const [date, setDate] = useState(todayIso());
  const [reference, setReference] = useState('');
  const recordPayment = useRecordPayment();

  React.useEffect(() => {
    if (open) setValue(majorToMinor(amount, resolvedCurrency));
  }, [amount, open, resolvedCurrency]);

  const handle = (): void => {
    if (value <= 0) {
      toast.error(t('financeBankAccounts.payment.amountPositive'));
      return;
    }
    if (value > majorToMinor(amount, resolvedCurrency)) {
      toast.error(t('financeBankAccounts.payment.amountExceedsDue'));
      return;
    }
    if (!date || !reference.trim()) {
      toast.error(t('financeBankAccounts.payment.dateAndReferenceRequired'));
      return;
    }
    const majorAmount = minorToMajor(value, resolvedCurrency);
    recordPayment.mutate(
      {
        id: invoiceId,
        amount: majorAmount,
        currency: resolvedCurrency,
        paymentDate: date,
        method,
        reference: reference.trim(),
      },
      {
        onSuccess: () => {
          onRecorded?.(majorAmount);
          toast.success(`Recorded ${formatMoney(majorAmount, resolvedCurrency)} payment`);
          onClose();
        },
        onError: (error) => toast.error(getOneKitErrorMessage(error, 'Failed to record payment')),
      }
    );
  };

  return (
    <Modal
      open={open}
      onClose={onClose}
      title={`Record payment for ${invoiceId}`}
      width={420}
      footer={
        <>
          <button
            type="button"
            style={ghostBtn}
            onClick={onClose}
            disabled={recordPayment.isPending}
          >
            {t('financeCommon.cancel')}
          </button>
          <button
            type="button"
            style={{ ...primaryBtn, opacity: recordPayment.isPending ? 0.6 : 1 }}
            onClick={handle}
            disabled={recordPayment.isPending}
          >
            <CheckCircle size={13} />
            {recordPayment.isPending ? 'Recording...' : 'Record Payment'}
          </button>
        </>
      }
    >
      <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
          <label style={labelStyle}>{t('financeCommon.amount')}</label>
          <CurrencyInput value={value} currency={resolvedCurrency} onChange={setValue} />
        </div>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
          <label style={labelStyle}>{t('financeBankAccounts.payment.method')}</label>
          <select value={method} onChange={(e) => setMethod(e.target.value)} style={inputStyle}>
            {METHODS.map((m) => (
              <option key={m}>{m}</option>
            ))}
          </select>
        </div>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
          <label style={labelStyle}>{t('financeCommon.date')}</label>
          <input
            type="date"
            value={date}
            onChange={(e) => setDate(e.target.value)}
            style={inputStyle}
          />
        </div>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
          <label style={labelStyle}>{t('financeBankAccounts.payment.reference')}</label>
          <input
            value={reference}
            onChange={(e) => setReference(e.target.value)}
            placeholder={t('financeBankAccounts.payment.referencePlaceholder')}
            style={inputStyle}
          />
        </div>
      </div>
    </Modal>
  );
};

export default RecordPaymentModal;
