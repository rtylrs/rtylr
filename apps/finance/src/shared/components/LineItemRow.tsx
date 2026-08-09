import React from 'react';
import { Trash2 } from 'lucide-react';
import CurrencyInput from './CurrencyInput';
import MoneyDisplay from './MoneyDisplay';
import { inputStyle } from '../lib/tokens';

export interface LineItem {
  id: string;
  description: string;
  quantity: number;
  unitPrice: number;
  taxRate: number;
}

interface LineItemRowProps {
  item: LineItem;
  onChange: (item: LineItem) => void;
  onRemove: () => void;
  disabled?: boolean;
}

const LineItemRow: React.FC<LineItemRowProps> = ({ item, onChange, onRemove, disabled }) => {
  const subtotal = item.quantity * item.unitPrice;
  const tax = subtotal * (item.taxRate / 100);
  const total = subtotal + tax;

  return (
    <tr style={{ borderBottom: '1px solid var(--hub-border)' }}>
      <td style={{ padding: 8, width: '40%' }}>
        <input
          value={item.description}
          onChange={(e) => onChange({ ...item, description: e.target.value })}
          placeholder="Description"
          disabled={disabled}
          style={inputStyle}
        />
      </td>
      <td style={{ padding: 8, width: 90 }}>
        <input
          type="number"
          value={item.quantity}
          onChange={(e) => {
            const v = parseFloat(e.target.value);
            if (!isNaN(v) && v > 0) onChange({ ...item, quantity: v });
          }}
          disabled={disabled}
          style={{ ...inputStyle, textAlign: 'right' }}
        />
      </td>
      <td style={{ padding: 8, width: 140 }}>
        <CurrencyInput
          value={item.unitPrice}
          onChange={(v) => onChange({ ...item, unitPrice: v })}
          disabled={disabled}
        />
      </td>
      <td style={{ padding: 8, width: 90 }}>
        <input
          type="number"
          step="0.1"
          value={item.taxRate}
          onChange={(e) => {
            const v = parseFloat(e.target.value);
            if (!isNaN(v) && v >= 0) onChange({ ...item, taxRate: v });
          }}
          disabled={disabled}
          style={{ ...inputStyle, textAlign: 'right' }}
        />
      </td>
      <td
        style={{
          padding: 8,
          width: 120,
          textAlign: 'right',
          fontSize: 13,
          color: 'var(--hub-text)',
          fontWeight: 600,
        }}
      >
        <MoneyDisplay amount={total} />
      </td>
      <td style={{ padding: 8, width: 32 }}>
        <button
          onClick={onRemove}
          disabled={disabled}
          style={{
            background: 'none',
            border: 'none',
            cursor: 'pointer',
            color: '#ef4444',
            padding: 4,
          }}
          title="Remove line"
        >
          <Trash2 size={14} />
        </button>
      </td>
    </tr>
  );
};

export default LineItemRow;
