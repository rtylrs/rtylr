import React, { useState } from 'react';
import { CheckCircle } from 'lucide-react';
import { toast } from 'sonner';
import { getOneKitErrorMessage } from '@voxire/api';
import Modal from './Modal';
import CurrencyInput from './CurrencyInput';
import { inputStyle, primaryBtn, ghostBtn, labelStyle } from '../lib/tokens';
import { useRecordPayment } from '../../features/invoices/useInvoices';
import { formatMoney } from '../lib/money';
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
  const [value, setValue] = useState(amount);
  const [method, setMethod] = useState(METHODS[0]);
  const [date, setDate] = useState(todayIso());
  const [reference, setReference] = useState('');
  const recordPayment = useRecordPayment();

  const handle = (): void => {
    if (value <= 0) {
      toast.error('Amount must be greater than zero');
      return;
    }
    if (!date || !reference.trim()) {
      toast.error('Payment date and reference are required');
      return;
    }
    recordPayment.mutate(
      {
        id: invoiceId,
        amount: value,
        currency: currency || 'USD',
        paymentDate: date,
        method,
        reference: reference.trim(),
      },
      {
        onSuccess: () => {
          onRecorded?.(value);
          toast.success(`Recorded ${formatMoney(value)} payment`);
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
            Cancel
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
          <label style={labelStyle}>Amount</label>
          <CurrencyInput value={value} onChange={setValue} />
        </div>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
          <label style={labelStyle}>Method</label>
          <select value={method} onChange={(e) => setMethod(e.target.value)} style={inputStyle}>
            {METHODS.map((m) => (
              <option key={m}>{m}</option>
            ))}
          </select>
        </div>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
          <label style={labelStyle}>Date</label>
          <input
            type="date"
            value={date}
            onChange={(e) => setDate(e.target.value)}
            style={inputStyle}
          />
        </div>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
          <label style={labelStyle}>Reference</label>
          <input
            value={reference}
            onChange={(e) => setReference(e.target.value)}
            placeholder="Transaction ID or receipt number"
            style={inputStyle}
          />
        </div>
      </div>
    </Modal>
  );
};

export default RecordPaymentModal;
