import React, { useState } from 'react';
import { CheckCircle } from 'lucide-react';
import { toast } from 'sonner';
import Modal from './Modal';
import CurrencyInput from './CurrencyInput';
import { inputStyle, primaryBtn, ghostBtn, labelStyle } from '../lib/tokens';

interface RecordPaymentModalProps {
  open: boolean;
  onClose: () => void;
  invoiceId: string;
  amount: number;
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
  onRecorded,
}) => {
  const [value, setValue] = useState(amount);
  const [method, setMethod] = useState(METHODS[0]);
  const [date, setDate] = useState(new Date().toISOString().slice(0, 10));
  const [reference, setReference] = useState('');

  const handle = () => {
    if (value <= 0) {
      toast.error('Amount must be greater than zero');
      return;
    }
    onRecorded?.(value);
    toast.success(
      `Recorded ${value.toLocaleString('en-US', { style: 'currency', currency: 'USD' })} payment`
    );
    onClose();
  };

  return (
    <Modal
      open={open}
      onClose={onClose}
      title={`Record payment for ${invoiceId}`}
      width={420}
      footer={
        <>
          <button style={ghostBtn} onClick={onClose}>
            Cancel
          </button>
          <button style={primaryBtn} onClick={handle}>
            <CheckCircle size={13} />
            Record Payment
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
            placeholder="Optional txn ID"
            style={inputStyle}
          />
        </div>
      </div>
    </Modal>
  );
};

export default RecordPaymentModal;
